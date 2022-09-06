package operation

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
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
	var semaphore = make(chan struct{}, option.CONCURRENCY_NUM)

	for _, bucket := range operator.resources {
		bucket := bucket
		eg.Go(func() error {
			semaphore <- struct{}{}

			if err := operator.DeleteBucket(bucket.PhysicalResourceId); err != nil {
				return err
			}
			<-semaphore

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
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
		return fmt.Errorf("DeleteObjects Error: %v", errors)
	}

	if err := operator.client.DeleteBucket(bucketName); err != nil {
		return err
	}

	return nil
}
