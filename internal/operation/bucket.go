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

const sleepTimeSecForS3 = 10

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

func (operator *BucketOperator) AddResource(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *BucketOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *BucketOperator) DeleteResources(ctx context.Context) error {
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, bucket := range operator.resources {
		bucket := bucket
		sem.Acquire(ctx, 1)
		eg.Go(func() error {
			defer sem.Release(1)

			return operator.DeleteBucket(ctx, bucket.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (operator *BucketOperator) DeleteBucket(ctx context.Context, bucketName *string) error {
	exists, err := operator.client.CheckBucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	versions, err := operator.client.ListObjectVersions(ctx, bucketName)
	if err != nil {
		return err
	}

	if len(versions) > 0 {
		errors, err := operator.client.DeleteObjects(ctx, bucketName, versions, sleepTimeSecForS3)
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
	if err := operator.client.DeleteBucket(ctx, bucketName); err != nil {
		return err
	}

	return nil
}
