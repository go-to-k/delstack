package operation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/app"
	"github.com/go-to-k/delstack/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ Operator = (*BucketOperator)(nil)

type BucketOperator struct {
	client    *client.S3
	resources []*types.StackResourceSummary
}

func NewBucketOperator(config aws.Config) *BucketOperator {
	client := client.NewS3(config)
	return &BucketOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *BucketOperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *BucketOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *BucketOperator) DeleteResources() error {
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(app.ConcurrencyNum))

	for _, bucket := range operator.resources {
		bucket := bucket
		eg.Go(func() error {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			return operator.DeleteBucket(bucket.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (operator *BucketOperator) DeleteBucket(bucketName *string) error {
	versions, err := operator.client.ListObjectVersions(bucketName)
	if err != nil {
		return err
	}

	errors, err := operator.client.DeleteObjects(bucketName, versions)
	if err != nil {
		return err
	}
	if len(errors) > 0 {
		return fmt.Errorf("DeleteObjectsError: %v", errors)
	}

	if err := operator.client.DeleteBucket(bucketName); err != nil {
		return err
	}

	return nil
}
