package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*AthenaWorkGroupOperator)(nil)

type AthenaWorkGroupOperator struct {
	client    client.IAthena
	resources []*types.StackResourceSummary
}

func NewAthenaWorkGroupOperator(client client.IAthena) *AthenaWorkGroupOperator {
	return &AthenaWorkGroupOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *AthenaWorkGroupOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *AthenaWorkGroupOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *AthenaWorkGroupOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		resource := resource
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteAthenaWorkGroup(ctx, resource.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (o *AthenaWorkGroupOperator) DeleteAthenaWorkGroup(ctx context.Context, workGroupName *string) error {
	exists, err := o.client.CheckAthenaWorkGroupExists(ctx, workGroupName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return o.client.DeleteWorkGroup(ctx, workGroupName)
}
