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

// EC2SecurityGroupOperator force-deletes SecurityGroups that are stuck in
// DELETE_FAILED because AWS Lambda has not yet released the VPC ENIs it provisioned
// for a now-deleted VPC-attached function. The Lambda function itself is already
// gone, but its ENIs remain in "available" state attached to this security group.
// This operator first removes those orphan Lambda ENIs, then deletes the security
// group itself.
var _ IOperator = (*EC2SecurityGroupOperator)(nil)

type EC2SecurityGroupOperator struct {
	client    client.IEC2
	resources []*types.StackResourceSummary
}

func NewEC2SecurityGroupOperator(client client.IEC2) *EC2SecurityGroupOperator {
	return &EC2SecurityGroupOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *EC2SecurityGroupOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *EC2SecurityGroupOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *EC2SecurityGroupOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteEC2SecurityGroup(ctx, resource.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *EC2SecurityGroupOperator) DeleteEC2SecurityGroup(ctx context.Context, securityGroupId *string) error {
	if err := cleanupOrphanLambdaENIsByFilter(ctx, o.client, "group-id", aws.ToString(securityGroupId)); err != nil {
		return err
	}

	return o.client.DeleteSecurityGroup(ctx, securityGroupId)
}
