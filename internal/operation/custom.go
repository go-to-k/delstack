package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var _ IOperator = (*CustomOperator)(nil)

type CustomOperator struct {
	resources []*types.StackResourceSummary
}

func NewCustomOperator() *CustomOperator {
	return &CustomOperator{
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *CustomOperator) AddResource(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *CustomOperator) GetResourcesLength() int {
	return len(operator.resources)
}

// Implicit implements (these resources will be deleted on its own)
func (operator *CustomOperator) DeleteResources(context.Context) error {
	return nil
}
