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

// mockDependencyAnalyzer returns a pre-built graph.
type mockDependencyAnalyzer struct {
	graph *operation.StackDependencyGraph
	err   error
}

func (m *mockDependencyAnalyzer) Analyze(_ context.Context, _ []string, _ *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
	return m.graph, m.err
}

// mockStackExecutor records deletions.
type mockStackExecutor struct {
	mu      sync.Mutex
	deleted []string
	err     error
}

func (m *mockStackExecutor) Execute(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
	m.mu.Lock()
	m.deleted = append(m.deleted, stack)
	m.mu.Unlock()
	return m.err
}

func buildGraph(deps map[string][]string) *operation.StackDependencyGraph {
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

func TestStackDeleter_DeleteStacksConcurrently_IndependentStacks(t *testing.T) {
	io.NewLogger(false)

	executor := &mockStackExecutor{}
	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: UnspecifiedConcurrencyNumber,
		analyzer:          &mockDependencyAnalyzer{graph: buildGraph(map[string][]string{"A": {}, "B": {}, "C": {}})},
		executor:          executor,
	}

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B", "C"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executor.deleted) != 3 {
		t.Errorf("expected 3 stacks deleted, got %d: %v", len(executor.deleted), executor.deleted)
	}
}

func TestStackDeleter_DeleteStacksConcurrently_DependencyOrder(t *testing.T) {
	io.NewLogger(false)

	executor := &mockStackExecutor{}
	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: UnspecifiedConcurrencyNumber,
		analyzer:          &mockDependencyAnalyzer{graph: buildGraph(map[string][]string{"A": {}, "B": {"A"}})},
		executor:          executor,
	}

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executor.deleted) != 2 {
		t.Fatalf("expected 2 stacks deleted, got %d: %v", len(executor.deleted), executor.deleted)
	}
	if executor.deleted[0] != "B" {
		t.Errorf("expected B deleted first, got %s", executor.deleted[0])
	}
	if executor.deleted[1] != "A" {
		t.Errorf("expected A deleted second, got %s", executor.deleted[1])
	}
}

func TestStackDeleter_DeleteStacksConcurrently_ComplexDeps(t *testing.T) {
	io.NewLogger(false)

	executor := &mockStackExecutor{}
	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: UnspecifiedConcurrencyNumber,
		analyzer: &mockDependencyAnalyzer{graph: buildGraph(map[string][]string{
			"A": {}, "B": {}, "C": {"A"}, "D": {"A"}, "E": {"B"}, "F": {"C", "D", "E"},
		})},
		executor: executor,
	}

	err := d.DeleteStacksConcurrently(context.Background(), []string{"A", "B", "C", "D", "E", "F"}, aws.Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(executor.deleted) != 6 {
		t.Fatalf("expected 6 stacks deleted, got %d: %v", len(executor.deleted), executor.deleted)
	}

	indexOf := make(map[string]int)
	for i, s := range executor.deleted {
		indexOf[s] = i
	}
	assertBefore := func(a, b string) {
		if indexOf[a] >= indexOf[b] {
			t.Errorf("expected %s before %s, got order: %v", a, b, executor.deleted)
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
		analyzer:          &mockDependencyAnalyzer{err: fmt.Errorf("graph error")},
		executor:          &mockStackExecutor{},
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

	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: UnspecifiedConcurrencyNumber,
		analyzer:          &mockDependencyAnalyzer{graph: buildGraph(map[string][]string{"A": {"B"}, "B": {"A"}})},
		executor:          &mockStackExecutor{},
	}

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

	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: UnspecifiedConcurrencyNumber,
		analyzer:          &mockDependencyAnalyzer{graph: buildGraph(map[string][]string{"Stack": {}})},
		executor:          &mockStackExecutor{err: fmt.Errorf("deletion failed")},
	}

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

	d := &StackDeleter{
		forceMode:         false,
		concurrencyNumber: 1,
		analyzer:          &mockDependencyAnalyzer{graph: buildGraph(map[string][]string{"A": {}, "B": {}, "C": {}})},
		executor:          &mockStackExecutor{},
	}
	// Override executor with concurrency tracking
	d.executor = &concurrencyTrackingExecutor{
		mu:                &mu,
		maxConcurrent:     &maxConcurrent,
		currentConcurrent: &currentConcurrent,
		deleted:           &deleted,
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

type concurrencyTrackingExecutor struct {
	mu                *sync.Mutex
	maxConcurrent     *int
	currentConcurrent *int
	deleted           *[]string
}

func (e *concurrencyTrackingExecutor) Execute(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
	e.mu.Lock()
	*e.currentConcurrent++
	if *e.currentConcurrent > *e.maxConcurrent {
		*e.maxConcurrent = *e.currentConcurrent
	}
	e.mu.Unlock()

	e.mu.Lock()
	*e.currentConcurrent--
	*e.deleted = append(*e.deleted, stack)
	e.mu.Unlock()
	return nil
}
