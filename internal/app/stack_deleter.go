package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
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

func (d *StackDeleter) DeleteStacksConcurrently(
	ctx context.Context,
	stackNames []string,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
) error {
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()

	io.Logger.Info().Msg("Analyzing stack dependencies...")
	graph, err := cloudformationStackOperator.BuildDependencyGraph(ctx, stackNames)
	if err != nil {
		return fmt.Errorf("DependencyAnalysisError: failed to build dependency graph: %w", err)
	}

	cyclePath := graph.DetectCircularDependency()
	if len(cyclePath) > 0 {
		return fmt.Errorf("DependencyAnalysisError: circular dependency detected: %s", strings.Join(cyclePath, " -> "))
	}

	io.Logger.Info().Msgf("Starting deletion of %d stack(s) with dynamic scheduling...", len(stackNames))

	return d.deleteStacksDynamically(ctx, graph, config, operatorFactory)
}

func (d *StackDeleter) deleteStacksDynamically(
	ctx context.Context,
	graph *operation.StackDependencyGraph,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
) error {
	// Calculate reverse in-degree: how many stacks depend on this stack
	dependencies := graph.GetDependencies()
	allStacks := graph.GetAllStacks()
	totalStackCount := len(allStacks)
	reverseInDegree := make(map[string]int, totalStackCount)

	for stack := range allStacks {
		reverseInDegree[stack] = 0
	}
	for _, deps := range dependencies {
		for depStack := range deps {
			reverseInDegree[depStack]++
		}
	}

	completionChan := make(chan string, totalStackCount)
	errorChan := make(chan error)

	var deletedCount int
	deletedStacks := []string{}

	var weight int64
	if d.concurrencyNumber == UnspecifiedConcurrencyNumber {
		weight = int64(totalStackCount)
	} else {
		weight = min(int64(d.concurrencyNumber), int64(totalStackCount))
	}
	sem := semaphore.NewWeighted(weight)

	deleteCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	// for goroutine
	startDeletion := func(stackName string) {
		defer wg.Done()

		// Acquire semaphore inside goroutine to avoid blocking the main goroutine.
		// This ensures completion messages from other stacks are processed immediately,
		// even when waiting for concurrency limit.
		if err := sem.Acquire(deleteCtx, 1); err != nil {
			select {
			case errorChan <- err:
			default:
			}
			return
		}
		defer sem.Release(1)

		if err := d.deleteSingleStack(deleteCtx, stackName, config, operatorFactory, true); err != nil {
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
	for stack := range allStacks {
		if reverseInDegree[stack] == 0 {
			wg.Add(1)
			go startDeletion(stack)
		}
	}

	// Ensure cleanup on all exit paths (success, error, cancellation)
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
			// Update reverse in-degree and start newly available stacks
			deletedCount++
			deletedStacks = append(deletedStacks, deletedStackName)
			io.Logger.Info().Msgf("Progress: %d/%d stacks deleted [%s]", deletedCount, totalStackCount, strings.Join(deletedStacks, ", "))

			for depStack := range dependencies[deletedStackName] {
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

func (d *StackDeleter) deleteSingleStack(
	ctx context.Context,
	stack string,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
	isRootStack bool,
) error {
	operatorCollection := operation.NewOperatorCollection(config, operatorFactory)
	operatorManager := operation.NewOperatorManager(operatorCollection)
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()

	io.Logger.Info().Msgf("[%v]: Start deletion. Please wait a few minutes...", stack)

	if d.forceMode {
		if err := cloudformationStackOperator.RemoveDeletionPolicy(ctx, aws.String(stack)); err != nil {
			io.Logger.Error().Msgf("[%v]: Failed to remove deletion policy: %v", stack, err)
			return err
		}
	}

	if err := cloudformationStackOperator.DeleteCloudFormationStack(ctx, aws.String(stack), isRootStack, operatorManager); err != nil {
		io.Logger.Error().Msgf("[%v]: Failed to delete: %v", stack, err)
		return err
	}

	io.Logger.Info().Msgf("[%v]: Successfully deleted!!", stack)
	return nil
}
