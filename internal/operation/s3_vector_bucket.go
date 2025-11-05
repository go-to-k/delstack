package operation

import (
	"context"
	"runtime"

	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Too Many Requests error often occurs, so limit the value
const SemaphoreWeightForS3Vectors = 8

var _ IOperator = (*S3VectorBucketOperator)(nil)

type S3VectorBucketOperator struct {
	client    client.IS3Vectors
	resources []*cfntypes.StackResourceSummary
}

func NewS3VectorBucketOperator(client client.IS3Vectors) *S3VectorBucketOperator {
	return &S3VectorBucketOperator{
		client:    client,
		resources: []*cfntypes.StackResourceSummary{},
	}
}

func (o *S3VectorBucketOperator) AddResource(resource *cfntypes.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *S3VectorBucketOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *S3VectorBucketOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, bucket := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteS3VectorBucket(ctx, bucket.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *S3VectorBucketOperator) DeleteS3VectorBucket(ctx context.Context, vectorBucketName *string) error {
	exists, err := o.client.CheckVectorBucketExists(ctx, vectorBucketName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	eg := errgroup.Group{}
	sem := semaphore.NewWeighted(SemaphoreWeightForS3Vectors)
	var nextToken *string
	for {
		select {
		case <-ctx.Done():
			return &client.ClientError{
				ResourceName: vectorBucketName,
				Err:          ctx.Err(),
			}
		default:
		}

		output, err := o.client.ListIndexesByPage(
			ctx,
			vectorBucketName,
			nextToken,
			nil,
		)
		if err != nil {
			return err
		}
		if len(output.Indexes) == 0 {
			break
		}

		for _, index := range output.Indexes {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			eg.Go(func() error {
				defer sem.Release(1)
				return o.client.DeleteIndex(ctx, index.IndexName, vectorBucketName)
			})
		}

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := o.client.DeleteVectorBucket(ctx, vectorBucketName); err != nil {
		return err
	}

	return nil
}
