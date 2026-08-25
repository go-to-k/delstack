package operation

import (
	"context"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	// vpcEndpointConnectionRejectionBatchSize caps how many VPC endpoint IDs are sent
	// to a single RejectVpcEndpointConnections call. AWS documents no maximum, and the
	// number of connections an endpoint service can hold is an adjustable quota, so the
	// list is split conservatively.
	vpcEndpointConnectionRejectionBatchSize = 50

	// vpcEndpointConnectionRejectionMaxAttempts and vpcEndpointConnectionRejectionInterval
	// bound the wait for rejected connections to leave the blocking states. The rejection
	// is not always reflected immediately, and DeleteVpcEndpointServiceConfigurations
	// keeps failing until it is.
	vpcEndpointConnectionRejectionMaxAttempts = 6
	vpcEndpointConnectionRejectionInterval    = 5 * time.Second
)

// EC2VPCEndpointServiceOperator force-deletes VPC endpoint services (AWS PrivateLink /
// Gateway Load Balancer endpoint services) that are stuck in DELETE_FAILED because
// consumers still have endpoints connected to them. AWS refuses to delete an endpoint
// service while any connection is in the `Available` or `PendingAcceptance` state, and
// those endpoints live outside the stack, so CloudFormation cannot clear them. This
// operator rejects the blocking connections and then deletes the endpoint service
// itself. The consumer endpoints are left in place: rejecting only severs the
// connection to this service.
var _ IOperator = (*EC2VPCEndpointServiceOperator)(nil)

type EC2VPCEndpointServiceOperator struct {
	client    client.IEC2
	resources []*types.StackResourceSummary
	// rejectionInterval is stored as a field (rather than using the constant directly)
	// so that tests can override it to avoid long waits.
	rejectionInterval time.Duration
}

func NewEC2VPCEndpointServiceOperator(client client.IEC2) *EC2VPCEndpointServiceOperator {
	return &EC2VPCEndpointServiceOperator{
		client:            client,
		resources:         []*types.StackResourceSummary{},
		rejectionInterval: vpcEndpointConnectionRejectionInterval,
	}
}

func (o *EC2VPCEndpointServiceOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *EC2VPCEndpointServiceOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *EC2VPCEndpointServiceOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteEC2VPCEndpointService(ctx, resource.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *EC2VPCEndpointServiceOperator) DeleteEC2VPCEndpointService(ctx context.Context, serviceId *string) error {
	if err := o.rejectVpcEndpointConnections(ctx, serviceId); err != nil {
		return err
	}

	return o.client.DeleteVpcEndpointServiceConfiguration(ctx, serviceId)
}

// rejectVpcEndpointConnections rejects every connection that blocks the endpoint service
// deletion, and waits for the rejections to be reflected. If connections are still
// blocking once the attempts run out, it returns without an error on purpose: the
// following DeleteVpcEndpointServiceConfigurations reports the authoritative reason.
func (o *EC2VPCEndpointServiceOperator) rejectVpcEndpointConnections(ctx context.Context, serviceId *string) error {
	for attempt := 0; attempt < vpcEndpointConnectionRejectionMaxAttempts; attempt++ {
		connections, err := o.client.DescribeVpcEndpointConnections(ctx, serviceId)
		if err != nil {
			return err
		}

		vpcEndpointIds := blockingVpcEndpointIds(connections)
		if len(vpcEndpointIds) == 0 {
			return nil
		}

		for start := 0; start < len(vpcEndpointIds); start += vpcEndpointConnectionRejectionBatchSize {
			end := min(start+vpcEndpointConnectionRejectionBatchSize, len(vpcEndpointIds))

			if err := o.client.RejectVpcEndpointConnections(ctx, serviceId, vpcEndpointIds[start:end]); err != nil {
				return err
			}
		}

		if attempt == vpcEndpointConnectionRejectionMaxAttempts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return &client.ClientError{
				ResourceName: serviceId,
				Err:          ctx.Err(),
			}
		case <-time.After(o.rejectionInterval):
		}
	}

	return nil
}

// blockingVpcEndpointIds returns the endpoints whose connection state prevents the
// endpoint service from being deleted. Connections in any other state (`Rejected`,
// `Deleted`, `Failed`, ...) no longer block it.
func blockingVpcEndpointIds(connections []ec2types.VpcEndpointConnection) []string {
	vpcEndpointIds := []string{}

	for _, connection := range connections {
		if connection.VpcEndpointState != ec2types.StateAvailable && connection.VpcEndpointState != ec2types.StatePendingAcceptance {
			continue
		}
		if connection.VpcEndpointId == nil {
			continue
		}
		vpcEndpointIds = append(vpcEndpointIds, aws.ToString(connection.VpcEndpointId))
	}

	return vpcEndpointIds
}
