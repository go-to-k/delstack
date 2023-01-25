package operation

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*BucketOperator)(nil)

type BucketOperator struct {
	client    client.IS3
	resources []*types.StackResourceSummary
}

func NewBucketOperator(client client.IS3) *BucketOperator {
	return &BucketOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *BucketOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *BucketOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *BucketOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, bucket := range o.resources {
		bucket := bucket
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteBucket(ctx, bucket.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *BucketOperator) DeleteBucket(ctx context.Context, bucketName *string) error {
	exists, err := o.client.CheckBucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	versions, err := o.client.ListObjectVersions(ctx, bucketName)
	if err != nil {
		return err
	}

	if len(versions) > 0 {
		errors, err := o.client.DeleteObjects(ctx, bucketName, versions)
		if err != nil {
			return err
		}
		if len(errors) > 0 {
			errorStr := ""
			for _, error := range errors {
				errorStr += fmt.Sprintf("\nCode: %v\n", *error.Code)
				errorStr += fmt.Sprintf("Key: %v\n", *error.Key)
				errorStr += fmt.Sprintf("VersionId: %v\n", *error.VersionId)
				errorStr += fmt.Sprintf("Message: %v\n", *error.Message)
			}
			return fmt.Errorf("DeleteObjectsError: followings %v", errorStr)
		}
	}
	if err := o.client.DeleteBucket(ctx, bucketName); err != nil {
		return err
	}

	return nil
}
