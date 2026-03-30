package app

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
)

// mockGraph builds a StackDependencyGraph from a simple dependency map.
// deps: stackName -> list of stacks it depends on (i.e. must be deleted before this stack)
func mockGraph(deps map[string][]string) *operation.StackDependencyGraph {
	stackNames := make([]string, 0, len(deps))
	for name := range deps {
		stackNames = append(stackNames, name)
	}
	graph := operation.NewStackDependencyGraph(stackNames)
	for from, toList := range deps {
		for _, to := range toList {
			graph.AddDependency(from, to)
		}
	}
	return graph
}

func newTestStackDeleter(
	graph *operation.StackDependencyGraph,
	deleteFn func(ctx context.Context, stack string, config aws.Config, operatorFactory *operation.OperatorFactory, forceMode bool, isRootStack bool) error,
) *StackDeleter {
	return &StackDeleter{
		forceMode:         false,
		concurrencyNumber: UnspecifiedConcurrencyNumber,
		buildDependencyGraph: func(_ context.Context, _ []string, _ *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
			return graph, nil
		},
		deleteSingleStack: deleteFn,
	}
}

func TestStackDeleter_DeleteStacksConcurrently_IndependentStacks(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deleted []string

	graph := mockGraph(map[string][]string{
		"StackA": {},
		"StackB": {},
		"StackC": {},
	})

	d := newTestStackDeleter(graph, func(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
		mu.Lock()
		deleted = append(deleted, stack)
		mu.Unlock()
		return nil
	})

	err := d.DeleteStacksConcurrently(context.Background(), []string{"StackA", "StackB", "StackC"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deleted) != 3 {
		t.Errorf("expected 3 stacks deleted, got %d: %v", len(deleted), deleted)
	}
}

func TestStackDeleter_DeleteStacksConcurrently_DependencyOrder(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deleted []string

	// B depends on A (B must be deleted before A)
	graph := mockGraph(map[string][]string{
		"A": {},
		"B": {"A"},
	})

	d := newTestStackDeleter(graph, func(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
		mu.Lock()
		deleted = append(deleted, stack)
		mu.Unlock()
		return nil
	})

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deleted) != 2 {
		t.Fatalf("expected 2 stacks deleted, got %d: %v", len(deleted), deleted)
	}
	if deleted[0] != "B" {
		t.Errorf("expected B deleted first, got %s", deleted[0])
	}
	if deleted[1] != "A" {
		t.Errorf("expected A deleted second, got %s", deleted[1])
	}
}

func TestStackDeleter_DeleteStacksConcurrently_ComplexDeps(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deleted []string

	// F -> C, D, E; C -> A; D -> A; E -> B
	graph := mockGraph(map[string][]string{
		"A": {},
		"B": {},
		"C": {"A"},
		"D": {"A"},
		"E": {"B"},
		"F": {"C", "D", "E"},
	})

	d := newTestStackDeleter(graph, func(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
		mu.Lock()
		deleted = append(deleted, stack)
		mu.Unlock()
		return nil
	})

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B", "C", "D", "E", "F"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deleted) != 6 {
		t.Fatalf("expected 6 stacks deleted, got %d: %v", len(deleted), deleted)
	}

	// Verify ordering constraints
	indexOf := make(map[string]int)
	for i, s := range deleted {
		indexOf[s] = i
	}
	assertBefore := func(a, b string) {
		if indexOf[a] >= indexOf[b] {
			t.Errorf("expected %s before %s, got order: %v", a, b, deleted)
		}
	}
	assertBefore("F", "C")
	assertBefore("F", "D")
	assertBefore("F", "E")
	assertBefore("C", "A")
	assertBefore("D", "A")
	assertBefore("E", "B")
}

func TestStackDeleter_DeleteStacksConcurrently_BuildGraphError(t *testing.T) {
	io.NewLogger(false)

	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: 0,
		buildDependencyGraph: func(_ context.Context, _ []string, _ *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
			return nil, fmt.Errorf("graph error")
		},
		deleteSingleStack: defaultDeleteSingleStack,
	}

	err := d.DeleteStacksConcurrently(context.Background(), []string{"Stack"}, aws.Config{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsString(err.Error(), "DependencyAnalysisError") {
		t.Errorf("expected DependencyAnalysisError, got %q", err.Error())
	}
}

func TestStackDeleter_DeleteStacksConcurrently_CircularDependency(t *testing.T) {
	io.NewLogger(false)

	graph := mockGraph(map[string][]string{
		"A": {"B"},
		"B": {"A"},
	})

	d := newTestStackDeleter(graph, func(_ context.Context, _ string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
		return nil
	})

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B"}, aws.Config{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsString(err.Error(), "circular dependencies") {
		t.Errorf("expected circular dependency error, got %q", err.Error())
	}
}

func TestStackDeleter_DeleteStacksConcurrently_DeletionError(t *testing.T) {
	io.NewLogger(false)

	graph := mockGraph(map[string][]string{
		"Stack": {},
	})

	d := newTestStackDeleter(graph, func(_ context.Context, _ string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
		return fmt.Errorf("deletion failed")
	})

	err := d.DeleteStacksConcurrently(context.Background(), []string{"Stack"}, aws.Config{}, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStackDeleter_DeleteStacksConcurrently_Concurrency(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	maxConcurrent := 0
	currentConcurrent := 0
	var deleted []string

	graph := mockGraph(map[string][]string{
		"A": {},
		"B": {},
		"C": {},
	})

	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: 1, // limit to 1
		buildDependencyGraph: func(_ context.Context, _ []string, _ *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
			return graph, nil
		},
		deleteSingleStack: func(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			mu.Lock()
			currentConcurrent--
			deleted = append(deleted, stack)
			mu.Unlock()
			return nil
		},
	}

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B", "C"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deleted) != 3 {
		t.Errorf("expected 3 stacks deleted, got %d", len(deleted))
	}
	if maxConcurrent > 1 {
		t.Errorf("expected max concurrency 1, got %d", maxConcurrent)
	}
}
