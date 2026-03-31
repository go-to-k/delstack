package app

import (
	"context"
	"fmt"
	"os"

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

	// Step 4: Confirm TerminationProtection and deletion
	confirmer := NewCdkStackConfirmer(a.forceMode)
	ok, err := confirmer.ConfirmTPStacks(targetStacks)
	if err != nil {
		return err
	}
	if !ok {
		io.Logger.Info().Msg("Canceled.")
		return nil
	}

	if !confirmer.ConfirmDeletion(targetStacks) {
		io.Logger.Info().Msg("Canceled.")
		return nil
	}

	// Step 5: Delete stacks
	return NewCdkDeleter(a.profile, a.forceMode, a.concurrencyNumber).DeleteStacks(ctx, targetStacks)
}

func (a *CdkAction) isDirectory() bool {
	info, err := os.Stat(a.appPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
