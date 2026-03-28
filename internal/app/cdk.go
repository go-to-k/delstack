package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/urfave/cli/v2"
)

func (a *App) getCdkAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if a.ConcurrencyNumber < UnspecifiedConcurrencyNumber {
			return fmt.Errorf("InvalidOptionError: You must specify a positive number for the -n option.")
		}

		io.AutoYes = a.YesMode

		// Step 1: Synthesize or read existing cdk.out
		cdkOutDir := cdk.DefaultCdkOutDir
		if a.CdkAppPath != "" {
			cdkOutDir = a.CdkAppPath
		} else {
			synthesizer := cdk.NewSynthesizer()
			if err := synthesizer.Synth(c.Context, a.CdkContexts.Value()); err != nil {
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
		defaultRegion, err := a.resolveDefaultRegion(c.Context)
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

		// Step 5: Show confirmation
		if !a.showCdkConfirmation(selectedStacks) {
			io.Logger.Info().Msg("Canceled.")
			return nil
		}

		// Step 6: Group by region and delete
		return a.deleteCdkStacks(c.Context, selectedStacks)
	}
}

func (a *App) resolveDefaultRegion(ctx context.Context) (string, error) {
	if a.Region != "" {
		return a.Region, nil
	}
	cfg, err := client.LoadAWSConfig(ctx, "", a.Profile)
	if err != nil {
		return "", fmt.Errorf("failed to resolve default region: %w", err)
	}
	return cfg.Region, nil
}

func (a *App) selectCdkStacks(stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	specifiedNames := a.StackNames.Value()

	if len(specifiedNames) > 0 {
		// Filter by -s flag
		nameSet := make(map[string]struct{})
		for _, name := range specifiedNames {
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

	if a.InteractiveMode {
		return a.selectCdkStacksInteractively(stacks)
	}

	// Default: all stacks
	return stacks, nil
}

func (a *App) selectCdkStacksInteractively(stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	// Build display names with region info
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

	// Map selected display names back to StackInfo
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

func (a *App) showCdkConfirmation(stacks []cdk.StackInfo) bool {
	fmt.Fprintf(os.Stderr, "The following stacks will be deleted:\n")
	for _, s := range stacks {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.StackName, s.Region)
	}
	fmt.Fprintln(os.Stderr)

	return io.GetYesNo("Are you sure you want to delete these stacks?")
}

func (a *App) deleteCdkStacks(ctx context.Context, stacks []cdk.StackInfo) error {
	// Group stacks by region
	regionStacks := make(map[string][]cdk.StackInfo)
	for _, s := range stacks {
		regionStacks[s.Region] = append(regionStacks[s.Region], s)
	}

	// Sort regions for deterministic processing
	var regions []string
	for region := range regionStacks {
		regions = append(regions, region)
	}
	sort.Strings(regions)

	// Build cross-region dependency graph
	// stackName -> StackInfo for quick lookup
	stackMap := make(map[string]cdk.StackInfo)
	for _, s := range stacks {
		stackMap[s.StackName] = s
	}

	// Check if there are cross-region dependencies
	hasCrossRegionDeps := false
	for _, s := range stacks {
		for _, dep := range s.Dependencies {
			if depStack, ok := stackMap[dep]; ok {
				if depStack.Region != s.Region {
					hasCrossRegionDeps = true
					break
				}
			}
		}
		if hasCrossRegionDeps {
			break
		}
	}

	if !hasCrossRegionDeps {
		// Simple case: delete each region independently
		for _, region := range regions {
			regionStackInfos := regionStacks[region]
			if err := a.deleteStacksInRegion(ctx, region, regionStackInfos); err != nil {
				return err
			}
		}
		return nil
	}

	// Complex case: cross-region dependencies
	// Use topological ordering across all stacks
	return a.deleteStacksWithCrossRegionDeps(ctx, stacks, stackMap)
}

func (a *App) deleteStacksInRegion(ctx context.Context, region string, stacks []cdk.StackInfo) error {
	io.Logger.Info().Msgf("Deleting %d stack(s) in %s...", len(stacks), region)

	config, err := client.LoadAWSConfig(ctx, region, a.Profile)
	if err != nil {
		return fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
	}

	operatorFactory := operation.NewOperatorFactory(config, a.ForceMode)
	deleter := NewStackDeleter(a.ForceMode, a.ConcurrencyNumber)

	stackNames := make([]string, len(stacks))
	for i, s := range stacks {
		stackNames[i] = s.StackName
	}

	return deleter.DeleteStacksConcurrently(ctx, stackNames, config, operatorFactory)
}

func (a *App) deleteStacksWithCrossRegionDeps(ctx context.Context, stacks []cdk.StackInfo, stackMap map[string]cdk.StackInfo) error {
	io.Logger.Info().Msg("Cross-region dependencies detected. Deleting stacks in dependency order...")

	// Build reverse in-degree (how many stacks depend on each stack)
	reverseInDegree := make(map[string]int)
	dependents := make(map[string][]string) // stack -> stacks that depend on it

	for _, s := range stacks {
		if _, ok := reverseInDegree[s.StackName]; !ok {
			reverseInDegree[s.StackName] = 0
		}
		for _, dep := range s.Dependencies {
			if _, ok := stackMap[dep]; ok {
				reverseInDegree[dep]++
				dependents[s.StackName] = append(dependents[s.StackName], dep)
			}
		}
	}

	// Process stacks in topological order (delete dependents first)
	deleted := make(map[string]bool)
	configCache := make(map[string]aws.Config)

	for len(deleted) < len(stacks) {
		// Find stacks with reverse in-degree 0
		var ready []cdk.StackInfo
		for _, s := range stacks {
			if !deleted[s.StackName] && reverseInDegree[s.StackName] == 0 {
				ready = append(ready, s)
			}
		}

		if len(ready) == 0 {
			return fmt.Errorf("circular dependency detected among remaining stacks")
		}

		// Group ready stacks by region and delete in parallel per region
		readyByRegion := make(map[string][]cdk.StackInfo)
		for _, s := range ready {
			readyByRegion[s.Region] = append(readyByRegion[s.Region], s)
		}

		for region, regionStacks := range readyByRegion {
			config, ok := configCache[region]
			if !ok {
				var err error
				config, err = client.LoadAWSConfig(ctx, region, a.Profile)
				if err != nil {
					return fmt.Errorf("failed to load AWS config for region %s: %w", region, err)
				}
				configCache[region] = config
			}

			operatorFactory := operation.NewOperatorFactory(config, a.ForceMode)
			deleter := NewStackDeleter(a.ForceMode, a.ConcurrencyNumber)

			stackNames := make([]string, len(regionStacks))
			for i, s := range regionStacks {
				stackNames[i] = s.StackName
			}

			if err := deleter.DeleteStacksConcurrently(ctx, stackNames, config, operatorFactory); err != nil {
				return err
			}
		}

		// Mark as deleted and update in-degrees
		for _, s := range ready {
			deleted[s.StackName] = true
			for _, dep := range dependents[s.StackName] {
				reverseInDegree[dep]--
			}
		}
	}

	return nil
}
