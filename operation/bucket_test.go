package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
)

var _ client.IS3 = (*mockS3)(nil)
var _ client.IS3 = (*allErrorMockS3)(nil)
var _ client.IS3 = (*deleteObjectsErrorMockS3)(nil)
var _ client.IS3 = (*deleteObjectsOutputErrorMockS3)(nil)
var _ client.IS3 = (*listObjectVersionsErrorMockS3)(nil)
var _ client.IS3 = (*deleteBucketErrorMockS3)(nil)

/*
	Mocks for client
*/
type mockS3 struct{}

func NewMockS3() *mockS3 {
	return &mockS3{}
}

func (m *mockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *mockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *mockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	output := []types.ObjectIdentifier{
		{
			Key:       aws.String("KeyForVersions"),
			VersionId: aws.String("VersionIdForVersions"),
		},
		{
			Key:       aws.String("KeyForDeleteMarkers"),
			VersionId: aws.String("VersionIdForDeleteMarkers"),
		},
	}
	return output, nil
}

type allErrorMockS3 struct{}

func NewAllErrorMockS3() *allErrorMockS3 {
	return &allErrorMockS3{}
}

func (m *allErrorMockS3) DeleteBucket(bucketName *string) error {
	return fmt.Errorf("DeleteBucketError")
}

func (m *allErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, fmt.Errorf("DeleteObjectsError")
}

func (m *allErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	return nil, fmt.Errorf("ListObjectVersionsError")
}

type deleteBucketErrorMockS3 struct{}

func NewDeleteBucketErrorMockS3() *deleteBucketErrorMockS3 {
	return &deleteBucketErrorMockS3{}
}

func (m *deleteBucketErrorMockS3) DeleteBucket(bucketName *string) error {
	return fmt.Errorf("DeleteBucketError")
}

func (m *deleteBucketErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *deleteBucketErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	output := []types.ObjectIdentifier{
		{
			Key:       aws.String("KeyForVersions"),
			VersionId: aws.String("VersionIdForVersions"),
		},
		{
			Key:       aws.String("KeyForDeleteMarkers"),
			VersionId: aws.String("VersionIdForDeleteMarkers"),
		},
	}
	return output, nil
}

type deleteObjectsErrorMockS3 struct{}

func NewDeleteObjectsErrorMockS3() *deleteObjectsErrorMockS3 {
	return &deleteObjectsErrorMockS3{}
}

func (m *deleteObjectsErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *deleteObjectsErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, fmt.Errorf("DeleteObjectsError")
}

func (m *deleteObjectsErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	output := []types.ObjectIdentifier{
		{
			Key:       aws.String("KeyForVersions"),
			VersionId: aws.String("VersionIdForVersions"),
		},
		{
			Key:       aws.String("KeyForDeleteMarkers"),
			VersionId: aws.String("VersionIdForDeleteMarkers"),
		},
	}
	return output, nil
}

type deleteObjectsOutputErrorMockS3 struct{}

func NewDeleteObjectsOutputErrorMockS3() *deleteObjectsOutputErrorMockS3 {
	return &deleteObjectsOutputErrorMockS3{}
}

func (m *deleteObjectsOutputErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *deleteObjectsOutputErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	output := []types.Error{
		{
			Key:       aws.String("Key"),
			Code:      aws.String("Code"),
			Message:   aws.String("Message"),
			VersionId: aws.String("VersionId"),
		},
	}
	return output, nil
}

func (m *deleteObjectsOutputErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	output := []types.ObjectIdentifier{
		{
			Key:       aws.String("KeyForVersions"),
			VersionId: aws.String("VersionIdForVersions"),
		},
		{
			Key:       aws.String("KeyForDeleteMarkers"),
			VersionId: aws.String("VersionIdForDeleteMarkers"),
		},
	}
	return output, nil
}

type listObjectVersionsErrorMockS3 struct{}

func NewListObjectVersionsErrorMockS3() *listObjectVersionsErrorMockS3 {
	return &listObjectVersionsErrorMockS3{}
}

func (m *listObjectVersionsErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *listObjectVersionsErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *listObjectVersionsErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	return nil, fmt.Errorf("ListObjectVersionsError")
}

/*
	Test Cases
*/
func TestDeleteBucket(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockS3()
	allErrorMock := NewAllErrorMockS3()
	deleteBucketErrorMock := NewDeleteBucketErrorMockS3()
	deleteObjectsErrorMock := NewDeleteObjectsErrorMockS3()
	deleteObjectsOutputErrorMock := NewDeleteObjectsOutputErrorMockS3()
	listObjectVersionsErrorMock := NewListObjectVersionsErrorMockS3()

	type args struct {
		ctx        context.Context
		bucketName *string
		client     client.IS3
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
		},
		{
			name: "delete bucket failure for all errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     allErrorMock,
			},
			want:    fmt.Errorf("ListObjectVersionsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for list object versions errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     listObjectVersionsErrorMock,
			},
			want:    fmt.Errorf("ListObjectVersionsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for delete objects errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     deleteObjectsErrorMock,
			},
			want:    fmt.Errorf("DeleteObjectsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for delete objects output errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     deleteObjectsOutputErrorMock,
			},
			want:    fmt.Errorf("DeleteObjectsError: followings \nCode: Code\nKey: Key\nVersionId: VersionId\nMessage: Message\n"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for delete bucket errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     deleteBucketErrorMock,
			},
			want:    fmt.Errorf("DeleteBucketError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bucketOperator := NewBucketOperator(tt.args.client)

			err := bucketOperator.DeleteBucket(tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestDeleteResourcesForBucket(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockS3()
	allErrorMock := NewAllErrorMockS3()

	type args struct {
		ctx    context.Context
		client client.IS3
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx:    ctx,
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    ctx,
				client: allErrorMock,
			},
			want:    fmt.Errorf("ListObjectVersionsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bucketOperator := NewBucketOperator(tt.args.client)
			bucketOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String(resourcetype.S3_STACK),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := bucketOperator.DeleteResources()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
