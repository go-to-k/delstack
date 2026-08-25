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
	// bound the wait for connections to leave the blocking states: the rejection is not
	// always reflected immediately, and a connection that is still being provisioned has
	// to settle before it can be rejected at all. Together they allow ~90 seconds, the
	// same budget ENIDetachmentWaitTime gives the other asynchronous EC2 cleanup. This
	// wait is the only chance the operator gets: the CFN delete loop passes every
	// DELETE_FAILED resource as RetainResources on the next DeleteStack, so
	// CloudFormation never retries the endpoint service by itself.
	vpcEndpointConnectionRejectionMaxAttempts = 18
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
	rejectedVpcEndpointIds := map[string]struct{}{}

	for attempt := 0; attempt < vpcEndpointConnectionRejectionMaxAttempts; attempt++ {
		connections, err := o.client.DescribeVpcEndpointConnections(ctx, serviceId)
		if err != nil {
			return err
		}

		if !hasBlockingVpcEndpointConnections(connections) {
			return nil
		}

		vpcEndpointIds := rejectableVpcEndpointIds(connections, rejectedVpcEndpointIds)
		for start := 0; start < len(vpcEndpointIds); start += vpcEndpointConnectionRejectionBatchSize {
			end := min(start+vpcEndpointConnectionRejectionBatchSize, len(vpcEndpointIds))

			if err := o.client.RejectVpcEndpointConnections(ctx, serviceId, vpcEndpointIds[start:end]); err != nil {
				return err
			}
		}
		for _, vpcEndpointId := range vpcEndpointIds {
			rejectedVpcEndpointIds[vpcEndpointId] = struct{}{}
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

// hasBlockingVpcEndpointConnections reports whether any connection still stands in the
// way of the endpoint service deletion. `Available` and `PendingAcceptance` block it
// right now; `Pending` is a connection that has been accepted and is still being
// provisioned, so it turns into `Available` shortly and has to be waited out rather
// than deleted through. Any other state (`Rejected`, `Deleting`, `Deleted`, `Failed`,
// ...) no longer blocks it.
func hasBlockingVpcEndpointConnections(connections []ec2types.VpcEndpointConnection) bool {
	for _, connection := range connections {
		switch connection.VpcEndpointState {
		case ec2types.StateAvailable, ec2types.StatePendingAcceptance, ec2types.StatePending:
			return true
		}
	}

	return false
}

// rejectableVpcEndpointIds returns the endpoints that can be rejected right now, minus
// the ones already rejected. Rejecting the same connection twice is not harmless: AWS
// reports a per-resource error for a connection that is no longer in a rejectable
// state, and DescribeVpcEndpointConnections can still report a just-rejected connection
// as `Available` for a moment.
func rejectableVpcEndpointIds(connections []ec2types.VpcEndpointConnection, rejectedVpcEndpointIds map[string]struct{}) []string {
	vpcEndpointIds := []string{}

	for _, connection := range connections {
		if connection.VpcEndpointState != ec2types.StateAvailable && connection.VpcEndpointState != ec2types.StatePendingAcceptance {
			continue
		}
		if connection.VpcEndpointId == nil {
			continue
		}

		vpcEndpointId := aws.ToString(connection.VpcEndpointId)
		if _, ok := rejectedVpcEndpointIds[vpcEndpointId]; ok {
			continue
		}
		vpcEndpointIds = append(vpcEndpointIds, vpcEndpointId)
	}

	return vpcEndpointIds
}
