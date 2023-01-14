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

func (o *RoleOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *RoleOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *RoleOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, role := range o.resources {
		role := role
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteRole(ctx, role.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *RoleOperator) DeleteRole(ctx context.Context, roleName *string) error {
	exists, err := o.client.CheckRoleExists(ctx, roleName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	policies, err := o.client.ListAttachedRolePolicies(ctx, roleName)
	if err != nil {
		return err
	}

	if len(policies) > 0 {
		if err := o.client.DetachRolePolicies(ctx, roleName, policies, sleepTimeSecForIam); err != nil {
			return err
		}
	}

	if err := o.client.DeleteRole(ctx, roleName, sleepTimeSecForIam); err != nil {
		return err
	}

	return nil
}
