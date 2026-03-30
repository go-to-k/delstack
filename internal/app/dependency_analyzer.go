package app

import (
	"context"

	"github.com/go-to-k/delstack/internal/operation"
)

// IDependencyAnalyzer builds a dependency graph for CloudFormation stacks.
type IDependencyAnalyzer interface {
	Analyze(ctx context.Context, stackNames []string, operatorFactory *operation.OperatorFactory) (*operation.StackDependencyGraph, error)
}

type DependencyAnalyzer struct{}

func (a *DependencyAnalyzer) Analyze(ctx context.Context, stackNames []string, operatorFactory *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()
	return cloudformationStackOperator.BuildDependencyGraph(ctx, stackNames)
}
