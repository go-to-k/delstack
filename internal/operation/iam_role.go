package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*IamRoleOperator)(nil)

type IamRoleOperator struct {
	client    client.IIam
	resources []*types.StackResourceSummary
}

func NewIamRoleOperator(client client.IIam) *IamRoleOperator {
	return &IamRoleOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *IamRoleOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *IamRoleOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *IamRoleOperator) DeleteResources(ctx context.Context) error {
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

func (o *IamRoleOperator) DeleteRole(ctx context.Context, roleName *string) error {
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
		if err := o.client.DetachRolePolicies(ctx, roleName, policies); err != nil {
			return err
		}
	}

	if err := o.client.DeleteRole(ctx, roleName); err != nil {
		return err
	}

	return nil
}
