package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-to-k/delstack/logger"
)

var _ IS3SDKClient = (*mockS3SDKClient)(nil)
var _ IS3SDKClient = (*errorMockS3SDKClient)(nil)

type mockS3SDKClient struct{}

func NewMockS3SDKClient() *mockS3SDKClient {
	return &mockS3SDKClient{}
}

func (m *mockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *mockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *mockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, nil
}

type errorMockS3SDKClient struct{}

func NewErrorMockS3SDKClient() *errorMockS3SDKClient {
	return &errorMockS3SDKClient{}
}

func (m *errorMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, fmt.Errorf("DeleteBucketError")
}

func (m *errorMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, fmt.Errorf("DeleteBucketError")
}

func (m *errorMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, fmt.Errorf("DeleteBucketError")
}

func TestDeleteBucket(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()

	type args struct {
		ctx        context.Context
		bucketName *string
		client     IS3SDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete bucket successfully",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     mock,
			},
			want:    nil,
			wantErr: false,
		}, {
			name: "delete bucket failure",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     errorMock,
			},
			want:    fmt.Errorf("DeleteBucketError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s3Client := NewS3(tt.args.client)

			err := s3Client.DeleteBucket(tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}

}
