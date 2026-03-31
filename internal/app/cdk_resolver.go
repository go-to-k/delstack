package app

import (
	"context"
	"fmt"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
)

type CdkStackResolver struct {
	selector         *CdkStackSelector
	region           string
	forceMode        bool
	configLoader     IConfigLoader
	existenceChecker IStackExistenceChecker
}

func NewCdkStackResolver(selector *CdkStackSelector, profile, region string, forceMode bool) *CdkStackResolver {
	return &CdkStackResolver{
		selector:         selector,
		region:           region,
		forceMode:        forceMode,
		configLoader:     &ConfigLoader{},
		existenceChecker: NewStackExistenceChecker(profile, forceMode),
	}
}

// Resolve takes raw stacks from the CDK manifest and returns the final list
// of stacks to delete: regions resolved, checked for existence and TP status,
// filtered by pattern/interactive selection.
func (r *CdkStackResolver) Resolve(ctx context.Context, stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	// Resolve unknown regions
	if err := r.resolveRegions(ctx, stacks); err != nil {
		return nil, err
	}

	if r.selector.interactiveMode {
		// Interactive mode: check existence/TP first so the selector can display TP info
		existing, err := r.filterAndAnnotate(ctx, stacks)
		if err != nil {
			return nil, err
		}
		if len(existing) == 0 {
			io.Logger.Info().Msg("No deployed stacks found.")
			return nil, nil
		}

		selected, err := r.selector.Select(existing)
		if err != nil {
			return nil, err
		}
		if len(selected) == 0 {
			io.Logger.Info().Msg("No stacks selected.")
			return nil, nil
		}
		return selected, nil
	}

	// Non-interactive mode: select first (pattern match against manifest), then check existence
	selected, err := r.selector.Select(stacks)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		io.Logger.Info().Msg("No stacks selected.")
		return nil, nil
	}

	existing, err := r.filterAndAnnotate(ctx, selected)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		io.Logger.Info().Msg("No deployed stacks found.")
		return nil, nil
	}

	return existing, nil
}

// filterAndAnnotate checks each stack's existence and TerminationProtection status.
// Returns only existing stacks with TP status annotated.
func (r *CdkStackResolver) filterAndAnnotate(ctx context.Context, stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	var existing []cdk.StackInfo
	for _, s := range stacks {
		result, err := r.existenceChecker.Check(ctx, s.Region, s.StackName)
		if err != nil {
			return nil, err
		}
		if !result.Exists {
			io.Logger.Info().Msgf("Stack %s not found in %s, skipping.", s.StackName, s.Region)
			continue
		}
		s.TerminationProtection = result.TerminationProtection
		existing = append(existing, s)
	}
	return existing, nil
}

func (r *CdkStackResolver) resolveRegions(ctx context.Context, stacks []cdk.StackInfo) error {
	defaultRegion, err := r.resolveDefaultRegion(ctx)
	if err != nil {
		return err
	}
	for i := range stacks {
		if stacks[i].Region == "unknown-region" || stacks[i].Region == "" {
			stacks[i].Region = defaultRegion
		}
	}
	return nil
}

func (r *CdkStackResolver) resolveDefaultRegion(ctx context.Context) (string, error) {
	if r.region != "" {
		return r.region, nil
	}
	cfg, err := r.configLoader.LoadConfig(ctx, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to resolve default region: %w", err)
	}
	return cfg.Region, nil
}
