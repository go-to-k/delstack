package operation

import (
	"fmt"
	"sort"
	"strings"
)

/*
stack_dependency_graph.go - CloudFormation stack dependency graph and concurrent deletion logic

This file provides logic to analyze Output/Import dependencies between CloudFormation stacks
and determine the concurrent deletion order considering dependencies.

## Basic Concepts

### Dependency Definition
- Stack A has an Output (ExportName)
- Stack B has an Import (ImportValue) that references Stack A's Export
- In this case, we define "Stack B depends on Stack A"
- Deletion order: Stack A must be deleted before Stack B

### Graph Structure
- dependencies[B][A] = struct{}{} means "B depends on A"
- This represents the dependency direction: A → B
- Deletion proceeds from the dependency source (A) to the dependent (B)

## Deletion Logic

### Grouping by Topological Sort
Stacks are divided into "concurrently deletable groups".
Stacks within each group can be deleted concurrently, while groups are deleted serially.

### Algorithm (Topological Sort Based on In-Degree)
1. Calculate the in-degree (number of dependencies) for each stack
2. Add stacks with in-degree 0 (no dependencies) to the current group
3. Mark stacks in the group as "deleted"
4. Decrease the in-degree of other stacks that depended on the deleted stacks
5. Repeat steps 2-4 until all stacks are processed

## Examples

### Example 1: Linear Dependency (Completely Serial Deletion)
```
Stack Configuration:
  A (Export: ExportA)
  B (Import: ExportA, Export: ExportB)
  C (Import: ExportB)

Dependencies:
  B → A (B depends on A)
  C → B (C depends on B)

Graph Representation:
  dependencies["B"]["A"] = struct{}{}
  dependencies["C"]["B"] = struct{}{}

Deletion Groups:
  Group 0: [A]
  Group 1: [B]
  Group 2: [C]

Execution:
  1. Delete A (wait for completion)
  2. Delete B (wait for completion)
  3. Delete C (wait for completion)
```

### Example 2: Diamond Dependency (Mixed Concurrent and Serial)
```
Stack Configuration:
  A (Export: ExportA)
  B (Import: ExportA, Export: ExportB)
  C (Import: ExportA, Export: ExportC)
  D (Import: ExportB, ExportC)

Dependencies:
  B → A (B depends on A)
  C → A (C depends on A)
  D → B (D depends on B)
  D → C (D depends on C)

Graph Representation:
  dependencies["B"]["A"] = struct{}{}
  dependencies["C"]["A"] = struct{}{}
  dependencies["D"]["B"] = struct{}{}
  dependencies["D"]["C"] = struct{}{}

In-Degree Calculation:
  A: 0 (no dependencies)
  B: 1 (depends on A)
  C: 1 (depends on A)
  D: 2 (depends on B and C)

Deletion Groups:
  Group 0: [A]           (in-degree 0)
  Group 1: [B, C]        (in-degree becomes 0 after A is deleted)
  Group 2: [D]           (in-degree becomes 0 after B and C are deleted)

Execution:
  1. Delete A (wait for completion)
  2. Delete B and C concurrently (wait for both)
  3. Delete D (wait for completion)
```

### Example 3: Complex Dependencies (Multiple Levels of Parallelism)
```
Stack Configuration:
  A (Export: ExportA)
  B (Export: ExportB)
  C (Import: ExportA, Export: ExportC)
  D (Import: ExportA, Export: ExportD)
  E (Import: ExportB, Export: ExportE)
  F (Import: ExportC, ExportD, ExportE)

Dependencies:
  C → A
  D → A
  E → B
  F → C
  F → D
  F → E

Graph Representation:
  dependencies["C"]["A"] = struct{}{}
  dependencies["D"]["A"] = struct{}{}
  dependencies["E"]["B"] = struct{}{}
  dependencies["F"]["C"] = struct{}{}
  dependencies["F"]["D"] = struct{}{}
  dependencies["F"]["E"] = struct{}{}

In-Degree Calculation:
  A: 0
  B: 0
  C: 1 (A)
  D: 1 (A)
  E: 1 (B)
  F: 3 (C, D, E)

Deletion Groups:
  Group 0: [A, B]              (in-degree 0, concurrent deletion)
  Group 1: [C, D, E]           (concurrent deletion after A, B)
  Group 2: [F]                 (deletion after C, D, E)

Execution:
  1. Delete A and B concurrently (wait for both)
  2. Delete C, D, E concurrently (wait for all three)
  3. Delete F (wait for completion)
```

### Example 4: Independent Stacks
```
Stack Configuration:
  A (no dependencies)
  B (no dependencies)
  C (no dependencies)
  D (no dependencies)

Dependencies:
  None

Graph Representation:
  dependencies = {} (empty)

In-Degree Calculation:
  A: 0
  B: 0
  C: 0
  D: 0

Deletion Groups:
  Group 0: [A, B, C, D]

Execution:
  1. Delete A, B, C, D concurrently (wait for all four)

With concurrency limit (-n 2):
  1. Delete A and B concurrently (wait for completion)
  2. Delete C and D concurrently (wait for completion)
```

### Example 5: Multiple Outputs Referenced (Deduplication)
```
Stack Configuration:
  A (Export: ExportA1, ExportA2, ExportA3)
  B (Import: ExportA1, ExportA2, ExportA3)

Dependencies:
  B → A (depends on all three Exports)

Graph Representation:
  dependencies["B"]["A"] = struct{}{}  (only one entry, deduplicated)

Deletion Groups:
  Group 0: [A]
  Group 1: [B]

Execution:
  1. Delete A (wait for completion)
  2. Delete B (wait for completion)

Note:
  Even if AddDependency("B", "A") is called multiple times,
  the map[string]struct{} structure automatically deduplicates.
```

## Error Cases

### Circular Dependency Detection
```
Stack Configuration:
  A (Export: ExportA, Import: ExportC)
  B (Export: ExportB, Import: ExportA)
  C (Export: ExportC, Import: ExportB)

Dependencies:
  A → C
  B → A
  C → B

This is a circular dependency: A → C → B → A

DetectCircularDependency() Result:
  cyclePath: ["A", "C", "B", "A"]
  error: "CircularDependencyError: A -> C -> B -> A"

Behavior:
  Deletion is not executed, exits with error message
```

### References from Non-Target Stacks (Ignored)
```
Stack Configuration:
  X (Export: ExportX)  ← Not a deletion target
  A (Import: ExportX, Export: ExportA)
  B (Import: ExportA)

Deletion Target: A, B only

Dependency Graph (Only Among Target Stacks):
  B → A

Graph Representation:
  dependencies["B"]["A"] = struct{}{}

Deletion Groups:
  Group 0: [A]
  Group 1: [B]

Execution:
  1. Delete A (wait for completion)
  2. Delete B (wait for completion)
```

## Concurrency Limit

### -n Option Behavior
Concurrency limit is applied only within each group.

```
Deletion Groups:
  Group 0: [A, B, C, D, E]  (5 stacks)
  Group 1: [F, G]           (2 stacks)

With -n 2:
  Group 0:
    1. Delete A, B concurrently
    2. Delete C, D concurrently
    3. Delete E
  Group 1:
    4. Delete F, G concurrently

Without -n:
  Group 0:
    1. Delete A, B, C, D, E concurrently (5 at once)
  Group 1:
    2. Delete F, G concurrently (2 at once)
```

## Performance Characteristics

- Time Complexity: O(V + E) (V: number of stacks, E: number of dependencies)
- Space Complexity: O(V + E)
- Topological Sort uses Kahn's Algorithm (in-degree based)
- Circular dependency detection uses DFS (Depth-First Search)
*/

