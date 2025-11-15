package operation

/*
stack_dependency_graph.go - CloudFormation stack dependency graph and concurrent deletion logic

This file provides logic to analyze Output/Import dependencies between CloudFormation stacks
and determine the concurrent deletion order considering dependencies.

## Basic Concepts

### Dependency Definition
- Stack A has an Output (ExportName)
- Stack B has an Import (ImportValue) that references Stack A's Export
- In this case, we define "Stack B depends on Stack A"
- Deletion order: Stack B must be deleted before Stack A (because B imports A's export)

### Graph Structure
- dependencies[B][A] = struct{}{} means "B depends on A"
- This represents the dependency direction: B → A
- Deletion proceeds from the dependent (B) to the dependency source (A)

## Deletion Logic

### Dynamic Deletion with Topological Sort
Stacks are deleted dynamically based on dependency resolution.
When a stack deletion completes, any stacks that depend only on that stack
become immediately available for deletion, without waiting for other concurrent deletions.

### Algorithm (Dynamic Topological Sort Based on Reverse In-Degree)
1. Calculate the reverse in-degree (number of dependents) for each stack
2. Add stacks with reverse in-degree 0 (no dependents) to the deletion queue
3. Start deletion of all stacks in the queue (up to concurrency limit)
4. When a stack deletion completes:
   - Decrease the reverse in-degree of stacks it depends on
   - Add newly available stacks (reverse in-degree becomes 0) to the queue
   - Start their deletion immediately
5. Repeat step 4 until all stacks are deleted

### Key Benefits of Dynamic Deletion
- **Maximum Parallelism**: Stacks become deletable as soon as their dependencies are resolved
- **No Artificial Grouping**: Unlike group-based deletion, stacks don't wait for their "group" to complete
- **Progress Tracking**: Shows "X/Y stacks deleted" for clear progress indication

## Examples

### Example 1: Linear Dependency (Completely Serial Deletion)
```
Stack Configuration:
  A (Export: ExportA)
  B (Import: ExportA, Export: ExportB)
  C (Import: ExportB)

Dependencies:
  B → A (B depends on A, B imports A's export)
  C → B (C depends on B, C imports B's export)

Graph Representation:
  dependencies["B"]["A"] = struct{}{}
  dependencies["C"]["B"] = struct{}{}

Reverse In-Degree Calculation (how many stacks depend on this stack):
  A: 1 (B depends on A)
  B: 1 (C depends on B)
  C: 0 (no one depends on C)

Dynamic Deletion Flow:
  Initial queue: [C]  (reverse in-degree 0)

  Step 1: Start deleting C
  Step 2: C completes → decrease B's reverse in-degree (1→0) → add B to queue
  Step 3: Start deleting B
  Step 4: B completes → decrease A's reverse in-degree (1→0) → add A to queue
  Step 5: Start deleting A
  Step 6: A completes → all done

Progress Output:
  "Progress: 1/3 stacks deleted" (after C)
  "Progress: 2/3 stacks deleted" (after B)
  "Progress: 3/3 stacks deleted" (after A)
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

Reverse In-Degree Calculation (how many stacks depend on this stack):
  A: 2 (B and C depend on A)
  B: 1 (D depends on B)
  C: 1 (D depends on C)
  D: 0 (no one depends on D)

Dynamic Deletion Flow:
  Initial queue: [D]  (reverse in-degree 0)

  Step 1: Start deleting D
  Step 2: D completes → decrease B's and C's reverse in-degree (both 1→0) → add B, C to queue
  Step 3: Start deleting B and C concurrently
  Step 4: B completes → decrease A's reverse in-degree (2→1)
  Step 5: C completes → decrease A's reverse in-degree (1→0) → add A to queue
  Step 6: Start deleting A
  Step 7: A completes → all done

Progress Output:
  "Progress: 1/4 stacks deleted" (after D)
  "Progress: 2/4 stacks deleted" (after B or C)
  "Progress: 3/4 stacks deleted" (after C or B)
  "Progress: 4/4 stacks deleted" (after A)
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

Reverse In-Degree Calculation (how many stacks depend on this stack):
  A: 2 (C and D depend on A)
  B: 1 (E depends on B)
  C: 1 (F depends on C)
  D: 1 (F depends on D)
  E: 1 (F depends on E)
  F: 0 (no one depends on F)

Dynamic Deletion Flow:
  Initial queue: [F]  (reverse in-degree 0)

  Step 1: Start deleting F
  Step 2: F completes → decrease C's, D's, E's reverse in-degree (all 1→0) → add C, D, E to queue
  Step 3: Start deleting C, D, E concurrently
  Step 4: E completes → decrease B's reverse in-degree (1→0) → add B to queue
  Step 5: Start deleting B (doesn't wait for C, D!)
  Step 6: C completes → decrease A's reverse in-degree (2→1)
  Step 7: D completes → decrease A's reverse in-degree (1→0) → add A to queue
  Step 8: Start deleting A
  Step 9: B and A complete → all done

Progress Output:
  "Progress: 1/6 stacks deleted" (after F)
  "Progress: 2/6 stacks deleted" (after E)
  "Progress: 3/6 stacks deleted" (after C)
  "Progress: 4/6 stacks deleted" (after D)
  "Progress: 5/6 stacks deleted" (after B or A)
  "Progress: 6/6 stacks deleted" (after A or B)

Note: B can start as soon as E completes, without waiting for C and D!
      This is the key advantage of dynamic deletion.
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

Reverse In-Degree Calculation (how many stacks depend on this stack):
  A: 0
  B: 0
  C: 0
  D: 0

Dynamic Deletion Flow:
  Initial queue: [A, B, C, D]  (all have reverse in-degree 0)

  Without concurrency limit:
    Step 1: Start deleting A, B, C, D concurrently
    Step 2-5: All complete in any order

  With concurrency limit (-n 2):
    Step 1: Start deleting A and B (semaphore limit reached)
    Step 2: A completes → start deleting C
    Step 3: B completes → start deleting D
    Step 4: C and D complete

Progress Output:
  "Progress: 1/4 stacks deleted"
  "Progress: 2/4 stacks deleted"
  "Progress: 3/4 stacks deleted"
  "Progress: 4/4 stacks deleted"
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

Reverse In-Degree Calculation (how many stacks depend on this stack):
  A: 1 (B depends on A)
  B: 0 (no one depends on B)

Dynamic Deletion Flow:
  Initial queue: [B]  (reverse in-degree 0)

  Step 1: Start deleting B
  Step 2: B completes → decrease A's reverse in-degree (1→0) → add A to queue
  Step 3: Start deleting A
  Step 4: A completes → all done

Progress Output:
  "Progress: 1/2 stacks deleted" (after B)
  "Progress: 2/2 stacks deleted" (after A)

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

Reverse In-Degree Calculation (how many stacks depend on this stack):
  A: 1 (B depends on A)
  B: 0 (no one depends on B)

Dynamic Deletion Flow:
  Initial queue: [B]  (reverse in-degree 0)

  Step 1: Start deleting B
  Step 2: B completes → decrease A's reverse in-degree (1→0) → add A to queue
  Step 3: Start deleting A
  Step 4: A completes → all done

Progress Output:
  "Progress: 1/2 stacks deleted" (after B)
  "Progress: 2/2 stacks deleted" (after A)
```

## Concurrency Limit

### -n Option Behavior
Concurrency limit is applied globally across all stacks using a semaphore.
The limit controls the maximum number of stacks being deleted at any given time.

```
Example: 5 independent stacks [A, B, C, D, E]

Without -n:
  All 5 stacks start deletion simultaneously

With -n 2:
  Step 1: A and B start (semaphore: 2/2 acquired)
  Step 2: A completes → C starts (semaphore: A releases, C acquires)
  Step 3: B completes → D starts (semaphore: B releases, D acquires)
  Step 4: C completes → E starts (semaphore: C releases, E acquires)
  Step 5: D and E complete

Progress Output:
  "Progress: 1/5 stacks deleted"
  "Progress: 2/5 stacks deleted"
  "Progress: 3/5 stacks deleted"
  "Progress: 4/5 stacks deleted"
  "Progress: 5/5 stacks deleted"

Note: With dynamic deletion and -n limit, stacks are deleted as soon as:
      1. Their dependencies are resolved (reverse in-degree becomes 0)
      2. A semaphore slot is available
```

## Performance Characteristics

- Time Complexity: O(V + E) (V: number of stacks, E: number of dependencies)
- Space Complexity: O(V + E)
- Dynamic deletion uses reverse in-degree based topological sort
- Circular dependency detection uses DFS (Depth-First Search)
- Concurrency: Stacks are deleted as soon as dependencies are resolved (no artificial grouping delays)
*/

// StackDependencyGraph represents the dependency graph between stacks
type StackDependencyGraph struct {
	dependencies map[string]map[string]struct{}
	allStacks    map[string]struct{}
}

// NewStackDependencyGraph creates a new dependency graph
func NewStackDependencyGraph(stackNames []string) *StackDependencyGraph {
	allStacks := make(map[string]struct{}, len(stackNames))
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

// GetDependencies returns the dependencies map
func (g *StackDependencyGraph) GetDependencies() map[string]map[string]struct{} {
	return g.dependencies
}

// GetAllStacks returns all stacks in the graph
func (g *StackDependencyGraph) GetAllStacks() map[string]struct{} {
	return g.allStacks
}

// DetectCircularDependency detects circular dependencies using DFS (Depth-First Search)
// Returns the cycle path if a circular dependency is detected, nil otherwise
func (g *StackDependencyGraph) DetectCircularDependency() []string {
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
			return cyclePath
		}
	}

	return nil
}
