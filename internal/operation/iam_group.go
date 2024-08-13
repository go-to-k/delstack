package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*IamGroupOperator)(nil)

type IamGroupOperator struct {
	client    client.IIam
	resources []*types.StackResourceSummary
}

func NewIamGroupOperator(client client.IIam) *IamGroupOperator {
	return &IamGroupOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *IamGroupOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *IamGroupOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *IamGroupOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, Group := range o.resources {
		Group := Group
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteIamGroup(ctx, Group.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *IamGroupOperator) DeleteIamGroup(ctx context.Context, GroupName *string) error {
	exists, err := o.client.CheckGroupExists(ctx, GroupName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	users, err := o.client.GetGroupUsers(ctx, GroupName)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		if err := o.client.RemoveUsersFromGroup(ctx, GroupName, users); err != nil {
			return err
		}
	}

	if err := o.client.DeleteGroup(ctx, GroupName); err != nil {
		return err
	}

	return nil
}
