package operation

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	lambdaEdgeRetryInterval = 30 * time.Second

	// lambdaVPCENIDescriptionPrefix is the description prefix AWS Lambda assigns to
	// VPC (Hyperplane) ENIs it provisions for VPC-attached functions. AWS Lambda
	// releases these ENIs asynchronously after function deletion, which can leave
	// them orphan in `available` state and block Subnet/SecurityGroup deletion.
	lambdaVPCENIDescriptionPrefix = "AWS Lambda VPC ENI"
)

// LambdaFunctionOperator handles deletion of Lambda functions that fail to delete due to
// Lambda@Edge replicas. When a CloudFront distribution uses Lambda@Edge, CloudFormation
// deletes the distribution first, but the Lambda function deletion fails because edge
// replicas are still being cleaned up asynchronously by AWS.
//
// This is implemented as an Operator (not a Preprocessor) because pre-detaching
// Lambda@Edge associations from CloudFront before stack deletion does not reduce the
// total wait time: the replica cleanup duration is the same regardless of when the
// detachment happens. The Operator approach retries deletion within the existing
// DELETE_FAILED retry loop, which is simpler and equally effective.
var _ IOperator = (*LambdaFunctionOperator)(nil)

type LambdaFunctionOperator struct {
	client    client.ILambda
	resources []*types.StackResourceSummary
	// retryInterval is stored as a field (rather than using the constant directly)
	// so that tests can override it to avoid long waits.
	retryInterval time.Duration
}

func NewLambdaFunctionOperator(client client.ILambda) *LambdaFunctionOperator {
	return &LambdaFunctionOperator{
		client:        client,
		resources:     []*types.StackResourceSummary{},
		retryInterval: lambdaEdgeRetryInterval,
	}
}

func (o *LambdaFunctionOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *LambdaFunctionOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *LambdaFunctionOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteLambdaFunction(ctx, resource.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (o *LambdaFunctionOperator) DeleteLambdaFunction(ctx context.Context, functionName *string) error {
	exists, err := o.client.CheckLambdaFunctionExists(ctx, functionName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = o.client.DeleteFunction(ctx, functionName)
	if err == nil {
		return nil
	}

	if !o.isReplicatedFunctionError(err) {
		return err
	}

	io.Logger.Info().Msgf("Lambda@Edge function %s has replicas that are still being cleaned up. Waiting for AWS to finish removing edge replicas.", *functionName)

	for {
		select {
		case <-ctx.Done():
			return &client.ClientError{
				ResourceName: functionName,
				Err:          ctx.Err(),
			}
		case <-time.After(o.retryInterval):
		}

		err = o.client.DeleteFunction(ctx, functionName)
		if err == nil {
			io.Logger.Info().Msgf("Lambda@Edge function %s deleted successfully.", *functionName)
			return nil
		}

		if !o.isReplicatedFunctionError(err) {
			return err
		}
	}
}

func (o *LambdaFunctionOperator) isReplicatedFunctionError(err error) bool {
	return strings.Contains(err.Error(), "replicated function")
}

// cleanupOrphanLambdaENIsByFilter finds available ENIs whose description starts with
// "AWS Lambda VPC ENI" and that match the given filter (e.g. subnet-id or group-id),
// then deletes them in parallel. Used by EC2SubnetOperator / EC2SecurityGroupOperator
// to unblock deletion when AWS Lambda has not yet released its VPC ENIs after the
// function was already deleted. The helper lives here because it encodes Lambda VPC
// ENI domain knowledge (description prefix, asynchronous release semantics), even
// though it operates on EC2 APIs.
func cleanupOrphanLambdaENIsByFilter(ctx context.Context, ec2Client client.IEC2, filterName, filterValue string) error {
	filters := []ec2types.Filter{
		{
			Name:   aws.String(filterName),
			Values: []string{filterValue},
		},
		{
			Name:   aws.String("description"),
			Values: []string{lambdaVPCENIDescriptionPrefix + "*"},
		},
		{
			Name:   aws.String("status"),
			Values: []string{string(ec2types.NetworkInterfaceStatusAvailable)},
		},
	}

	enis, err := ec2Client.DescribeNetworkInterfaces(ctx, filters)
	if err != nil {
		return err
	}
	if len(enis) == 0 {
		return nil
	}

	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	for _, eni := range enis {
		eniId := eni.NetworkInterfaceId
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)
			return ec2Client.DeleteNetworkInterface(ctx, eniId)
		})
	}

	return eg.Wait()
}
