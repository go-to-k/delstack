package app

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/urfave/cli/v2"
)

type App struct {
	Cli             *cli.App
	StackNames      *cli.StringSlice
	Profile         string
	Region          string
	InteractiveMode bool
}

type targetStack struct {
	stackName           string
	targetResourceTypes []string
}

func NewApp(version string) *App {
	app := App{}
	app.StackNames = cli.NewStringSlice()

	app.Cli = &cli.App{
		Name:  "delstack",
		Usage: "A CLI tool to force delete the entire CloudFormation stack.",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "stackName",
				Aliases:     []string{"s"},
				Usage:       "CloudFormation stack names(one or more)",
				Destination: app.StackNames,
			},
			&cli.StringFlag{
				Name:        "profile",
				Aliases:     []string{"p"},
				Usage:       "AWS profile name",
				Destination: &app.Profile,
			},
			&cli.StringFlag{
				Name:        "region",
				Aliases:     []string{"r"},
				Usage:       "AWS region",
				Destination: &app.Region,
			},
			&cli.BoolFlag{
				Name:        "interactive",
				Aliases:     []string{"i"},
				Value:       false,
				Usage:       "Interactive Mode",
				Destination: &app.InteractiveMode,
			},
		},
	}

	app.Cli.Version = version
	app.Cli.Action = app.getAction()
	app.Cli.HideHelpCommand = true

	return &app
}

func (a *App) Run(ctx context.Context) error {
	return a.Cli.RunContext(ctx, os.Args)
}

func (a *App) getAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if !a.InteractiveMode && len(a.StackNames.Value()) == 0 {
			errMsg := fmt.Sprintln("At least one stack name must be specified in command options (-s) or a flow of the interactive mode (-i).")
			return fmt.Errorf("InvalidOptionError: %v", errMsg)
		}

		config, err := client.LoadAWSConfig(c.Context, a.Region, a.Profile)
		if err != nil {
			return err
		}

		operatorFactory := operation.NewOperatorFactory(config)
		cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(resourcetype.GetResourceTypes())

		deduplicatedStackNames := a.deduplicateStackNames()

		sortedStackNames, err := a.getSortedStackNames(c.Context, cloudformationStackOperator, deduplicatedStackNames)
		if err != nil {
			return err
		}
		// The case for interruption(Ctrl + C)
		if len(sortedStackNames) == 0 {
			return nil
		}

		targetStacks := a.attachTargetResourceTypes(sortedStackNames, deduplicatedStackNames)
		// Explanation of deletion order in the case of multiple stacks
		if len(targetStacks) > 1 {
			io.Logger.Info().Msg("The stacks are removed in order of the latest creation time, taking into account dependencies.")
		}

		isRootStack := true
		for _, stack := range targetStacks {
			operatorCollection := operation.NewOperatorCollection(config, operatorFactory, stack.targetResourceTypes)
			operatorManager := operation.NewOperatorManager(operatorCollection)
			cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(stack.targetResourceTypes)

			io.Logger.Info().Msgf("%v: Start deletion. Please wait a few minutes...", stack.stackName)

			if err := cloudformationStackOperator.DeleteCloudFormationStack(c.Context, aws.String(stack.stackName), isRootStack, operatorManager); err != nil {
				return err
			}

			io.Logger.Info().Msgf("%v: Successfully deleted!!", stack.stackName)
		}
		return nil
	}
}

