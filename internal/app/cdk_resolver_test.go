package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
)

type mockExistenceChecker struct {
	existingStacks map[string]bool
	err            error
}

func (m *mockExistenceChecker) Exists(_ context.Context, _, stackName string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.existingStacks[stackName], nil
}

func newTestResolver(patterns []string, existingStacks map[string]bool) *CdkStackResolver {
	selector := NewCdkStackSelector(patterns, false)
	return &CdkStackResolver{
		selector:         selector,
		region:           "us-east-1",
		configLoader:     &mockConfigLoader{},
		existenceChecker: &mockExistenceChecker{existingStacks: existingStacks},
	}
}

func TestCdkStackResolver_Resolve(t *testing.T) {
	io.NewLogger(false)

	tests := []struct {
		name           string
		stacks         []cdk.StackInfo
		patterns       []string
		existingStacks map[string]bool
		wantNames      []string
		wantErr        bool
	}{
		{
			name: "exact match, all exist",
			stacks: []cdk.StackInfo{
				{StackName: "StackA", Region: "us-east-1"},
				{StackName: "StackB", Region: "us-east-1"},
			},
			patterns:       []string{"StackA"},
			existingStacks: map[string]bool{"StackA": true},
			wantNames:      []string{"StackA"},
		},
		{
			name: "glob match, all exist",
			stacks: []cdk.StackInfo{
				{StackName: "ApiStack", Region: "us-east-1"},
				{StackName: "ApiWorkerStack", Region: "us-east-1"},
				{StackName: "WebStack", Region: "us-east-1"},
			},
			patterns:       []string{"Api*"},
			existingStacks: map[string]bool{"ApiStack": true, "ApiWorkerStack": true},
			wantNames:      []string{"ApiStack", "ApiWorkerStack"},
		},
		{
			name: "glob match, some not deployed",
			stacks: []cdk.StackInfo{
				{StackName: "ApiStack", Region: "us-east-1"},
				{StackName: "ApiWorkerStack", Region: "us-east-1"},
				{StackName: "WebStack", Region: "us-east-1"},
			},
			patterns:       []string{"Api*"},
			existingStacks: map[string]bool{"ApiStack": true, "ApiWorkerStack": false},
			wantNames:      []string{"ApiStack"},
		},
		{
			name: "no stacks selected returns nil",
			stacks: []cdk.StackInfo{
				{StackName: "StackA", Region: "us-east-1"},
			},
			patterns:       []string{"ZZZ*"},
			existingStacks: map[string]bool{},
			wantNames:      nil,
		},
		{
			name: "selected but none exist returns nil",
			stacks: []cdk.StackInfo{
				{StackName: "StackA", Region: "us-east-1"},
			},
			patterns:       []string{"StackA"},
			existingStacks: map[string]bool{"StackA": false},
			wantNames:      nil,
		},
		{
			name: "exact match not found in manifest returns error",
			stacks: []cdk.StackInfo{
				{StackName: "StackA", Region: "us-east-1"},
			},
			patterns:       []string{"NonExistent"},
			existingStacks: map[string]bool{},
			wantErr:        true,
		},
		{
			name: "resolves unknown regions",
			stacks: []cdk.StackInfo{
				{StackName: "StackA", Region: "unknown-region"},
				{StackName: "StackB", Region: ""},
			},
			patterns:       []string{"*"},
			existingStacks: map[string]bool{"StackA": true, "StackB": true},
			wantNames:      []string{"StackA", "StackB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := newTestResolver(tt.patterns, tt.existingStacks)
			result, err := resolver.Resolve(context.Background(), tt.stacks)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var gotNames []string
			for _, s := range result {
				gotNames = append(gotNames, s.StackName)
			}
			if !equalStrings(gotNames, tt.wantNames) {
				t.Errorf("resolved stacks = %v, want %v", gotNames, tt.wantNames)
			}
		})
	}
}

func TestCdkStackResolver_ResolveRegions(t *testing.T) {
	io.NewLogger(false)

	stacks := []cdk.StackInfo{
		{StackName: "StackA", Region: "unknown-region"},
		{StackName: "StackB", Region: ""},
		{StackName: "StackC", Region: "ap-northeast-1"},
	}

	resolver := newTestResolver(nil, nil)
	err := resolver.resolveRegions(context.Background(), stacks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"us-east-1", "us-east-1", "ap-northeast-1"}
	for i, s := range stacks {
		if s.Region != expected[i] {
			t.Errorf("stacks[%d].Region = %q, want %q", i, s.Region, expected[i])
		}
	}
}

func TestCdkStackResolver_ResolveDefaultRegion(t *testing.T) {
	io.NewLogger(false)

	t.Run("uses explicit region when set", func(t *testing.T) {
		resolver := &CdkStackResolver{
			region:       "eu-west-1",
			configLoader: &mockConfigLoader{},
		}
		region, err := resolver.resolveDefaultRegion(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if region != "eu-west-1" {
			t.Errorf("region = %q, want %q", region, "eu-west-1")
		}
	})

	t.Run("falls back to config when region is empty", func(t *testing.T) {
		resolver := &CdkStackResolver{
			region:       "",
			configLoader: &mockConfigLoader{},
		}
		region, err := resolver.resolveDefaultRegion(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// mockConfigLoader returns "mock-region"
		if region != "mock-region" {
			t.Errorf("region = %q, want %q", region, "mock-region")
		}
	})

	t.Run("returns error when config loader fails", func(t *testing.T) {
		resolver := &CdkStackResolver{
			region:       "",
			configLoader: &mockConfigLoader{err: fmt.Errorf("config error")},
		}
		_, err := resolver.resolveDefaultRegion(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCdkStackResolver_FilterExisting(t *testing.T) {
	io.NewLogger(false)

	t.Run("filters out non-existing stacks", func(t *testing.T) {
		resolver := newTestResolver(nil, map[string]bool{
			"StackA": true,
			"StackB": false,
			"StackC": true,
		})
		stacks := []cdk.StackInfo{
			{StackName: "StackA", Region: "us-east-1"},
			{StackName: "StackB", Region: "us-east-1"},
			{StackName: "StackC", Region: "us-east-1"},
		}
		result, err := resolver.filterExisting(context.Background(), stacks)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 stacks, got %d", len(result))
		}
		if result[0].StackName != "StackA" || result[1].StackName != "StackC" {
			t.Errorf("unexpected stacks: %v", result)
		}
	})

	t.Run("returns error on checker failure", func(t *testing.T) {
		resolver := &CdkStackResolver{
			existenceChecker: &mockExistenceChecker{err: fmt.Errorf("aws error")},
		}
		stacks := []cdk.StackInfo{
			{StackName: "StackA", Region: "us-east-1"},
		}
		_, err := resolver.filterExisting(context.Background(), stacks)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
