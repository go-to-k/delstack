package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
)

type CdkAction struct {
	stackNames        []string
	profile           string
	region            string
	interactiveMode   bool
	forceMode         bool
	yesMode           bool
	concurrencyNumber int
	appPath           string
	contexts          []string
}

func NewCdkAction(stackNames []string, profile, region string, interactiveMode, forceMode, yesMode bool, concurrencyNumber int, appPath string, contexts []string) *CdkAction {
	return &CdkAction{
		stackNames:        stackNames,
		profile:           profile,
		region:            region,
		interactiveMode:   interactiveMode,
		forceMode:         forceMode,
		yesMode:           yesMode,
		concurrencyNumber: concurrencyNumber,
		appPath:           appPath,
		contexts:          contexts,
	}
}

func (a *CdkAction) Run(ctx context.Context) error {
	if a.interactiveMode && len(a.stackNames) != 0 {
		return fmt.Errorf("InvalidOptionError: Stack names (-s) cannot be specified when using Interactive Mode (-i)")
	}
	if a.concurrencyNumber < UnspecifiedConcurrencyNumber {
		return fmt.Errorf("InvalidOptionError: You must specify a positive number for the -n option")
	}

	io.AutoYes = a.yesMode

	// Step 1: Synthesize or read existing cdk.out
	cdkOutDir := cdk.DefaultCdkOutDir
	if a.appPath != "" {
		if isDirectory(a.appPath) {
			// -a points to an existing cdk.out directory, skip synthesis
			cdkOutDir = a.appPath
		} else {
			// -a is an app command (e.g. "npx ts-node bin/app.ts"), run cdk synth --app
			synthesizer := cdk.NewSynthesizer()
			if err := synthesizer.Synth(ctx, a.contexts, a.appPath); err != nil {
				return err
			}
		}
	} else {
		synthesizer := cdk.NewSynthesizer()
		if err := synthesizer.Synth(ctx, a.contexts, ""); err != nil {
			return err
		}
	}

	// Step 2: Parse manifest
	stacks, err := cdk.ParseManifest(cdkOutDir)
	if err != nil {
		return err
	}
	if len(stacks) == 0 {
		io.Logger.Info().Msg("No stacks found in CDK app.")
		return nil
	}

	// Step 3: Resolve unknown regions
	defaultRegion, err := a.resolveDefaultRegion(ctx)
	if err != nil {
		return err
	}
	for i := range stacks {
		if stacks[i].Region == "unknown-region" || stacks[i].Region == "" {
			stacks[i].Region = defaultRegion
		}
	}

	// Step 4: Filter stacks by -s or -i
	selectedStacks, err := a.selectCdkStacks(stacks)
	if err != nil {
		return err
	}
	if len(selectedStacks) == 0 {
		io.Logger.Info().Msg("No stacks selected.")
		return nil
	}

	// Step 5: Filter out stacks that don't exist in AWS
	existingStacks, err := a.filterExistingStacks(ctx, selectedStacks)
	if err != nil {
		return err
	}
	if len(existingStacks) == 0 {
		io.Logger.Info().Msg("No deployed stacks found.")
		return nil
	}

	// Step 6: Show confirmation
	if !a.showCdkConfirmation(existingStacks) {
		io.Logger.Info().Msg("Canceled.")
		return nil
	}

	// Step 7: Group by region and delete
	return NewCdkDeleter(a.profile, a.forceMode, a.concurrencyNumber).DeleteStacks(ctx, existingStacks)
}

func (a *CdkAction) filterExistingStacks(ctx context.Context, stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	// Group by region to create one operator per region
	operatorCache := make(map[string]*operation.CloudFormationStackOperator)

	var existing []cdk.StackInfo
	for _, s := range stacks {
		op, ok := operatorCache[s.Region]
		if !ok {
			cfg, err := client.LoadAWSConfig(ctx, s.Region, a.profile)
			if err != nil {
				return nil, fmt.Errorf("failed to load AWS config for region %s: %w", s.Region, err)
			}
			factory := operation.NewOperatorFactory(cfg, a.forceMode)
			op = factory.CreateCloudFormationStackOperator()
			operatorCache[s.Region] = op
		}

		exists, err := op.StackExists(ctx, aws.String(s.StackName))
		if err != nil {
			return nil, err
		}
		if exists {
			existing = append(existing, s)
		} else {
			io.Logger.Info().Msgf("Stack %s not found in %s, skipping.", s.StackName, s.Region)
		}
	}

	return existing, nil
}

func (a *CdkAction) resolveDefaultRegion(ctx context.Context) (string, error) {
	if a.region != "" {
		return a.region, nil
	}
	cfg, err := client.LoadAWSConfig(ctx, "", a.profile)
	if err != nil {
		return "", fmt.Errorf("failed to resolve default region: %w", err)
	}
	return cfg.Region, nil
}

func (a *CdkAction) selectCdkStacks(stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	if len(a.stackNames) > 0 {
		nameSet := make(map[string]struct{})
		for _, name := range a.stackNames {
			nameSet[name] = struct{}{}
		}

		var selected []cdk.StackInfo
		for _, s := range stacks {
			if _, ok := nameSet[s.StackName]; ok {
				selected = append(selected, s)
				delete(nameSet, s.StackName)
			}
		}

		if len(nameSet) > 0 {
			var notFound []string
			for name := range nameSet {
				notFound = append(notFound, name)
			}
			return nil, fmt.Errorf("stacks not found in CDK app: %s", strings.Join(notFound, ", "))
		}
		return selected, nil
	}

	if a.interactiveMode {
		return a.selectCdkStacksInteractively(stacks)
	}

	return stacks, nil
}

func (a *CdkAction) selectCdkStacksInteractively(stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	displayNames := make([]string, len(stacks))
	for i, s := range stacks {
		displayNames[i] = fmt.Sprintf("%s (%s)", s.StackName, s.Region)
	}

	label := []string{"Select stacks to delete."}
	selectedNames, continuation, err := io.GetCheckboxes(label, displayNames, false)
	if err != nil {
		return nil, err
	}
	if !continuation {
		return nil, nil
	}

	selectedSet := make(map[string]struct{})
	for _, name := range selectedNames {
		selectedSet[name] = struct{}{}
	}

	var selected []cdk.StackInfo
	for i, s := range stacks {
		if _, ok := selectedSet[displayNames[i]]; ok {
			selected = append(selected, s)
		}
	}

	return selected, nil
}

func (a *CdkAction) showCdkConfirmation(stacks []cdk.StackInfo) bool {
	fmt.Fprintf(os.Stderr, "The following stacks will be deleted:\n")
	for _, s := range stacks {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.StackName, s.Region)
	}
	fmt.Fprintln(os.Stderr)

	return io.GetYesNo("Are you sure you want to delete these stacks?")
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
