package operations

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
)

func DeleteBuckets(config aws.Config, resources []cfnTypes.StackResourceSummary) error {
	// TODO: Concurrency Delete
	s3Client := client.NewS3(config)
	for _, bucket := range resources {
		err := DeleteBucket(s3Client, *bucket.PhysicalResourceId)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteBucket(s3Client *client.S3, bucketName string) error {
	versions, err := s3Client.ListObjectVersions(&bucketName)
	if err != nil {
		return err
	}

	errors, err := s3Client.DeleteObjects(&bucketName, versions)
	if err != nil {
		return err
	}
	if len(errors) > 0 {
		return fmt.Errorf("DeleteObjects Error: %v", errors)
	}

	if err := s3Client.DeleteBucket(&bucketName); err != nil {
		return err
	}

	return nil
}