func (a *App) deduplicateStackNames() []string {
	deduplicatedStackNames := []string{}

	for _, stackName := range a.StackNames.Value() {
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

func (a *App) getSortedStackNames(ctx context.Context, cloudformationStackOperator *operation.CloudFormationStackOperator, specifiedStackNames []string) ([]string, error) {
	sortedStackNames := []string{}

	if len(specifiedStackNames) != 0 {
		stackNames, err := cloudformationStackOperator.GetSortedStackNames(ctx, specifiedStackNames)
		if err != nil {
			return sortedStackNames, err
		}
		sortedStackNames = append(sortedStackNames, stackNames...)
	} else if a.InteractiveMode {
		keyword := a.inputKeywordForFilter()
		stacks, err := cloudformationStackOperator.ListStacksFilteredByKeyword(ctx, aws.String(keyword))
		if err != nil {
			return sortedStackNames, err
		}

		// The `ListStacksFilteredByKeyword` with SDK's `DescribeStacks` returns the stacks in descending order of CreationTime.
		// Therefore, by deleting stacks in the same order, we can delete from a new stack that is not depended on by any stack.
		stackNames := a.selectStackNames(stacks)
		sortedStackNames = append(sortedStackNames, stackNames...)
	}

	return sortedStackNames, nil
}

func (a *App) attachTargetResourceTypes(sortedStackNames []string, specifiedStackNames []string) []targetStack {
	targetStacks := []targetStack{}

	// If stackNames are specified with an interactive mode option, select ResourceTypes in the order specified (not sorted order).
	if a.InteractiveMode && len(specifiedStackNames) != 0 {
		var selectedResourceTypes []targetStack
		for _, stackName := range specifiedStackNames {
			targetResourceTypes, continuation := a.selectResourceTypes(stackName)
			if !continuation {
				return nil
			}
			selectedResourceTypes = append(selectedResourceTypes, targetStack{
				stackName:           stackName,
				targetResourceTypes: targetResourceTypes,
			})
		}
		for _, stackName := range sortedStackNames {
			for _, selectedResourceType := range selectedResourceTypes {
				if stackName == selectedResourceType.stackName {
					targetStacks = append(targetStacks, selectedResourceType)
				}
			}
		}
	}
	if a.InteractiveMode && len(specifiedStackNames) == 0 {
		for _, stackName := range sortedStackNames {
			targetResourceTypes, continuation := a.selectResourceTypes(stackName)
			if !continuation {
				return nil
			}
			targetStacks = append(targetStacks, targetStack{
				stackName:           stackName,
				targetResourceTypes: targetResourceTypes,
			})
		}
	}
	if !a.InteractiveMode {
		for _, stackName := range sortedStackNames {
			targetResourceTypes := resourcetype.GetResourceTypes()
			targetStacks = append(targetStacks, targetStack{
				stackName:           stackName,
				targetResourceTypes: targetResourceTypes,
			})
		}
	}

	return targetStacks
}

func (a *App) inputKeywordForFilter() string {
	label := "Filter a keyword of stack names(case-insensitive): "
	return io.InputKeywordForFilter(label)
}

func (a *App) selectResourceTypes(stackName string) ([]string, bool) {
	var checkboxes []string

	label := stackName +
		"\n" +
		"Select ResourceTypes you wish to delete even if DELETE_FAILED.\n" +
		"However, if a resource can be deleted without becoming DELETE_FAILED by the normal CloudFormation stack deletion feature, the resource will be deleted even if you do not select that resource type. " +
		"\n"

	opts := resourcetype.GetResourceTypes()

	for {
		checkboxes = io.GetCheckboxes(label, opts)

		if len(checkboxes) == 0 {
			ok := io.GetYesNo("No selection?")
			if ok {
				return checkboxes, true
			}

			// The case for interruption(Ctrl + C)
			ok = io.GetYesNo("Do you want to finish?")
			if ok {
				io.Logger.Info().Msg("Finished...")
				return checkboxes, false
			}
			continue
		}

		ok := io.GetYesNo("OK?")
		if ok {
			return checkboxes, true
		}
	}
}

func (a *App) selectStackNames(stackNames []string) []string {
	var selectedStackNames []string

	label := "Select StackNames." + "\n" +
		"Nested child stacks, XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks and EnableTerminationProtection stacks are not displayed." +
		"\n"

	for {
		selectedStackNames = io.GetCheckboxes(label, stackNames)

		if len(selectedStackNames) == 0 {
			// The case for interruption(Ctrl + C)
			ok := io.GetYesNo("Do you want to finish?")
			if ok {
				io.Logger.Info().Msg("Finished...")
				return selectedStackNames
			}
			continue
		}

		ok := io.GetYesNo("OK?")
		if ok {
			return selectedStackNames
		}
	}
}
