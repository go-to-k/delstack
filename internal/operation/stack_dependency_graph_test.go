package operation

import (
	"reflect"
	"testing"
)

func TestNewStackDependencyGraph(t *testing.T) {
	type args struct {
		stackNames []string
	}

	cases := []struct {
		name string
		args args
		want *StackDependencyGraph
	}{
		{
			name: "create graph with multiple stacks",
			args: args{
				stackNames: []string{"stack-a", "stack-b", "stack-c"},
			},
			want: &StackDependencyGraph{
				dependencies: make(map[string]map[string]struct{}),
				allStacks: map[string]struct{}{
					"stack-a": {},
					"stack-b": {},
					"stack-c": {},
				},
			},
		},
		{
			name: "create graph with single stack",
			args: args{
				stackNames: []string{"stack-a"},
			},
			want: &StackDependencyGraph{
				dependencies: make(map[string]map[string]struct{}),
				allStacks: map[string]struct{}{
					"stack-a": {},
				},
			},
		},
		{
			name: "create graph with empty stacks",
			args: args{
				stackNames: []string{},
			},
			want: &StackDependencyGraph{
				dependencies: make(map[string]map[string]struct{}),
				allStacks:    map[string]struct{}{},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStackDependencyGraph(tt.args.stackNames)

			if !reflect.DeepEqual(got.allStacks, tt.want.allStacks) {
				t.Errorf("NewStackDependencyGraph() allStacks = %v, want %v", got.allStacks, tt.want.allStacks)
			}
			if !reflect.DeepEqual(got.dependencies, tt.want.dependencies) {
				t.Errorf("NewStackDependencyGraph() dependencies = %v, want %v", got.dependencies, tt.want.dependencies)
			}
		})
	}
}

func TestStackDependencyGraph_AddDependency(t *testing.T) {
	type args struct {
		fromStack string
		toStack   string
	}

	cases := []struct {
		name         string
		initialGraph *StackDependencyGraph
		args         args
		want         map[string]map[string]struct{}
	}{
		{
			name:         "add single dependency",
			initialGraph: NewStackDependencyGraph([]string{"stack-a", "stack-b"}),
			args: args{
				fromStack: "stack-b",
				toStack:   "stack-a",
			},
			want: map[string]map[string]struct{}{
				"stack-b": {
					"stack-a": {},
				},
			},
		},
		{
			name: "add multiple dependencies from same stack",
			initialGraph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				g.AddDependency("stack-c", "stack-a")
				return g
			}(),
			args: args{
				fromStack: "stack-c",
				toStack:   "stack-b",
			},
			want: map[string]map[string]struct{}{
				"stack-c": {
					"stack-a": {},
					"stack-b": {},
				},
			},
		},
		{
			name: "add duplicate dependency (should deduplicate)",
			initialGraph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b"})
				g.AddDependency("stack-b", "stack-a")
				return g
			}(),
			args: args{
				fromStack: "stack-b",
				toStack:   "stack-a",
			},
			want: map[string]map[string]struct{}{
				"stack-b": {
					"stack-a": {},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tt.initialGraph.AddDependency(tt.args.fromStack, tt.args.toStack)

			if !reflect.DeepEqual(tt.initialGraph.dependencies, tt.want) {
				t.Errorf("AddDependency() = %v, want %v", tt.initialGraph.dependencies, tt.want)
			}
		})
	}
}

func TestStackDependencyGraph_DetectCircularDependency(t *testing.T) {
	cases := []struct {
		name      string
		graph     *StackDependencyGraph
		wantCycle []string
		wantErr   bool
	}{
		{
			name: "no circular dependency",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				g.AddDependency("stack-b", "stack-a")
				g.AddDependency("stack-c", "stack-b")
				return g
			}(),
			wantCycle: nil,
			wantErr:   false,
		},
		{
			name: "circular dependency with 2 stacks",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b"})
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-a")
				return g
			}(),
			wantCycle: []string{"stack-a", "stack-b", "stack-a"},
			wantErr:   true,
		},
		{
			name: "circular dependency with 3 stacks",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-c")
				g.AddDependency("stack-c", "stack-a")
				return g
			}(),
			wantCycle: []string{"stack-a", "stack-b", "stack-c", "stack-a"},
			wantErr:   true,
		},
		{
			name: "no dependency at all",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				return g
			}(),
			wantCycle: nil,
			wantErr:   false,
		},
		{
			name: "diamond dependency (no cycle)",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c", "stack-d"})
				g.AddDependency("stack-b", "stack-a")
				g.AddDependency("stack-c", "stack-a")
				g.AddDependency("stack-d", "stack-b")
				g.AddDependency("stack-d", "stack-c")
				return g
			}(),
			wantCycle: nil,
			wantErr:   false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			gotCycle, gotErr := tt.graph.DetectCircularDependency()

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("DetectCircularDependency() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if tt.wantErr {
				// Verify cycle is valid (first and last elements should be same, length should match)
				if len(gotCycle) != len(tt.wantCycle) {
					t.Errorf("DetectCircularDependency() cycle length = %v, want %v", len(gotCycle), len(tt.wantCycle))
					return
				}
				if len(gotCycle) > 0 && gotCycle[0] != gotCycle[len(gotCycle)-1] {
					t.Errorf("DetectCircularDependency() cycle does not close: %v", gotCycle)
				}
			}
		})
	}
}
