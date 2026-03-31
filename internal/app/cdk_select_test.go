package app

import (
	"sort"
	"testing"

	"github.com/go-to-k/delstack/internal/cdk"
)

func makeStacks(names ...string) []cdk.StackInfo {
	stacks := make([]cdk.StackInfo, len(names))
	for i, n := range names {
		stacks[i] = cdk.StackInfo{StackName: n, Region: "us-east-1"}
	}
	return stacks
}

func stackNames(stacks []cdk.StackInfo) []string {
	names := make([]string, len(stacks))
	for i, s := range stacks {
		names[i] = s.StackName
	}
	sort.Strings(names)
	return names
}

func TestCdkStackSelector_matchByPatterns(t *testing.T) {
	allStacks := makeStacks(
		"MyStack",
		"CdkStack",
		"CdkStackA",
		"CdkStackB",
		"AppProdStack",
		"AppDevStack",
		"TestStack",
	)

	tests := []struct {
		name        string
		patterns    []string
		wantNames   []string
		wantUnmatch []string
		wantErr     bool
	}{
		{
			name:      "exact match single",
			patterns:  []string{"MyStack"},
			wantNames: []string{"MyStack"},
		},
		{
			name:      "exact match multiple",
			patterns:  []string{"MyStack", "TestStack"},
			wantNames: []string{"MyStack", "TestStack"},
		},
		{
			name:        "exact match not found",
			patterns:    []string{"NonExistent"},
			wantNames:   nil,
			wantUnmatch: []string{"NonExistent"},
		},
		{
			name:      "glob star suffix",
			patterns:  []string{"Cdk*"},
			wantNames: []string{"CdkStack", "CdkStackA", "CdkStackB"},
		},
		{
			name:      "glob star prefix",
			patterns:  []string{"*Stack"},
			wantNames: []string{"AppDevStack", "AppProdStack", "CdkStack", "MyStack", "TestStack"},
		},
		{
			name:      "glob star middle",
			patterns:  []string{"App*Stack"},
			wantNames: []string{"AppDevStack", "AppProdStack"},
		},
		{
			name:      "glob match all",
			patterns:  []string{"*"},
			wantNames: []string{"AppDevStack", "AppProdStack", "CdkStack", "CdkStackA", "CdkStackB", "MyStack", "TestStack"},
		},
		{
			name:      "glob question mark",
			patterns:  []string{"CdkStack?"},
			wantNames: []string{"CdkStackA", "CdkStackB"},
		},
		{
			name:      "glob character class",
			patterns:  []string{"CdkStack[AB]"},
			wantNames: []string{"CdkStackA", "CdkStackB"},
		},
		{
			name:      "mix of exact and glob",
			patterns:  []string{"MyStack", "Cdk*"},
			wantNames: []string{"CdkStack", "CdkStackA", "CdkStackB", "MyStack"},
		},
		{
			name:        "mix with unmatched exact",
			patterns:    []string{"NonExistent", "Cdk*"},
			wantNames:   []string{"CdkStack", "CdkStackA", "CdkStackB"},
			wantUnmatch: []string{"NonExistent"},
		},
		{
			name:      "glob matching nothing is not an error",
			patterns:  []string{"ZZZ*"},
			wantNames: nil,
		},
		{
			name:     "invalid glob pattern",
			patterns: []string{"[invalid"},
			wantErr:  true,
		},
		{
			name:      "no duplicate when exact and glob both match",
			patterns:  []string{"CdkStack", "Cdk*"},
			wantNames: []string{"CdkStack", "CdkStackA", "CdkStackB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewCdkStackSelector(tt.patterns, false, false)
			selected, unmatched, err := s.matchByPatterns(allStacks)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotNames := stackNames(selected)
			if len(gotNames) == 0 {
				gotNames = nil
			}
			if !equalStrings(gotNames, tt.wantNames) {
				t.Errorf("matched stacks = %v, want %v", gotNames, tt.wantNames)
			}

			sort.Strings(unmatched)
			if len(unmatched) == 0 {
				unmatched = nil
			}
			wantUnmatch := tt.wantUnmatch
			sort.Strings(wantUnmatch)
			if !equalStrings(unmatched, wantUnmatch) {
				t.Errorf("unmatched = %v, want %v", unmatched, wantUnmatch)
			}
		})
	}
}

func TestCdkStackSelector_isGlobPattern(t *testing.T) {
	s := &CdkStackSelector{}

	tests := []struct {
		pattern string
		want    bool
	}{
		{"MyStack", false},
		{"Cdk*", true},
		{"Stack?", true},
		{"[AB]Stack", true},
		{"plain-name-123", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			if got := s.isGlobPattern(tt.pattern); got != tt.want {
				t.Errorf("isGlobPattern(%q) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
