package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
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
	Test Cases
*/

func TestBucketOperator_DeleteBucket(t *testing.T) {
	io.NewLogger(false)
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
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for all errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     allErrorMock,
			},
			want:    fmt.Errorf("ListBucketsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for check bucket exists errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     checkBucketExistsErrorMock,
			},
			want:    fmt.Errorf("ListBucketsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for bucket not exists",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     checkBucketNotExistsMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for list object versions errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     listObjectVersionsErrorMock,
			},
			want:    fmt.Errorf("ListObjectVersionsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for delete objects errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     deleteObjectsErrorMock,
			},
			want:    fmt.Errorf("DeleteObjectsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for delete objects errors after zero length",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     deleteObjectsErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for delete objects output errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     deleteObjectsOutputErrorMock,
			},
			want:    fmt.Errorf("DeleteObjectsError: followings \nCode: Code\nKey: Key\nVersionId: VersionId\nMessage: Message\n"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for delete objects output errors after zero length",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				client:     deleteObjectsOutputErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for delete bucket errors",
			args: args{
				ctx:        context.Background(),
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

			err := bucketOperator.DeleteBucket(tt.args.ctx, tt.args.bucketName)
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
	io.NewLogger(false)
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
				ctx:    context.Background(),
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    context.Background(),
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
				ResourceType:       aws.String("AWS::S3::Bucket"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := bucketOperator.DeleteResources(tt.args.ctx)
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
