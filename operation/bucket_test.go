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

var _ client.IS3 = (*MockS3)(nil)
var _ client.IS3 = (*AllErrorMockS3)(nil)
var _ client.IS3 = (*DeleteObjectsErrorMockS3)(nil)
var _ client.IS3 = (*DeleteObjectsOutputErrorMockS3)(nil)
var _ client.IS3 = (*ListObjectVersionsErrorMockS3)(nil)
var _ client.IS3 = (*DeleteBucketErrorMockS3)(nil)

/*
	Mocks for client
*/
type MockS3 struct{}

func NewMockS3() *MockS3 {
	return &MockS3{}
}

func (m *MockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *MockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *MockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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

type AllErrorMockS3 struct{}

func NewAllErrorMockS3() *AllErrorMockS3 {
	return &AllErrorMockS3{}
}

func (m *AllErrorMockS3) DeleteBucket(bucketName *string) error {
	return fmt.Errorf("DeleteBucketError")
}

func (m *AllErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, fmt.Errorf("DeleteObjectsError")
}

func (m *AllErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	return nil, fmt.Errorf("ListObjectVersionsError")
}

type DeleteBucketErrorMockS3 struct{}

func NewDeleteBucketErrorMockS3() *DeleteBucketErrorMockS3 {
	return &DeleteBucketErrorMockS3{}
}

func (m *DeleteBucketErrorMockS3) DeleteBucket(bucketName *string) error {
	return fmt.Errorf("DeleteBucketError")
}

func (m *DeleteBucketErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *DeleteBucketErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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

type DeleteObjectsErrorMockS3 struct{}

func NewDeleteObjectsErrorMockS3() *DeleteObjectsErrorMockS3 {
	return &DeleteObjectsErrorMockS3{}
}

func (m *DeleteObjectsErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *DeleteObjectsErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, fmt.Errorf("DeleteObjectsError")
}

func (m *DeleteObjectsErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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

type DeleteObjectsOutputErrorMockS3 struct{}

func NewDeleteObjectsOutputErrorMockS3() *DeleteObjectsOutputErrorMockS3 {
	return &DeleteObjectsOutputErrorMockS3{}
}

func (m *DeleteObjectsOutputErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *DeleteObjectsOutputErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
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

func (m *DeleteObjectsOutputErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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

type ListObjectVersionsErrorMockS3 struct{}

func NewListObjectVersionsErrorMockS3() *ListObjectVersionsErrorMockS3 {
	return &ListObjectVersionsErrorMockS3{}
}

func (m *ListObjectVersionsErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *ListObjectVersionsErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *ListObjectVersionsErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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
				ResourceType:       aws.String(resourcetype.S3_BUCKET),
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
