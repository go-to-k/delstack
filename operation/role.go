package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ Operator = (*RoleOperator)(nil)

type RoleOperator struct {
	client    client.IIam
	resources []*types.StackResourceSummary
}

func NewRoleOperator(client client.IIam) *RoleOperator {
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
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(option.ConcurrencyNum))

	for _, role := range operator.resources {
		role := role
		eg.Go(func() error {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			return operator.DeleteRole(role.PhysicalResourceId)
		})
	}

	return eg.Wait()
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
