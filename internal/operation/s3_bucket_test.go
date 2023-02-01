package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "github.com/golang/mock/gomock"
)

/*
	Test Cases
*/

func TestS3BucketOperator_DeleteS3Bucket(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx        context.Context
		bucketName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3)
		want          error
		wantErr       bool
	}{
		{
			name: "delete bucket successfully",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteBucket(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for check bucket exists errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("ListBucketsError"))
			},
			want:    fmt.Errorf("ListBucketsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for bucket not exists",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for list object versions errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(nil, fmt.Errorf("ListObjectVersionsError"))
			},
			want:    fmt.Errorf("ListObjectVersionsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for delete objects errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("test"), gomock.Any()).Return([]types.Error{}, fmt.Errorf("DeleteObjectsError"))
			},
			want:    fmt.Errorf("DeleteObjectsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for delete objects errors after zero length",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("test"), gomock.Any()).Return([]types.Error{}, fmt.Errorf("DeleteObjectsErrorAfterZeroLength"))
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for delete objects output errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("test"), gomock.Any()).Return([]types.Error{
					{
						Key:       aws.String("Key"),
						Code:      aws.String("Code"),
						Message:   aws.String("Message"),
						VersionId: aws.String("VersionId"),
					},
				}, nil)
			},
			want:    fmt.Errorf("DeleteObjectsError: followings \nCode: Code\nKey: Key\nVersionId: VersionId\nMessage: Message\n"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for delete objects output errors after zero length",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("test"), gomock.Any()).Return([]types.Error{
					{
						Key:       aws.String("Key"),
						Code:      aws.String("Code"),
						Message:   aws.String("Message"),
						VersionId: aws.String("VersionId"),
					},
				}, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure for delete bucket errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("test")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("test"), gomock.Any()).Return([]types.Error{}, nil)
				m.EXPECT().DeleteBucket(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteBucketError"))
			},
			want:    fmt.Errorf("DeleteBucketError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3Mock := client.NewMockIS3(ctrl)
			tt.prepareMockFn(s3Mock)

			s3BucketOperator := NewS3BucketOperator(s3Mock)

			err := s3BucketOperator.DeleteS3Bucket(tt.args.ctx, tt.args.bucketName)
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

func TestS3BucketOperator_DeleteResourcesForS3Bucket(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx    context.Context
		client client.IS3
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().ListObjectVersions(gomock.Any(), aws.String("PhysicalResourceId1")).Return(
					[]types.ObjectIdentifier{
						{
							Key:       aws.String("KeyForVersions"),
							VersionId: aws.String("VersionIdForVersions"),
						},
						{
							Key:       aws.String("KeyForDeleteMarkers"),
							VersionId: aws.String("VersionIdForDeleteMarkers"),
						},
					}, nil)
				m.EXPECT().DeleteObjects(gomock.Any(), aws.String("PhysicalResourceId1"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteBucket(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3) {
				m.EXPECT().CheckBucketExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("ListBucketsError"))
			},
			want:    fmt.Errorf("ListBucketsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3Mock := client.NewMockIS3(ctrl)
			tt.prepareMockFn(s3Mock)

			s3BucketOperator := NewS3BucketOperator(s3Mock)

			s3BucketOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::S3::Bucket"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := s3BucketOperator.DeleteResources(tt.args.ctx)
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
