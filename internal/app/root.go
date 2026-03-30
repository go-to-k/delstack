package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
)

type RootAction struct {
	stackNames        []string
	profile           string
	region            string
	interactiveMode   bool
	forceMode         bool
	yesMode           bool
	concurrencyNumber int
}

func NewRootAction(stackNames []string, profile, region string, interactiveMode, forceMode, yesMode bool, concurrencyNumber int) *RootAction {
	return &RootAction{
		stackNames:        stackNames,
		profile:           profile,
		region:            region,
		interactiveMode:   interactiveMode,
		forceMode:         forceMode,
		yesMode:           yesMode,
		concurrencyNumber: concurrencyNumber,
	}
}

func (a *RootAction) Run(ctx context.Context) error {
	if !a.interactiveMode && len(a.stackNames) == 0 {
		errMsg := fmt.Sprintln("At least one stack name must be specified in command options (-s) or a flow of the interactive mode (-i).")
		return fmt.Errorf("InvalidOptionError: %v", errMsg)
	}
	if a.interactiveMode && len(a.stackNames) != 0 {
		errMsg := fmt.Sprintln("Stack names (-s) cannot be specified when using Interactive Mode (-i).")
		return fmt.Errorf("InvalidOptionError: %v", errMsg)
	}
	if a.concurrencyNumber < UnspecifiedConcurrencyNumber {
		errMsg := fmt.Sprintln("You must specify a positive number for the -n option.")
		return fmt.Errorf("InvalidOptionError: %v", errMsg)
	}

	io.AutoYes = a.yesMode

	config, err := client.LoadAWSConfig(ctx, a.region, a.profile)
	if err != nil {
		return err
	}

	operatorFactory := operation.NewOperatorFactory(config, a.forceMode)
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()

	deduplicatedStackNames := a.deduplicateStackNames()

	sortedStackNames, tpStackNames, continuation, err := a.getSortedStackNames(ctx, cloudformationStackOperator, deduplicatedStackNames)
	if err != nil {
		return err
	}
	if !continuation {
		return nil
	}

	if len(tpStackNames) > 0 {
		fmt.Fprintf(os.Stderr, "The following stacks have TerminationProtection enabled:\n")
		for _, name := range tpStackNames {
			fmt.Fprintf(os.Stderr, "  - %s\n", name)
		}
		fmt.Fprintf(os.Stderr, "\nTerminationProtection will be disabled before deletion.\n")
		if !io.GetYesNo("Do you want to proceed?") {
			io.Logger.Info().Msg("Canceled.")
			return nil
		}
	}

	stackLength := len(sortedStackNames)
	if stackLength > 1 && (a.concurrencyNumber == UnspecifiedConcurrencyNumber || a.concurrencyNumber > 1) {
		var concurrency int
		if a.concurrencyNumber == UnspecifiedConcurrencyNumber {
			concurrency = stackLength
		} else {
			concurrency = min(a.concurrencyNumber, stackLength)
		}
		io.Logger.Info().Msgf("The stacks will be removed concurrently, taking into account dependencies. (concurrency: %d)", concurrency)
	}
	if err := NewStackDeleter(a.forceMode, a.concurrencyNumber, &DependencyAnalyzer{}, &StackExecutor{}).DeleteStacksConcurrently(ctx, sortedStackNames, config, operatorFactory); err != nil {
		return err
	}
	return nil
}

func (a *RootAction) deduplicateStackNames() []string {
	deduplicatedStackNames := []string{}

	for _, stackName := range a.stackNames {
		var isDuplicated bool
		for _, deduplicatedStackName := range deduplicatedStackNames {
			if stackName == deduplicatedStackName {
				isDuplicated = true
				break
			}
		}
		if !isDuplicated {
			deduplicatedStackNames = append(deduplicatedStackNames, stackName)
		}
	}

	return deduplicatedStackNames
}

func (a *RootAction) getSortedStackNames(ctx context.Context, cloudformationStackOperator *operation.CloudFormationStackOperator, specifiedStackNames []string) ([]string, []string, bool, error) {
	if len(specifiedStackNames) != 0 {
		stackNames, tpStackNames, err := cloudformationStackOperator.GetSortedStackNames(ctx, specifiedStackNames, a.forceMode)
		if err != nil {
			return nil, nil, false, err
		}
		return stackNames, tpStackNames, true, nil
	}

	if a.interactiveMode {
		keyword := a.inputKeywordForFilter()
		stacks, err := cloudformationStackOperator.ListStacksFilteredByKeyword(ctx, aws.String(keyword), a.forceMode)
		if err != nil {
			return nil, nil, false, err
		}

		// The `ListStacksFilteredByKeyword` with SDK's `DescribeStacks` returns the stacks in descending order of CreationTime.
		stackNames, continuation, err := a.selectStackNames(stacks)
		if err != nil {
			return nil, nil, false, err
		}
		if !continuation {
			return nil, nil, false, nil
		}

		// Strip TP marker and build TP stack list
		var cleanStackNames []string
		var tpStackNames []string
		for _, name := range stackNames {
			if strings.HasPrefix(name, operation.TerminationProtectionMarker) {
				cleanName := strings.TrimPrefix(name, operation.TerminationProtectionMarker)
				cleanStackNames = append(cleanStackNames, cleanName)
				tpStackNames = append(tpStackNames, cleanName)
			} else {
				cleanStackNames = append(cleanStackNames, name)
			}
		}
		return cleanStackNames, tpStackNames, true, nil
	}

	// never reach here
	return nil, nil, false, nil
}

func (a *RootAction) inputKeywordForFilter() string {
	label := "Filter a keyword of stack names(case-insensitive): "
	return io.InputKeywordForFilter(label)
}

func (a *RootAction) selectStackNames(stackNames []string) ([]string, bool, error) {
	label := []string{
		"Select StackNames.",
	}
	if a.forceMode {
		label = append(label, "Nested child stacks and XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks are not displayed.")
		label = append(label, "(* = TerminationProtection)")
	} else {
		label = append(label, "Nested child stacks, XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks and EnableTerminationProtection stacks are not displayed.")
	}

	selectedStackNames, continuation, err := io.GetCheckboxes(label, stackNames, false)
	if err != nil {
		return nil, false, err
	}
	return selectedStackNames, continuation, nil
}
