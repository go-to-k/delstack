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

func (o *CustomOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *CustomOperator) GetResourcesLength() int {
	return len(o.resources)
}

// Implicit implements (these resources will be deleted on its own)
func (o *CustomOperator) DeleteResources(ctx context.Context) error {
	return nil
}
