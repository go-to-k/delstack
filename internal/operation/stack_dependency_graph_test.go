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
		name       string
		graph      *StackDependencyGraph
		wantCycles [][]string
		hasCycle   bool
	}{
		{
			name: "no circular dependency",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				g.AddDependency("stack-b", "stack-a")
				g.AddDependency("stack-c", "stack-b")
				return g
			}(),
			wantCycles: nil,
			hasCycle:   false,
		},
		{
			name: "circular dependency with 2 stacks",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b"})
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-a")
				return g
			}(),
			wantCycles: [][]string{
				{"stack-a", "stack-b", "stack-a"},
			},
			hasCycle: true,
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
			wantCycles: [][]string{
				{"stack-a", "stack-b", "stack-c", "stack-a"},
			},
			hasCycle: true,
		},
		{
			name: "no dependency at all",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				return g
			}(),
			wantCycles: nil,
			hasCycle:   false,
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
			wantCycles: nil,
			hasCycle:   false,
		},
		{
			name: "multiple independent circular dependencies",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c", "stack-d"})
				// First cycle: A -> B -> A
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-a")
				// Second cycle: C -> D -> C
				g.AddDependency("stack-c", "stack-d")
				g.AddDependency("stack-d", "stack-c")
				return g
			}(),
			wantCycles: [][]string{
				{"stack-a", "stack-b", "stack-a"},
				{"stack-c", "stack-d", "stack-c"},
			},
			hasCycle: true,
		},
		{
			name: "complex graph with multiple cycles",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c", "stack-d", "stack-e"})
				// First cycle: A -> B -> C -> A
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-c")
				g.AddDependency("stack-c", "stack-a")
				// Second cycle: D -> E -> D
				g.AddDependency("stack-d", "stack-e")
				g.AddDependency("stack-e", "stack-d")
				return g
			}(),
			wantCycles: [][]string{
				{"stack-a", "stack-b", "stack-c", "stack-a"},
				{"stack-d", "stack-e", "stack-d"},
			},
			hasCycle: true,
		},
		{
			name: "non-cycle node depending on cycle member (A->B->A, C->B)",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				// Cycle: A -> B -> A
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-a")
				// Non-cycle node C depends on cycle member B
				g.AddDependency("stack-c", "stack-b")
				return g
			}(),
			wantCycles: [][]string{
				{"stack-a", "stack-b", "stack-a"},
			},
			hasCycle: true,
		},
		{
			name: "non-cycle node depending on cycle member (A->B->A, C->A)",
			graph: func() *StackDependencyGraph {
				g := NewStackDependencyGraph([]string{"stack-a", "stack-b", "stack-c"})
				// Cycle: A -> B -> A
				g.AddDependency("stack-a", "stack-b")
				g.AddDependency("stack-b", "stack-a")
				// Non-cycle node C depends on cycle member A
				g.AddDependency("stack-c", "stack-a")
				return g
			}(),
			wantCycles: [][]string{
				{"stack-a", "stack-b", "stack-a"},
			},
			hasCycle: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			gotCycles := tt.graph.DetectCircularDependency()

			if (len(gotCycles) > 0) != tt.hasCycle {
				t.Errorf("DetectCircularDependency() hasCycle = %v, want %v", len(gotCycles) > 0, tt.hasCycle)
				return
			}

			if tt.hasCycle {
				// Verify number of cycles
				if len(gotCycles) != len(tt.wantCycles) {
					t.Errorf("DetectCircularDependency() number of cycles = %v, want %v", len(gotCycles), len(tt.wantCycles))
					return
				}

				// Verify each cycle is valid (first and last elements should be same)
				for i, cycle := range gotCycles {
					if len(cycle) == 0 {
						t.Errorf("DetectCircularDependency() cycle[%d] is empty", i)
						continue
					}
					if cycle[0] != cycle[len(cycle)-1] {
						t.Errorf("DetectCircularDependency() cycle[%d] does not close: %v", i, cycle)
					}
				}
			}
		})
	}
}
