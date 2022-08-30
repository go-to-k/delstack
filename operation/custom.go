package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var _ IOperator = (*CustomOperator)(nil)

type CustomOperator struct {
	// client    *client.Custom
	resources []*types.StackResourceSummary
}

func NewCustomOperator(config aws.Config) *CustomOperator {
	// client := client.NewCustom(config)
	return &CustomOperator{
		// client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *CustomOperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *CustomOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *CustomOperator) DeleteResources() error {
	// TODO: Concurrency Delete
	return nil
}
