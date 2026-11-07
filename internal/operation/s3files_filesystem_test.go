package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestS3FilesFileSystemOperator_DeleteS3FilesFileSystem(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx          context.Context
		fileSystemId *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Files)
		want          error
		wantErr       bool
	}{
		{
			name: "delete s3 files file system successfully",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3Files) {
				m.EXPECT().CheckFileSystemExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DeleteFileSystem(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete s3 files file system failure",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3Files) {
				m.EXPECT().CheckFileSystemExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DeleteFileSystem(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteFileSystemError"))
			},
			want:    fmt.Errorf("DeleteFileSystemError"),
			wantErr: true,
		},
		{
			name: "delete s3 files file system failure for check s3 files file system exists errors",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3Files) {
				m.EXPECT().CheckFileSystemExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("GetFileSystemError"))
			},
			want:    fmt.Errorf("GetFileSystemError"),
			wantErr: true,
		},
		{
			name: "delete s3 files file system successfully for s3 files file system not exists",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIS3Files) {
				m.EXPECT().CheckFileSystemExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3FilesMock := client.NewMockIS3Files(ctrl)
			tt.prepareMockFn(s3FilesMock)

			s3FilesFileSystemOperator := NewS3FilesFileSystemOperator(s3FilesMock)

			err := s3FilesFileSystemOperator.DeleteS3FilesFileSystem(tt.args.ctx, tt.args.fileSystemId)
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

func TestS3FilesFileSystemOperator_DeleteResourcesForS3FilesFileSystem(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Files)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3Files) {
				m.EXPECT().CheckFileSystemExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().DeleteFileSystem(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3Files) {
				m.EXPECT().CheckFileSystemExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("GetFileSystemError"))
			},
			want:    fmt.Errorf("GetFileSystemError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3FilesMock := client.NewMockIS3Files(ctrl)
			tt.prepareMockFn(s3FilesMock)

			s3FilesFileSystemOperator := NewS3FilesFileSystemOperator(s3FilesMock)
			s3FilesFileSystemOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::S3Files::FileSystem"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := s3FilesFileSystemOperator.DeleteResources(tt.args.ctx)
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
