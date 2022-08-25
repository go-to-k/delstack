package operations

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/client"
)

func DeleteBuckets(config aws.Config, resources []cfnTypes.StackResourceSummary) error {
	// TODO: Concurrency Delete
	s3Client := client.NewS3(config)
	for _, bucket := range resources {
		objectErrors, err := DeleteBucketResources(s3Client, *bucket.PhysicalResourceId)
		if err != nil {
			return err
		} else if objectErrors != nil {
			return fmt.Errorf("%v", objectErrors)
		}
	}
	return nil
}

func DeleteBucketResources(s3Client *client.S3, bucketName string) ([]s3Types.Error, error) {
	version, err := s3Client.ListObjectVersions(&bucketName)
	if err != nil {
		return nil, err
	}

	errors, err := s3Client.DeleteObjects(&bucketName, version)
	if err != nil {
		return nil, err
	}

	return errors, nil
}
