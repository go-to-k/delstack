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
	configLoader     IConfigLoader
	existenceChecker IStackExistenceChecker
}

func NewCdkStackResolver(selector *CdkStackSelector, profile, region string, forceMode bool) *CdkStackResolver {
	return &CdkStackResolver{
		selector:         selector,
		region:           region,
		configLoader:     &ConfigLoader{},
		existenceChecker: NewStackExistenceChecker(profile, forceMode),
	}
}

// Resolve takes raw stacks from the CDK manifest and returns the final list
// of stacks to delete: regions resolved, filtered by pattern/interactive selection,
// and verified to exist in AWS.
func (r *CdkStackResolver) Resolve(ctx context.Context, stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	// Resolve unknown regions
	if err := r.resolveRegions(ctx, stacks); err != nil {
		return nil, err
	}

	// Filter stacks by -s or -i
	selected, err := r.selector.Select(stacks)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		io.Logger.Info().Msg("No stacks selected.")
		return nil, nil
	}

	// Filter out stacks that don't exist in AWS
	existing, err := r.filterExisting(ctx, selected)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		io.Logger.Info().Msg("No deployed stacks found.")
		return nil, nil
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

func (r *CdkStackResolver) filterExisting(ctx context.Context, stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	var existing []cdk.StackInfo
	for _, s := range stacks {
		exists, err := r.existenceChecker.Exists(ctx, s.Region, s.StackName)
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
