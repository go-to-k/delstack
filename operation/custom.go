package operation

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var _ Operator = (*CustomOperator)(nil)

type CustomOperator struct {
	resources []*types.StackResourceSummary
}

func NewCustomOperator() *CustomOperator {
	return &CustomOperator{
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *CustomOperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *CustomOperator) GetResourcesLength() int {
	return len(operator.resources)
}

// Implicit implements (these resources will be deleted on its own)
func (operator *CustomOperator) DeleteResources() error {
	return nil
}
