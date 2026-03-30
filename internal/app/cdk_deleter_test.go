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

func newTestCdkDeleter(
	deleteStacksFn func(ctx context.Context, stackNames []string, config aws.Config, operatorFactory *operation.OperatorFactory, forceMode bool, concurrencyNumber int) error,
	deleteSingleFn func(ctx context.Context, stack string, config aws.Config, operatorFactory *operation.OperatorFactory, forceMode bool, isRootStack bool) error,
) *CdkDeleter {
	if deleteStacksFn == nil {
		deleteStacksFn = func(_ context.Context, _ []string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ int) error {
			return nil
		}
	}
	if deleteSingleFn == nil {
		deleteSingleFn = func(_ context.Context, _ string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
			return nil
		}
	}
	return &CdkDeleter{
		profile:           "test",
		forceMode:         false,
		concurrencyNumber: 0,
		loadConfig: func(_ context.Context, _, _ string) (aws.Config, error) {
			return aws.Config{Region: "mock-region"}, nil
		},
		deleteStacksConcurrently: deleteStacksFn,
		deleteSingleStack:        deleteSingleFn,
	}
}

func TestCdkDeleter_DeleteStacks_SingleRegion(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deleted []string
	d := newTestCdkDeleter(
		func(_ context.Context, stackNames []string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ int) error {
			mu.Lock()
			deleted = append(deleted, stackNames...)
			mu.Unlock()
			return nil
		},
		nil,
	)

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
		{StackName: "StackB", Region: "us-east-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deleted) != 2 {
		t.Errorf("expected 2 stacks deleted, got %d: %v", len(deleted), deleted)
	}
}

func TestCdkDeleter_DeleteStacks_MultiRegionNoDeps(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deleted []string
	d := newTestCdkDeleter(
		func(_ context.Context, stackNames []string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ int) error {
			mu.Lock()
			deleted = append(deleted, stackNames...)
			mu.Unlock()
			return nil
		},
		nil,
	)

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
		{StackName: "StackB", Region: "ap-northeast-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(deleted) != 2 {
		t.Errorf("expected 2 stacks deleted, got %d: %v", len(deleted), deleted)
	}
}

func TestCdkDeleter_DeleteStacks_CrossRegionDeps(t *testing.T) {
	io.NewLogger(false)

	var mu sync.Mutex
	var deletionOrder []string

	d := newTestCdkDeleter(
		nil,
		func(_ context.Context, stack string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
			mu.Lock()
			deletionOrder = append(deletionOrder, stack)
			mu.Unlock()
			return nil
		},
	)

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
		loadConfig: func(_ context.Context, _, _ string) (aws.Config, error) {
			return aws.Config{}, fmt.Errorf("config error")
		},
		deleteStacksConcurrently: func(_ context.Context, _ []string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ int) error {
			return nil
		},
		deleteSingleStack: func(_ context.Context, _ string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ bool) error {
			return nil
		},
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

	d := newTestCdkDeleter(
		func(_ context.Context, _ []string, _ aws.Config, _ *operation.OperatorFactory, _ bool, _ int) error {
			return fmt.Errorf("deletion failed")
		},
		nil,
	)

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "us-east-1"},
	}

	err := d.DeleteStacks(context.Background(), stacks)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
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
