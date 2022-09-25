package operation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*BucketOperator)(nil)
var sleepTimeSecForS3 = 10

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

func (operator *BucketOperator) DeleteResources() error {
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(option.ConcurrencyNum))

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
	exists, err := operator.client.CheckBucketExists(bucketName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	versions, err := operator.client.ListObjectVersions(bucketName)
	if err != nil {
		return err
	}

	if len(versions) > 0 {
		errors, err := operator.client.DeleteObjects(bucketName, versions, sleepTimeSecForS3)
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
	if err := operator.client.DeleteBucket(bucketName); err != nil {
		return err
	}

	return nil
}
