package client

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
)

const sleepTimeSecForS3 = 1

/*
	Test Cases
*/

func TestS3_DeleteBucket(t *testing.T) {
	type args struct {
		ctx                context.Context
		bucketName         *string
		withAPIOptionsFunc func(*middleware.Stack) error
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteBucketMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteBucketOutput{},
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
			name: "delete bucket failure",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteBucketErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("DeleteBucketError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error S3: DeleteBucket, DeleteBucketError"),
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

			client := s3.NewFromConfig(cfg)
			s3Client := NewS3(client)

			err = s3Client.DeleteBucket(tt.args.ctx, tt.args.bucketName)
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

func TestS3_DeleteObjects(t *testing.T) {
	objectsOverLimit := []types.ObjectIdentifier{}
	s3DeleteObjectsSizeOverLimit := s3DeleteObjectsSizeLimit*int(runtime.NumCPU())*2 + 1 // loop over cpu core size for channel waiting when next loop
	for i := 0; i < s3DeleteObjectsSizeOverLimit; i++ {
		objectsOverLimit = append(objectsOverLimit, types.ObjectIdentifier{
			Key:       aws.String("Key"),
			VersionId: aws.String("VersionId"),
		})
	}

	type args struct {
		ctx                context.Context
		bucketName         *string
		objects            []types.ObjectIdentifier
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output []types.Error
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "delete objects successfully",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteObjectsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteObjectsOutput{
										Deleted: []types.DeletedObject{},
										Errors:  []types.Error{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.Error{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects successfully if zero objects",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				objects:    []types.ObjectIdentifier{},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteObjectsIfZeroObjectsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteObjectsOutput{
										Deleted: []types.DeletedObject{},
										Errors:  []types.Error{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.Error{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects over limit successfully",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				objects:    objectsOverLimit,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteObjectsOverLimitMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteObjectsOutput{
										Deleted: []types.DeletedObject{},
										Errors:  []types.Error{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.Error{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects failure",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteObjectsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteObjectsOutput{
										Deleted: []types.DeletedObject{},
										Errors:  []types.Error{},
									},
								}, middleware.Metadata{}, fmt.Errorf("DeleteObjectsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("operation error S3: DeleteObjects, DeleteObjectsError"),
			},
			wantErr: true,
		},
		{
			name: "delete objects failure for api error",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteObjectsApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteObjectsOutput{
										Deleted: []types.DeletedObject{},
										Errors:  []types.Error{},
									},
								}, middleware.Metadata{}, fmt.Errorf("api error SlowDown")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("RetryCountOverError: test, operation error S3: DeleteObjects, api error SlowDown\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			},
			wantErr: true,
		},
		{
			name: "delete objects failure for output errors",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteObjectsOutputErrorsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.DeleteObjectsOutput{
										Deleted: []types.DeletedObject{},
										Errors: []types.Error{
											{
												Key:       aws.String("Key"),
												Code:      aws.String("Code"),
												Message:   aws.String("Message"),
												VersionId: aws.String("VersionId"),
											},
										},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.Error{
					{
						Key:       aws.String("Key"),
						Code:      aws.String("Code"),
						Message:   aws.String("Message"),
						VersionId: aws.String("VersionId"),
					},
				},
				err: nil,
			},
			wantErr: false,
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

			client := s3.NewFromConfig(cfg)
			s3Client := NewS3(client)

			output, err := s3Client.DeleteObjects(tt.args.ctx, tt.args.bucketName, tt.args.objects, sleepTimeSecForS3)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}

func TestS3_ListObjectVersions(t *testing.T) {
	type args struct {
		ctx                context.Context
		bucketName         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output []types.ObjectIdentifier
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list objects versions successfully",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListObjectVersionsOutput{
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
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForVersions"),
						VersionId: aws.String("VersionIdForVersions"),
					},
					{
						Key:       aws.String("KeyForDeleteMarkers"),
						VersionId: aws.String("VersionIdForDeleteMarkers"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions failure",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListObjectVersionsOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListObjectVersionsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("operation error S3: ListObjectVersions, ListObjectVersionsError"),
			},
			wantErr: true,
		},
		{
			name: "list objects versions successfully(empty)",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListObjectVersionsOutput{
										Versions:      []types.ObjectVersion{},
										DeleteMarkers: []types.DeleteMarkerEntry{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.ObjectIdentifier{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions successfully(versions only)",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsWithVersionsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListObjectVersionsOutput{
										Versions: []types.ObjectVersion{
											{
												Key:       aws.String("KeyForVersions"),
												VersionId: aws.String("VersionIdForVersions"),
											},
										},
										DeleteMarkers: []types.DeleteMarkerEntry{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForVersions"),
						VersionId: aws.String("VersionIdForVersions"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions successfully(delete markers only)",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsWithDeleteMarkersMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListObjectVersionsOutput{
										Versions: []types.ObjectVersion{},
										DeleteMarkers: []types.DeleteMarkerEntry{
											{
												Key:       aws.String("KeyForDeleteMarkers"),
												VersionId: aws.String("VersionIdForDeleteMarkers"),
											},
										},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForDeleteMarkers"),
						VersionId: aws.String("VersionIdForDeleteMarkers"),
					},
				},
				err: nil,
			},
			wantErr: false,
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

			client := s3.NewFromConfig(cfg)
			s3Client := NewS3(client)

			output, err := s3Client.ListObjectVersions(tt.args.ctx, tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}

func TestS3_CheckBucketExists(t *testing.T) {
	type args struct {
		ctx                context.Context
		bucketName         *string
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
			name: "check bucket for bucket exists",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBucketsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListBucketsOutput{
										Buckets: []types.Bucket{
											{
												Name: aws.String("test"),
											},
											{
												Name: aws.String("test2"),
											},
										},
									},
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
			name: "check bucket for bucket do not exist",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBucketsNotExistMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3.ListBucketsOutput{
										Buckets: []types.Bucket{
											{
												Name: aws.String("test0"),
											},
											{
												Name: aws.String("test2"),
											},
										},
									},
								}, middleware.Metadata{}, nil
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
			name: "check bucket exists failure",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBucketsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("ListBucketsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("operation error S3: ListBuckets, ListBucketsError"),
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

			client := s3.NewFromConfig(cfg)
			s3Client := NewS3(client)

			output, err := s3Client.CheckBucketExists(tt.args.ctx, tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.exists) {
				t.Errorf("output = %#v, want %#v", output, tt.want.exists)
			}
		})
	}
}
