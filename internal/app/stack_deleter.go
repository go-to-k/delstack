package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type StackDeleter struct {
	forceMode         bool
	concurrencyNumber int
}

func NewStackDeleter(forceMode bool, concurrencyNumber int) *StackDeleter {
	return &StackDeleter{
		forceMode:         forceMode,
		concurrencyNumber: concurrencyNumber,
	}
}

func (d *StackDeleter) DeleteStacksSequentially(
	ctx context.Context,
	targetStacks []targetStack,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
) error {
	isRootStack := true
	for _, stack := range targetStacks {
		if err := d.deleteSingleStack(ctx, stack, config, operatorFactory, isRootStack); err != nil {
			return err
		}
	}
	return nil
}

func (d *StackDeleter) DeleteStacksConcurrently(
	ctx context.Context,
	targetStacks []targetStack,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
) error {
	stackNames := make([]string, 0, len(targetStacks))
	targetStacksMap := make(map[string]targetStack)
	for _, stack := range targetStacks {
		stackNames = append(stackNames, stack.stackName)
		targetStacksMap[stack.stackName] = stack
	}

	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator([]string{})

	io.Logger.Info().Msg("Analyzing stack dependencies...")
	graph, err := cloudformationStackOperator.BuildDependencyGraph(ctx, stackNames)
	if err != nil {
		return fmt.Errorf("DependencyAnalysisError: failed to build dependency graph: %w", err)
	}

	cyclePath, err := graph.DetectCircularDependency()
	if err != nil {
		io.Logger.Error().Msgf("Circular dependency detected: %s", strings.Join(cyclePath, " -> "))
		return err
	}

	deletionGroups := graph.GetDeletionGroups()
	io.Logger.Info().Msgf("Deletion will be performed in %d group(s)", len(deletionGroups))

	for groupIndex, group := range deletionGroups {
		io.Logger.Info().Msgf("Group %d: Deleting %d stack(s) concurrently: %v",
			groupIndex+1, len(group), strings.Join(group, ", "))

		if err := d.deleteStackGroup(ctx, group, targetStacksMap, config, operatorFactory); err != nil {
			return fmt.Errorf("ConcurrentDeleteError: failed to delete group %d: %w", groupIndex+1, err)
		}
	}

	return nil
}

func (d *StackDeleter) deleteStackGroup(
	ctx context.Context,
	stackNames []string,
	targetStacksMap map[string]targetStack,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
) error {
	eg, ctx := errgroup.WithContext(ctx)

	var sem *semaphore.Weighted
	var weight int64

	if d.concurrencyNumber == UnspecifiedConcurrencyNumber {
		weight = int64(len(stackNames))
	} else {
		weight = min(int64(d.concurrencyNumber), int64(len(stackNames)))
	}
	sem = semaphore.NewWeighted(weight)

	isRootStack := true
	for _, stackName := range stackNames {
		stack := targetStacksMap[stackName]

		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return d.deleteSingleStack(ctx, stack, config, operatorFactory, isRootStack)
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (d *StackDeleter) deleteSingleStack(
	ctx context.Context,
	stack targetStack,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
	isRootStack bool,
) error {
	operatorCollection := operation.NewOperatorCollection(config, operatorFactory, stack.targetResourceTypes)
	operatorManager := operation.NewOperatorManager(operatorCollection)
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(stack.targetResourceTypes)

	io.Logger.Info().Msgf("%v: Start deletion. Please wait a few minutes...", stack.stackName)

	if d.forceMode {
		if err := cloudformationStackOperator.RemoveDeletionPolicy(ctx, aws.String(stack.stackName)); err != nil {
			io.Logger.Error().Msgf("%v: Failed to remove deletion policy: %v", stack.stackName, err)
			return err
		}
	}

	if err := cloudformationStackOperator.DeleteCloudFormationStack(ctx, aws.String(stack.stackName), isRootStack, operatorManager); err != nil {
		io.Logger.Error().Msgf("%v: Failed to delete: %v", stack.stackName, err)
		return err
	}

	io.Logger.Info().Msgf("%v: Successfully deleted!!", stack.stackName)
	return nil
}
