package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*RoleOperator)(nil)

const sleepTimeSecForIam = 5

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

func (operator *RoleOperator) AddResource(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *RoleOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *RoleOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, role := range operator.resources {
		role := role
		sem.Acquire(ctx, 1)
		eg.Go(func() error {
			defer sem.Release(1)

			return operator.DeleteRole(ctx, role.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (operator *RoleOperator) DeleteRole(ctx context.Context, roleName *string) error {
	exists, err := operator.client.CheckRoleExists(ctx, roleName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	policies, err := operator.client.ListAttachedRolePolicies(ctx, roleName)
	if err != nil {
		return err
	}

	if len(policies) > 0 {
		if err := operator.client.DetachRolePolicies(ctx, roleName, policies, sleepTimeSecForIam); err != nil {
			return err
		}
	}

	if err := operator.client.DeleteRole(ctx, roleName, sleepTimeSecForIam); err != nil {
		return err
	}

	return nil
}
