package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var _ IS3SDKClient = (*MockS3SDKClient)(nil)
var _ IS3SDKClient = (*ErrorMockS3SDKClient)(nil)
var _ IS3SDKClient = (*ApiErrorMockS3SDKClient)(nil)
var _ IS3SDKClient = (*OutputErrorForDeleteObjectsMockS3SDKClient)(nil)
var _ IS3SDKClient = (*EmptyMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*VersionsMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*DeleteMarkersMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*NotExistsMockForListBucketsS3SDKClient)(nil)

/*
	Mocks for SDK Client
*/

type MockS3SDKClient struct{}

func NewMockS3SDKClient() *MockS3SDKClient {
	return &MockS3SDKClient{}
}

func (m *MockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *MockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, nil
}

func (m *MockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{
				Key:       aws.String("KeyForVersions"),
				VersionId: aws.String("VersionIdForVersions"),
			},
		},
		DeleteMarkers: []types.DeleteMarkerEntry{
			{
				Key:       aws.String("KeyForDeleteMarkers"),
				VersionId: aws.String("VersionIdForDeleteMarkers"),
			},
		},
	}
	return output, nil
}

func (m *MockS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	output := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("test"),
			},
			{
				Name: aws.String("test2"),
			},
		},
	}
	return output, nil
}

type ErrorMockS3SDKClient struct{}

func NewErrorMockS3SDKClient() *ErrorMockS3SDKClient {
	return &ErrorMockS3SDKClient{}
}

func (m *ErrorMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, fmt.Errorf("DeleteBucketError")
}

func (m *ErrorMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, fmt.Errorf("DeleteObjectsError")
}

func (m *ErrorMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, fmt.Errorf("ListObjectVersionsError")
}

func (m *ErrorMockS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return nil, fmt.Errorf("ListBucketsError")
}

type ApiErrorMockS3SDKClient struct{}

func NewApiErrorMockS3SDKClient() *ApiErrorMockS3SDKClient {
	return &ApiErrorMockS3SDKClient{}
}

func (m *ApiErrorMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, fmt.Errorf("api error SlowDown")
}

func (m *ApiErrorMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, fmt.Errorf("api error SlowDown")
}

func (m *ApiErrorMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, fmt.Errorf("api error SlowDown")
}

func (m *ApiErrorMockS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	return nil, fmt.Errorf("api error SlowDown")
}

type OutputErrorForDeleteObjectsMockS3SDKClient struct{}

func NewOutputErrorForDeleteObjectsMockS3SDKClient() *OutputErrorForDeleteObjectsMockS3SDKClient {
	return &OutputErrorForDeleteObjectsMockS3SDKClient{}
}

func (m *OutputErrorForDeleteObjectsMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *OutputErrorForDeleteObjectsMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors: []types.Error{
			{
				Key:       aws.String("Key"),
				Code:      aws.String("Code"),
				Message:   aws.String("Message"),
				VersionId: aws.String("VersionId"),
			},
		},
	}
	return output, nil
}

func (m *OutputErrorForDeleteObjectsMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, nil
}

func (m *OutputErrorForDeleteObjectsMockS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	output := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("test"),
			},
			{
				Name: aws.String("test2"),
			},
		},
	}
	return output, nil
}

type EmptyMockForListObjectVersionsS3SDKClient struct{}

func NewEmptyMockForListObjectVersionsS3SDKClient() *EmptyMockForListObjectVersionsS3SDKClient {
	return &EmptyMockForListObjectVersionsS3SDKClient{}
}

func (m *EmptyMockForListObjectVersionsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *EmptyMockForListObjectVersionsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *EmptyMockForListObjectVersionsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions:      []types.ObjectVersion{},
		DeleteMarkers: []types.DeleteMarkerEntry{},
	}
	return output, nil
}

func (m *EmptyMockForListObjectVersionsS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	output := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("test"),
			},
			{
				Name: aws.String("test2"),
			},
		},
	}
	return output, nil
}

type VersionsMockForListObjectVersionsS3SDKClient struct{}

func NewVersionsMockForListObjectVersionsS3SDKClient() *VersionsMockForListObjectVersionsS3SDKClient {
	return &VersionsMockForListObjectVersionsS3SDKClient{}
}

func (m *VersionsMockForListObjectVersionsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *VersionsMockForListObjectVersionsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *VersionsMockForListObjectVersionsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{
				Key:       aws.String("KeyForVersions"),
				VersionId: aws.String("VersionIdForVersions"),
			},
		},
		DeleteMarkers: []types.DeleteMarkerEntry{},
	}
	return output, nil
}

func (m *VersionsMockForListObjectVersionsS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	output := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("test"),
			},
			{
				Name: aws.String("test2"),
			},
		},
	}
	return output, nil
}

type DeleteMarkersMockForListObjectVersionsS3SDKClient struct{}

func NewDeleteMarkersMockForListObjectVersionsS3SDKClient() *DeleteMarkersMockForListObjectVersionsS3SDKClient {
	return &DeleteMarkersMockForListObjectVersionsS3SDKClient{}
}

func (m *DeleteMarkersMockForListObjectVersionsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *DeleteMarkersMockForListObjectVersionsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *DeleteMarkersMockForListObjectVersionsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{},
		DeleteMarkers: []types.DeleteMarkerEntry{
			{
				Key:       aws.String("KeyForDeleteMarkers"),
				VersionId: aws.String("VersionIdForDeleteMarkers"),
			},
		},
	}
	return output, nil
}

func (m *DeleteMarkersMockForListObjectVersionsS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	output := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("test"),
			},
			{
				Name: aws.String("test2"),
			},
		},
	}
	return output, nil
}

type NotExistsMockForListBucketsS3SDKClient struct{}

func NewNotExistsMockForListBucketsS3SDKClient() *NotExistsMockForListBucketsS3SDKClient {
	return &NotExistsMockForListBucketsS3SDKClient{}
}

func (m *NotExistsMockForListBucketsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *NotExistsMockForListBucketsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, nil
}

func (m *NotExistsMockForListBucketsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{
				Key:       aws.String("KeyForVersions"),
				VersionId: aws.String("VersionIdForVersions"),
			},
		},
		DeleteMarkers: []types.DeleteMarkerEntry{
			{
				Key:       aws.String("KeyForDeleteMarkers"),
				VersionId: aws.String("VersionIdForDeleteMarkers"),
			},
		},
	}
	return output, nil
}

func (m *NotExistsMockForListBucketsS3SDKClient) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	output := &s3.ListBucketsOutput{
		Buckets: []types.Bucket{
			{
				Name: aws.String("test0"),
			},
			{
				Name: aws.String("test2"),
			},
		},
	}
	return output, nil
}
