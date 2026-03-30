package app

import (
	"context"

	"github.com/go-to-k/delstack/internal/operation"
)

// DependencyAnalyzer builds a dependency graph for CloudFormation stacks.
type DependencyAnalyzer interface {
	Analyze(ctx context.Context, stackNames []string, operatorFactory *operation.OperatorFactory) (*operation.StackDependencyGraph, error)
}

type DefaultDependencyAnalyzer struct{}

func (a *DefaultDependencyAnalyzer) Analyze(ctx context.Context, stackNames []string, operatorFactory *operation.OperatorFactory) (*operation.StackDependencyGraph, error) {
	cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()
	return cloudformationStackOperator.BuildDependencyGraph(ctx, stackNames)
}