// StackDependencyGraph represents the dependency graph between stacks
type StackDependencyGraph struct {
	dependencies map[string]map[string]struct{}
	allStacks    map[string]struct{}
}

// NewStackDependencyGraph creates a new dependency graph
func NewStackDependencyGraph(stackNames []string) *StackDependencyGraph {
	allStacks := make(map[string]struct{})
	for _, name := range stackNames {
		allStacks[name] = struct{}{}
	}

	return &StackDependencyGraph{
		dependencies: make(map[string]map[string]struct{}),
		allStacks:    allStacks,
	}
}

// AddDependency adds a dependency relationship (with automatic deduplication)
// fromStack depends on toStack (fromStack → toStack)
// This means toStack must be deleted before fromStack
func (g *StackDependencyGraph) AddDependency(fromStack, toStack string) {
	if _, exists := g.dependencies[fromStack]; !exists {
		g.dependencies[fromStack] = make(map[string]struct{})
	}
	g.dependencies[fromStack][toStack] = struct{}{}
}

// DetectCircularDependency detects circular dependencies using DFS
func (g *StackDependencyGraph) DetectCircularDependency() ([]string, error) {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	var cyclePath []string

	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recursionStack[node] = true
		cyclePath = append(cyclePath, node)

		for dep := range g.dependencies[node] {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if recursionStack[dep] {
				for i, n := range cyclePath {
					if n == dep {
						cyclePath = cyclePath[i:]
						cyclePath = append(cyclePath, dep)
						return true
					}
				}
			}
		}

		recursionStack[node] = false
		cyclePath = cyclePath[:len(cyclePath)-1]
		return false
	}

	for stack := range g.allStacks {
		if visited[stack] {
			continue
		}

		cyclePath = []string{}
		if dfs(stack) {
			return cyclePath, fmt.Errorf("CircularDependencyError: %v", strings.Join(cyclePath, " -> "))
		}
	}

	return nil, nil
}

// GetDeletionGroups groups stacks that can be deleted concurrently using topological sort
// Returns: outer array is deletion order, inner array is stacks that can be deleted concurrently
func (g *StackDependencyGraph) GetDeletionGroups() [][]string {
	inDegree := make(map[string]int)
	for stack := range g.allStacks {
		inDegree[stack] = len(g.dependencies[stack])
	}

	var groups [][]string
	processed := make(map[string]bool)

	for len(processed) < len(g.allStacks) {
		currentGroup := []string{}
		for stack := range g.allStacks {
			if processed[stack] {
				continue
			}
			if inDegree[stack] == 0 {
				currentGroup = append(currentGroup, stack)
			}
		}

		if len(currentGroup) == 0 {
			break
		}

		sort.Strings(currentGroup)
		groups = append(groups, currentGroup)

		for _, stack := range currentGroup {
			processed[stack] = true

			for depStack, deps := range g.dependencies {
				if _, dependsOnCurrent := deps[stack]; dependsOnCurrent {
					inDegree[depStack]--
				}
			}
		}
	}

	return groups
}
