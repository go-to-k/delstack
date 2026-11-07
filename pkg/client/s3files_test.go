package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/aws/smithy-go/middleware"
)

/*
	Test Cases
*/

func TestS3Files_DeleteFileSystem(t *testing.T) {
	type args struct {
		ctx                context.Context
		fileSystemId       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete file system with force delete successfully",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("fs-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"DeleteFileSystemMock",
							func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
								input, ok := in.Parameters.(*s3files.DeleteFileSystemInput)
								if !ok || !aws.ToBool(input.ForceDelete) {
									return middleware.InitializeOutput{}, middleware.Metadata{}, fmt.Errorf("ForceDelete is not true")
								}
								return middleware.InitializeOutput{
									Result: &s3files.DeleteFileSystemOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete file system successfully for file system not exists",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("fs-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteFileSystemNotFoundMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3files.DeleteFileSystemOutput{},
								}, middleware.Metadata{}, &types.ResourceNotFoundException{Message: aws.String("not found")}
							},
						),
						middleware.Before,
					)
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete file system failure",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("fs-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteFileSystemErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3files.DeleteFileSystemOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteFileSystemError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("fs-test"),
				Err:          fmt.Errorf("operation error S3Files: DeleteFileSystem, DeleteFileSystemError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := s3files.NewFromConfig(cfg)
			s3FilesClient := NewS3Files(client)

			err = s3FilesClient.DeleteFileSystem(tt.args.ctx, tt.args.fileSystemId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestS3Files_CheckFileSystemExists(t *testing.T) {
	type args struct {
		ctx                context.Context
		fileSystemId       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		exists bool
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "check file system exists successfully",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("fs-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFileSystemMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3files.GetFileSystemOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check file system not exists successfully",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("fs-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFileSystemNotFoundMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3files.GetFileSystemOutput{},
								}, middleware.Metadata{}, &types.ResourceNotFoundException{Message: aws.String("not found")}
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check file system exists failure",
			args: args{
				ctx:          context.Background(),
				fileSystemId: aws.String("fs-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFileSystemErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3files.GetFileSystemOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetFileSystemError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("fs-test"),
					Err:          fmt.Errorf("operation error S3Files: GetFileSystem, GetFileSystemError"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := s3files.NewFromConfig(cfg)
			s3FilesClient := NewS3Files(client)

			output, err := s3FilesClient.CheckFileSystemExists(tt.args.ctx, tt.args.fileSystemId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want.err)
			}
			if !tt.wantErr && output != tt.want.exists {
				t.Errorf("output = %#v, want %#v", output, tt.want.exists)
			}
		})
	}
}
