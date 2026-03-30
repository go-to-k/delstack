package operation

// NewStackDependencyGraphForTest creates a StackDependencyGraph from a simple dependency map for testing.
// deps maps stackName -> list of stacks it depends on.
func NewStackDependencyGraphForTest(deps map[string][]string) *StackDependencyGraph {
	allStacks := make(map[string]struct{})
	for stack := range deps {
		allStacks[stack] = struct{}{}
	}

	graph := &StackDependencyGraph{
		dependencies: make(map[string]map[string]struct{}),
		allStacks:    allStacks,
	}

	for fromStack, toStacks := range deps {
		for _, toStack := range toStacks {
			graph.AddDependency(fromStack, toStack)
		}
	}

	return graph
}
