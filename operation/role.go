package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var _ IOperator = (*RoleOperator)(nil)

type RoleOperator struct {
	// client    *client.IAM
	resources []*types.StackResourceSummary
}

func NewRoleOperator(config aws.Config) *RoleOperator {
	// client := client.NewIAM(config)
	return &RoleOperator{
		// client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *RoleOperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *RoleOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *RoleOperator) DeleteResources() error {
	// TODO: Concurrency Delete
	return nil
}
