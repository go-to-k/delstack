package app

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
)

// mockConfigLoader implements ConfigLoader for testing.
type mockConfigLoader struct {
	err error
}

func (m *mockConfigLoader) LoadConfig(_ context.Context, _, _ string) (aws.Config, error) {
	if m.err != nil {
		return aws.Config{}, m.err
	}
	return aws.Config{Region: "mock-region"}, nil
}

// passThroughAnalyzer builds a graph from the given stackNames with no dependencies.
type passThroughAnalyzer struct{}

func (a *passThroughAnalyzer) Analyze(_ context.Context, stackNames []string, _ *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
	deps := make(map[string][]string, len(stackNames))
	for _, name := range stackNames {
		deps[name] = nil
	}
	return buildGraph(deps), nil
}

func newTestCdkDeleter(executor StackExecutor) *CdkDeleter {
	if executor == nil {
		executor = &mockStackExecutor{}
	}
	return &CdkDeleter{
		profile:           "test",
		forceMode:         false,
		concurrencyNumber: 0,
		configLoader:      &mockConfigLoader{},
		analyzer:          &passThroughAnalyzer{},
		executor:          executor,
	}
}

func TestCdkDeleter_DeleteStacks_SingleRegion(t *testing.T) {
	io.NewLogger(false)

	executor := &mockStackExecutor{}
	d := newTestCdkDeleter(executor)

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
		{StackName: "StackB", Region: "us-east-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCdkDeleter_DeleteStacks_MultiRegionNoDeps(t *testing.T) {
	io.NewLogger(false)

	executor := &mockStackExecutor{}
	d := newTestCdkDeleter(executor)

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
		{StackName: "StackB", Region: "ap-northeast-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCdkDeleter_DeleteStacks_CrossRegionDeps(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deletionOrder []string

	executor := &mockStackExecutor{}
	executor.err = nil
	d := newTestCdkDeleter(&trackingExecutor{
		mu:    &mu,
		order: &deletionOrder,
	})

	stacks := []cdk.StackInfo{
		{StackName: "Edge", Region: "us-east-1"},
		{StackName: "Main", Region: "ap-northeast-1", Dependencies: []string{"Edge"}},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deletionOrder) != 2 {
		t.Fatalf("expected 2 stacks deleted, got %d: %v", len(deletionOrder), deletionOrder)
	}
	if deletionOrder[0] != "Main" {
		t.Errorf("expected Main to be deleted first, got %s", deletionOrder[0])
	}
	if deletionOrder[1] != "Edge" {
		t.Errorf("expected Edge to be deleted second, got %s", deletionOrder[1])
	}
}

func TestCdkDeleter_DeleteStacks_LoadConfigError(t *testing.T) {
	io.NewLogger(false)

	d := &CdkDeleter{
		profile:           "test",
		forceMode:         false,
		concurrencyNumber: 0,
		configLoader:      &mockConfigLoader{err: fmt.Errorf("config error")},
		executor:          &mockStackExecutor{},
	}

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCdkDeleter_DeleteStacks_DeletionError(t *testing.T) {
	io.NewLogger(false)

	d := newTestCdkDeleter(&mockStackExecutor{err: fmt.Errorf("deletion failed")})

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// trackingExecutor records deletion order for cross-region tests.
type trackingExecutor struct {
	mu    *sync.Mutex
	order *[]string
}

func (e *trackingExecutor) Execute(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
	e.mu.Lock()
	*e.order = append(*e.order, stack)
	e.mu.Unlock()
	return nil
}

func TestCdkDeleter_groupByRegion(t *testing.T) {
	io.NewLogger(false)
	d := &CdkDeleter{}

	tests := []struct {
		name            string
		stacks          []cdk.StackInfo
		wantRegionCount int
		wantRegions     map[string]int
	}{
		{
			name:            "single region",
			stacks:          []cdk.StackInfo{{StackName: "A", Region: "us-east-1"}, {StackName: "B", Region: "us-east-1"}},
			wantRegionCount: 1,
			wantRegions:     map[string]int{"us-east-1": 2},
		},
		{
			name:            "multiple regions",
			stacks:          []cdk.StackInfo{{StackName: "A", Region: "us-east-1"}, {StackName: "B", Region: "ap-northeast-1"}, {StackName: "C", Region: "us-east-1"}},
			wantRegionCount: 2,
			wantRegions:     map[string]int{"us-east-1": 2, "ap-northeast-1": 1},
		},
		{
			name:            "empty",
			stacks:          []cdk.StackInfo{},
			wantRegionCount: 0,
			wantRegions:     map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regionStacks, regions := d.groupByRegion(tt.stacks)
			if len(regions) != tt.wantRegionCount {
				t.Errorf("region count = %d, want %d", len(regions), tt.wantRegionCount)
			}
			for region, wantCount := range tt.wantRegions {
				if got := len(regionStacks[region]); got != wantCount {
					t.Errorf("region %s: stack count = %d, want %d", region, got, wantCount)
				}
			}
		})
	}
}

func TestCdkDeleter_hasCrossRegionDependencies(t *testing.T) {
	io.NewLogger(false)
	d := &CdkDeleter{}

	tests := []struct {
		name   string
		stacks []cdk.StackInfo
		want   bool
	}{
		{
			name: "no dependencies",
			stacks: []cdk.StackInfo{
				{StackName: "A", Region: "us-east-1"},
				{StackName: "B", Region: "ap-northeast-1"},
			},
			want: false,
		},
		{
			name: "same region dependency",
			stacks: []cdk.StackInfo{
				{StackName: "A", Region: "us-east-1"},
				{StackName: "B", Region: "us-east-1", Dependencies: []string{"A"}},
			},
			want: false,
		},
		{
			name: "cross region dependency",
			stacks: []cdk.StackInfo{
				{StackName: "Edge", Region: "us-east-1"},
				{StackName: "Main", Region: "ap-northeast-1", Dependencies: []string{"Edge"}},
			},
			want: true,
		},
		{
			name: "dependency on non-target stack",
			stacks: []cdk.StackInfo{
				{StackName: "A", Region: "us-east-1", Dependencies: []string{"External"}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stackMap := make(map[string]cdk.StackInfo)
			for _, s := range tt.stacks {
				stackMap[s.StackName] = s
			}
			got := d.hasCrossRegionDependencies(tt.stacks, stackMap)
			if got != tt.want {
				t.Errorf("hasCrossRegionDependencies = %v, want %v", got, tt.want)
			}
		})
	}
}
