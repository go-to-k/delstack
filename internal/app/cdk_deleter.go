package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
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

// deleteStacksWithCrossRegionDeps deletes stacks with cross-region dependencies
// using dynamic scheduling: as soon as a stack is deleted, dependent stacks
// become eligible immediately (without waiting for the entire level to complete).
func (a *App) deleteStacksWithCrossRegionDeps(ctx context.Context, stacks []cdk.StackInfo, stackMap map[string]cdk.StackInfo) error {
	io.Logger.Info().Msg("Cross-region dependencies detected. Deleting stacks in dependency order...")

	totalStackCount := len(stacks)

	// Build reverse in-degree and dependents map
	reverseInDegree := make(map[string]int, totalStackCount)
	dependents := make(map[string][]string) // stack -> stacks that depend on it (i.e. stacks that become unblocked when this stack is deleted)

	for _, s := range stacks {
		reverseInDegree[s.StackName] = 0
	}
	for _, s := range stacks {
		for _, dep := range s.Dependencies {
			if _, ok := stackMap[dep]; ok {
				reverseInDegree[dep]++
				dependents[s.StackName] = append(dependents[s.StackName], dep)
			}
		}
	}

	// Build per-region config and operatorFactory cache
	configCache := make(map[string]aws.Config)
	factoryCache := make(map[string]*operation.OperatorFactory)
	for _, s := range stacks {
		if _, ok := configCache[s.Region]; ok {
			continue
		}
		cfg, err := client.LoadAWSConfig(ctx, s.Region, a.Profile)
		if err != nil {
			return fmt.Errorf("failed to load AWS config for region %s: %w", s.Region, err)
		}
		configCache[s.Region] = cfg
		factoryCache[s.Region] = operation.NewOperatorFactory(cfg, a.ForceMode)
	}

	// Dynamic scheduling with channels (same pattern as deleteStacksDynamically)
	completionChan := make(chan string, totalStackCount)
	errorChan := make(chan error)

	var deletedCount int
	var deletedStacks []string

	var weight int64
	if a.ConcurrencyNumber == UnspecifiedConcurrencyNumber {
		weight = int64(totalStackCount)
	} else {
		weight = min(int64(a.ConcurrencyNumber), int64(totalStackCount))
	}
	sem := semaphore.NewWeighted(weight)

	deleteCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	startDeletion := func(stackName string) {
		defer wg.Done()

		if err := sem.Acquire(deleteCtx, 1); err != nil {
			select {
			case errorChan <- err:
			default:
			}
			return
		}
		defer sem.Release(1)

		s := stackMap[stackName]
		config := configCache[s.Region]
		operatorFactory := factoryCache[s.Region]

		deleter := NewStackDeleter(a.ForceMode, a.ConcurrencyNumber)
		if err := deleter.deleteSingleStack(deleteCtx, stackName, config, operatorFactory, true); err != nil {
			select {
			case errorChan <- err:
			default:
			}
			cancel()
			return
		}

		completionChan <- stackName
	}

	// Start initial deletions for stacks with reverse in-degree 0
	for _, s := range stacks {
		if reverseInDegree[s.StackName] == 0 {
			wg.Add(1)
			go startDeletion(s.StackName)
		}
	}

	defer func() {
		wg.Wait()
		close(completionChan)
		close(errorChan)
	}()

	for deletedCount < totalStackCount {
		select {
		case <-deleteCtx.Done():
			return deleteCtx.Err()

		case err := <-errorChan:
			return err

		case deletedStackName := <-completionChan:
			deletedCount++
			deletedStacks = append(deletedStacks, deletedStackName)
			io.Logger.Info().Msgf("Progress: %d/%d stacks deleted [%s]", deletedCount, totalStackCount, strings.Join(deletedStacks, ", "))

			for _, depStack := range dependents[deletedStackName] {
				reverseInDegree[depStack]--
				if reverseInDegree[depStack] == 0 {
					wg.Add(1)
					go startDeletion(depStack)
				}
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
