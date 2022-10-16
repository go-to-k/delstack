package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/go-to-k/delstack/pkg/logger"
)

var _ client.IS3 = (*MockS3)(nil)
var _ client.IS3 = (*AllErrorMockS3)(nil)
var _ client.IS3 = (*DeleteObjectsErrorMockS3)(nil)
var _ client.IS3 = (*DeleteObjectsErrorAfterZeroLengthMockS3)(nil)
var _ client.IS3 = (*DeleteObjectsOutputErrorMockS3)(nil)
var _ client.IS3 = (*DeleteObjectsOutputErrorAfterZeroLengthMockS3)(nil)
var _ client.IS3 = (*ListObjectVersionsErrorMockS3)(nil)
var _ client.IS3 = (*DeleteBucketErrorMockS3)(nil)
var _ client.IS3 = (*CheckBucketExistsErrorMockS3)(nil)
var _ client.IS3 = (*CheckBucketNotExistsMockS3)(nil)

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

func (m *MockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
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

func (m *AllErrorMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return false, fmt.Errorf("ListBucketsError")
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

func (m *DeleteBucketErrorMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
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

func (m *DeleteObjectsErrorMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
}

type DeleteObjectsErrorAfterZeroLengthMockS3 struct{}

func NewDeleteObjectsErrorAfterZeroLengthMockS3() *DeleteObjectsErrorAfterZeroLengthMockS3 {
	return &DeleteObjectsErrorAfterZeroLengthMockS3{}
}

func (m *DeleteObjectsErrorAfterZeroLengthMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *DeleteObjectsErrorAfterZeroLengthMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, fmt.Errorf("DeleteObjectsErrorAfterZeroLength")
}

func (m *DeleteObjectsErrorAfterZeroLengthMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	output := []types.ObjectIdentifier{}
	return output, nil
}

func (m *DeleteObjectsErrorAfterZeroLengthMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
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

func (m *DeleteObjectsOutputErrorMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
}

type DeleteObjectsOutputErrorAfterZeroLengthMockS3 struct{}

func NewDeleteObjectsOutputErrorAfterZeroLengthMockS3() *DeleteObjectsOutputErrorAfterZeroLengthMockS3 {
	return &DeleteObjectsOutputErrorAfterZeroLengthMockS3{}
}

func (m *DeleteObjectsOutputErrorAfterZeroLengthMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *DeleteObjectsOutputErrorAfterZeroLengthMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
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

func (m *DeleteObjectsOutputErrorAfterZeroLengthMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	output := []types.ObjectIdentifier{}
	return output, nil
}

func (m *DeleteObjectsOutputErrorAfterZeroLengthMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
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

func (m *ListObjectVersionsErrorMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return true, nil
}

type CheckBucketExistsErrorMockS3 struct{}

func NewCheckBucketExistsErrorMockS3() *CheckBucketExistsErrorMockS3 {
	return &CheckBucketExistsErrorMockS3{}
}

func (m *CheckBucketExistsErrorMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *CheckBucketExistsErrorMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *CheckBucketExistsErrorMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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

func (m *CheckBucketExistsErrorMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return false, fmt.Errorf("ListBucketsError")
}

type CheckBucketNotExistsMockS3 struct{}

func NewCheckBucketNotExistsMockS3() *CheckBucketNotExistsMockS3 {
	return &CheckBucketNotExistsMockS3{}
}

func (m *CheckBucketNotExistsMockS3) DeleteBucket(bucketName *string) error {
	return nil
}

func (m *CheckBucketNotExistsMockS3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	return []types.Error{}, nil
}

func (m *CheckBucketNotExistsMockS3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
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

func (m *CheckBucketNotExistsMockS3) CheckBucketExists(bucketName *string) (bool, error) {
	return false, nil
}

/*
	Test Cases
*/
func TestBucketOperator_DeleteBucket(t *testing.T) {
	logger.NewLogger(false)
	ctx := context.TODO()
	mock := NewMockS3()
	allErrorMock := NewAllErrorMockS3()
	deleteBucketErrorMock := NewDeleteBucketErrorMockS3()
	deleteObjectsErrorMock := NewDeleteObjectsErrorMockS3()
	deleteObjectsErrorAfterZeroLengthMock := NewDeleteObjectsErrorAfterZeroLengthMockS3()
	deleteObjectsOutputErrorMock := NewDeleteObjectsOutputErrorMockS3()
	deleteObjectsOutputErrorAfterZeroLengthMock := NewDeleteObjectsOutputErrorAfterZeroLengthMockS3()
	listObjectVersionsErrorMock := NewListObjectVersionsErrorMockS3()
	checkBucketExistsErrorMock := NewCheckBucketExistsErrorMockS3()
	checkBucketNotExistsMock := NewCheckBucketNotExistsMockS3()

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
			want:    fmt.Errorf("ListBucketsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for check bucket exists errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     checkBucketExistsErrorMock,
			},
			want:    fmt.Errorf("ListBucketsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for bucket not exists",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     checkBucketNotExistsMock,
			},
			want:    nil,
			wantErr: false,
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
			name: "delete bucket successfully for delete objects errors after zero length",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     deleteObjectsErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
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
			name: "delete bucket failure for delete objects output errors after zero length",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     deleteObjectsOutputErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
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

func TestBucketOperator_DeleteResourcesForBucket(t *testing.T) {
	logger.NewLogger(false)
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
			want:    fmt.Errorf("ListBucketsError"),
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
