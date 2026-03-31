package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
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
		if a.isDirectory() {
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

	// Step 3: Resolve regions, check existence/TP, select stacks
	selector := NewCdkStackSelector(a.stackNames, a.interactiveMode, a.forceMode)
	resolver := NewCdkStackResolver(selector, a.profile, a.region, a.forceMode)
	targetStacks, err := resolver.Resolve(ctx, stacks)
	if err != nil {
		return err
	}
	if len(targetStacks) == 0 {
		return nil
	}

	// Step 4: Handle TerminationProtection stacks
	tpStacks := filterTPStacks(targetStacks)
	if len(tpStacks) > 0 {
		if !a.forceMode {
			return fmt.Errorf("TerminationProtectionError: %s", joinStackNames(tpStacks))
		}
		if !a.showTPConfirmation(tpStacks) {
			io.Logger.Info().Msg("Canceled.")
			return nil
		}
	}

	// Step 5: Show confirmation
	if !a.showCdkConfirmation(targetStacks) {
		io.Logger.Info().Msg("Canceled.")
		return nil
	}

	// Step 6: Delete stacks
	return NewCdkDeleter(a.profile, a.forceMode, a.concurrencyNumber).DeleteStacks(ctx, targetStacks)
}

func filterTPStacks(stacks []cdk.StackInfo) []cdk.StackInfo {
	var tp []cdk.StackInfo
	for _, s := range stacks {
		if s.TerminationProtection {
			tp = append(tp, s)
		}
	}
	return tp
}

func joinStackNames(stacks []cdk.StackInfo) string {
	names := make([]string, len(stacks))
	for i, s := range stacks {
		names[i] = s.StackName
	}
	return strings.Join(names, ", ")
}

func (a *CdkAction) showTPConfirmation(tpStacks []cdk.StackInfo) bool {
	fmt.Fprintf(os.Stderr, "The following stacks have TerminationProtection enabled:\n")
	for _, s := range tpStacks {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.StackName, s.Region)
	}
	fmt.Fprintf(os.Stderr, "\nTerminationProtection will be disabled before deletion.\n")
	return io.GetYesNo("Do you want to proceed?")
}

func (a *CdkAction) showCdkConfirmation(stacks []cdk.StackInfo) bool {
	fmt.Fprintf(os.Stderr, "The following stacks will be deleted:\n")
	for _, s := range stacks {
		fmt.Fprintf(os.Stderr, "  - %s (%s)\n", s.StackName, s.Region)
	}
	fmt.Fprintln(os.Stderr)

	return io.GetYesNo("Are you sure you want to delete these stacks?")
}

func (a *CdkAction) isDirectory() bool {
	info, err := os.Stat(a.appPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
