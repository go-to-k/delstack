package client

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3 struct {
	client *s3.Client
}

func NewS3(config aws.Config) *S3 {
	client := s3.NewFromConfig(config)
	return &S3{
		client,
	}
}

func (s3Bucket *S3) DeleteBucket(bucketName *string) error {
	input := &s3.DeleteBucketInput{
		Bucket: bucketName,
	}

	_, err := s3Bucket.client.DeleteBucket(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed delete the s3 bucket, %v", err)
		return err
	}

	return nil
}

func (s3Bucket *S3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier) ([]types.Error, error) {
	// TODO: 1000以上はループさせてエラーにしない
	if len(objects) > 1000 {
		err := fmt.Errorf("over 1000 objects error")
		log.Fatalf("failed delete objects, %v", err)
		return nil, err
	}

	input := &s3.DeleteObjectsInput{
		Bucket: bucketName,
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   *aws.Bool(true),
		},
	}

	output, err := s3Bucket.client.DeleteObjects(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed delete objects, %v", err)
		return nil, err
	}

	return output.Errors, nil
}

func (s3Bucket *S3) ListObjects(bucketName *string) ([]types.ObjectIdentifier, error) {
	objects := []types.ObjectIdentifier{}
	nextContinuationToken := ""
	isTruncated := true

	for isTruncated {
		output, err := s3Bucket.iterateListObjects(bucketName, &nextContinuationToken)
		if err != nil {
			return nil, err
		}

		for _, v := range output.Contents {
			objects = append(objects, types.ObjectIdentifier{
				Key: v.Key,
			})
		}

		isTruncated = output.IsTruncated
		nextContinuationToken = *output.ContinuationToken
	}

	return objects, nil
}

func (s3Bucket *S3) iterateListObjects(bucketName *string, nextContinuationToken *string) (*s3.ListObjectsV2Output, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:            bucketName,
		ContinuationToken: nextContinuationToken,
	}

	output, err := s3Bucket.client.ListObjectsV2(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed list objects, %v", err)
		return nil, err
	}

	return output, nil
}

func (s3Bucket *S3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	versions := []types.ObjectIdentifier{}
	input := &s3.ListObjectVersionsInput{
		Bucket: bucketName,
	}

	output, err := s3Bucket.client.ListObjectVersions(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed list object versions, %v", err)
		return nil, err
	}

	// TODO: Next〜の必要あるならiterateListObjectsと合わせる
	for _, v := range output.Versions {
		versions = append(versions, types.ObjectIdentifier{
			Key:       v.Key,
			VersionId: v.VersionId,
		})
	}

	return versions, nil
}
