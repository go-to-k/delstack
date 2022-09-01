package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
)

var _ Operator = (*RoleOperator)(nil)

type RoleOperator struct {
	client    *client.IAM
	resources []*types.StackResourceSummary
}

func NewRoleOperator(config aws.Config) *RoleOperator {
	client := client.NewIAM(config)
	return &RoleOperator{
		client:    client,
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
	for _, role := range operator.resources {
		err := operator.DeleteRole(role.PhysicalResourceId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (operator *RoleOperator) DeleteRole(roleName *string) error {
	policies, err := operator.client.ListAttachedRolePolicies(roleName)
	if err != nil {
		return err
	}

	if err := operator.client.DetachRolePolicies(roleName, policies); err != nil {
		return err
	}

	if err := operator.client.DeleteRole(roleName); err != nil {
		return err
	}

	return nil
}
