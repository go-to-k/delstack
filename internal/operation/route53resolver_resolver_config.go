package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	resolvertypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*Route53ResolverConfigOperator)(nil)

type Route53ResolverConfigOperator struct {
	client    client.IRoute53Resolver
	resources []*types.StackResourceSummary
}

func NewRoute53ResolverConfigOperator(client client.IRoute53Resolver) *Route53ResolverConfigOperator {
	return &Route53ResolverConfigOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *Route53ResolverConfigOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *Route53ResolverConfigOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *Route53ResolverConfigOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			// The physical resource ID is the VPC (or Route 53 Profile) ID.
			return o.DeleteResolverConfig(ctx, resource.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

// DeleteResolverConfig mirrors the CloudFormation delete handler: it resets AutodefinedReverse to ENABLE.
// If the VPC is already gone, the resolver config is gone with it and there is nothing to do.
func (o *Route53ResolverConfigOperator) DeleteResolverConfig(ctx context.Context, resourceId *string) error {
	config, exists, err := o.client.GetResolverConfig(ctx, resourceId)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	switch config.AutodefinedReverse {
	case resolvertypes.ResolverAutodefinedReverseStatusEnabled, resolvertypes.ResolverAutodefinedReverseStatusEnabling:
		return nil
	}

	return o.client.UpdateResolverConfig(ctx, resourceId, resolvertypes.AutodefinedReverseFlagEnable)
}
