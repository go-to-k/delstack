package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-to-k/delstack/client"
)

type BucketOperatorFactory struct {
	config aws.Config
}

func NewBucketOperatorFactory(config aws.Config) *BucketOperatorFactory {
	return &BucketOperatorFactory{config}
}

func (factory *BucketOperatorFactory) CreateBucketOperator() *BucketOperator {
	return NewBucketOperator(
		factory.createBucketClient(),
	)
}

func (factory *BucketOperatorFactory) createBucketClient() *client.S3 {
	sdkBucketClient := s3.NewFromConfig(factory.config)

	return client.NewS3(
		factory.config,
		sdkBucketClient,
	)
}
