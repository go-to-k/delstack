package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-to-k/delstack/pkg/client"
)

type BucketOperatorFactory struct {
	config aws.Config
}

func NewBucketOperatorFactory(config aws.Config) *BucketOperatorFactory {
	return &BucketOperatorFactory{config}
}

func (f *BucketOperatorFactory) CreateBucketOperator() *BucketOperator {
	return NewBucketOperator(
		f.createBucketClient(),
	)
}

func (f *BucketOperatorFactory) createBucketClient() *client.S3 {
	sdkBucketClient := s3.NewFromConfig(f.config)

	return client.NewS3(
		sdkBucketClient,
	)
}
