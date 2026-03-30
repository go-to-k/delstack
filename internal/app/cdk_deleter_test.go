package app

import (
	"testing"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
)

func TestCdkDeleter_groupByRegion(t *testing.T) {
	io.NewLogger(false)
	d := &CdkDeleter{}

	tests := []struct {
		name            string
		stacks          []cdk.StackInfo
		wantRegionCount int
		wantRegions     map[string]int // region -> stack count
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
