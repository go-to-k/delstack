package client

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var _ IS3SDKClient = (*mockS3SDKClient)(nil)

type mockS3SDKClient struct {
}

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

func TestDeleteBucket(t *testing.T) {
	ctx := context.TODO()
	mock := NewMockS3SDKClient()
	s3Client := NewS3(mock)

	type args struct {
		ctx        context.Context
		bucketName *string
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete bucket",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := s3Client.DeleteBucket(tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("handler() = %v, want %v", got, tt.want)
			// }
		})
	}

}
