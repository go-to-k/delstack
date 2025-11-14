package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/internal/resourcetype"
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

	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(resourcetype.GetResourceTypes())

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

	return d.deleteStacksDynamically(ctx, graph, targetStacksMap, config, operatorFactory)
}

func (d *StackDeleter) deleteStacksDynamically(
	ctx context.Context,
	graph *operation.StackDependencyGraph,
	targetStacksMap map[string]targetStack,
	config aws.Config,
	operatorFactory *operation.OperatorFactory,
) error {
	// Calculate reverse in-degree: how many stacks depend on this stack
	reverseInDegree := make(map[string]int)
	dependencies := graph.GetDependencies()
	allStacks := graph.GetAllStacks()

	for stack := range allStacks {
		reverseInDegree[stack] = 0
	}
	for _, deps := range dependencies {
		for depStack := range deps {
			reverseInDegree[depStack]++
		}
	}

	// Initialize queue with stacks that have reverse in-degree 0
	var stateMutex sync.Mutex // Protects queue, reverseInDegree, and deletedStacks
	queue := []string{}
	for stack := range allStacks {
		if reverseInDegree[stack] == 0 {
			queue = append(queue, stack)
		}
	}

	// Channel to signal stack deletion completion
	completionChan := make(chan string, len(allStacks))
	errorChan := make(chan error)

	// Track deletion progress
	var deletedCount int
	totalStacks := len(allStacks)
	deletedStacks := []string{}

	// Semaphore for concurrency control
	var weight int64
	if d.concurrencyNumber == UnspecifiedConcurrencyNumber {
		weight = int64(totalStacks)
	} else {
		weight = min(int64(d.concurrencyNumber), int64(totalStacks))
	}
	sem := semaphore.NewWeighted(weight)

	// Context for goroutines
	deleteCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Worker to process deletions
	var wg sync.WaitGroup

	// Function to start deletion of a stack
	startDeletion := func(stackName string) {
		if err := sem.Acquire(deleteCtx, 1); err != nil {
			select {
			case errorChan <- err:
			default:
			}
			return
		}

		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			defer sem.Release(1)

			stack := targetStacksMap[name]
			if err := d.deleteSingleStack(deleteCtx, stack, config, operatorFactory, true); err != nil {
				select {
				case errorChan <- err:
				default:
				}
				cancel() // Cancel all other operations on error
				return
			}

			// Signal completion
			completionChan <- name
		}(stackName)
	}

	// Start initial deletions
	for _, stackName := range queue {
		startDeletion(stackName)
	}
	stateMutex.Lock()
	queue = []string{} // Clear the queue
	stateMutex.Unlock()

	// Process completions and start new deletions
	for deletedCount < totalStacks {
		select {
		case <-deleteCtx.Done():
			// Wait for all goroutines to finish
			wg.Wait()
			return deleteCtx.Err()

		case err := <-errorChan:
			// Wait for all goroutines to finish
			wg.Wait()
			return err

		case deletedStackName := <-completionChan:
			// Update reverse in-degree and find newly available stacks
			stateMutex.Lock()
			deletedCount++
			deletedStacks = append(deletedStacks, deletedStackName)
			io.Logger.Info().Msgf("Progress: %d/%d stacks deleted [%s]", deletedCount, totalStacks, strings.Join(deletedStacks, ", "))

			for depStack := range dependencies[deletedStackName] {
				reverseInDegree[depStack]--
				if reverseInDegree[depStack] == 0 {
					queue = append(queue, depStack)
				}
			}

			// Start deletions for newly available stacks
			newlyAvailableStacks := make([]string, len(queue))
			copy(newlyAvailableStacks, queue)
			queue = []string{} // Clear the queue
			stateMutex.Unlock()

			for _, stackName := range newlyAvailableStacks {
				startDeletion(stackName)
			}
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(completionChan)
	close(errorChan)

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
