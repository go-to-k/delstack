package app

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
)

func (a *App) deleteCdkStacks(ctx context.Context, stacks []cdk.StackInfo) error {
	regionStacks, regions := groupByRegion(stacks)

	stackMap := make(map[string]cdk.StackInfo)
	for _, s := range stacks {
		stackMap[s.StackName] = s
	}

	if hasCrossRegionDependencies(stacks, stackMap) {
		return a.deleteStacksWithCrossRegionDeps(ctx, stacks, stackMap)
	}

	// No cross-region deps: delete all regions in parallel
	var eg errgroup.Group
	for _, region := range regions {
		regionStackInfos := regionStacks[region]
		eg.Go(func() error {
			return a.deleteStacksInRegion(ctx, region, regionStackInfos)
		})
	}
	return eg.Wait()
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

	reverseInDegree := make(map[string]int)
	dependents := make(map[string][]string)

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

	deleted := make(map[string]bool)
	configCache := make(map[string]aws.Config)

	for len(deleted) < len(stacks) {
		var ready []cdk.StackInfo
		for _, s := range stacks {
			if !deleted[s.StackName] && reverseInDegree[s.StackName] == 0 {
				ready = append(ready, s)
			}
		}

		if len(ready) == 0 {
			return fmt.Errorf("circular dependency detected among remaining stacks")
		}

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

		for _, s := range ready {
			deleted[s.StackName] = true
			for _, dep := range dependents[s.StackName] {
				reverseInDegree[dep]--
			}
		}
	}

	return nil
}

func groupByRegion(stacks []cdk.StackInfo) (map[string][]cdk.StackInfo, []string) {
	regionStacks := make(map[string][]cdk.StackInfo)
	for _, s := range stacks {
		regionStacks[s.Region] = append(regionStacks[s.Region], s)
	}

	regions := make([]string, 0, len(regionStacks))
	for region := range regionStacks {
		regions = append(regions, region)
	}
	return regionStacks, regions
}

func hasCrossRegionDependencies(stacks []cdk.StackInfo, stackMap map[string]cdk.StackInfo) bool {
	for _, s := range stacks {
		for _, dep := range s.Dependencies {
			if depStack, ok := stackMap[dep]; ok {
				if depStack.Region != s.Region {
					return true
				}
			}
		}
	}
	return false
}
