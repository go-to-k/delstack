package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// EC2SubnetOperator force-deletes Subnets that are stuck in DELETE_FAILED because
// AWS Lambda has not yet released the VPC ENIs it provisioned for a now-deleted
// VPC-attached function. The Lambda function itself is already gone, but its ENIs
// remain in "available" state and block the subnet deletion. This operator first
// removes those orphan Lambda ENIs, then deletes the subnet itself.
var _ IOperator = (*EC2SubnetOperator)(nil)

type EC2SubnetOperator struct {
	client    client.IEC2
	resources []*types.StackResourceSummary
}

func NewEC2SubnetOperator(client client.IEC2) *EC2SubnetOperator {
	return &EC2SubnetOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *EC2SubnetOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *EC2SubnetOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *EC2SubnetOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteEC2Subnet(ctx, resource.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *EC2SubnetOperator) DeleteEC2Subnet(ctx context.Context, subnetId *string) error {
	if err := o.client.DeleteOrphanLambdaENIsByFilter(ctx, "subnet-id", aws.ToString(subnetId)); err != nil {
		return err
	}

	return o.client.DeleteSubnet(ctx, subnetId)
}
