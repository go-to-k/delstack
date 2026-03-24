package operation

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	lambdaEdgeRetryInterval = 30 * time.Second
	lambdaEdgeRetryTimeout  = 60 * time.Minute
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
	// retryInterval and retryTimeout are stored as fields (rather than using constants
	// directly) so that tests can override them to avoid waiting 60 minutes.
	retryInterval time.Duration
	retryTimeout  time.Duration
}

func NewLambdaFunctionOperator(client client.ILambda) *LambdaFunctionOperator {
	return &LambdaFunctionOperator{
		client:        client,
		resources:     []*types.StackResourceSummary{},
		retryInterval: lambdaEdgeRetryInterval,
		retryTimeout:  lambdaEdgeRetryTimeout,
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

	io.Logger.Info().Msgf("Lambda@Edge function %s has replicas that are still being cleaned up. This may take up to 60 minutes.", *functionName)

	deadline := time.Now().Add(o.retryTimeout)
	for time.Now().Before(deadline) {
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

	return &client.ClientError{
		ResourceName: functionName,
		Err:          fmt.Errorf("timed out waiting for Lambda@Edge replicas to be cleaned up after %v", o.retryTimeout),
	}
}

func (o *LambdaFunctionOperator) isReplicatedFunctionError(err error) bool {
	return strings.Contains(err.Error(), "replicated function")
}
