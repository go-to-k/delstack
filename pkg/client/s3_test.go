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

type keyMarkerKeyForS3 struct{}
type versionIdMarkerKeyForS3 struct{}

func getNextMarkerForS3Initialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	switch v := in.Parameters.(type) {
	case *s3.ListObjectVersionsInput:
		ctx = middleware.WithStackValue(ctx, keyMarkerKeyForS3{}, v.KeyMarker)
		ctx = middleware.WithStackValue(ctx, versionIdMarkerKeyForS3{}, v.VersionIdMarker)
	}
	return next.HandleInitialize(ctx, in)
}

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

func TestS3_deleteObjectsWithRetry(t *testing.T) {
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
		input              *s3.DeleteObjectsInput
		bucketName         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type partialOutput struct {
		deleted []types.DeletedObject
		errors  []types.Error
	}

	type want struct {
		output *partialOutput
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
				ctx: context.Background(),
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
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
				output: &partialOutput{
					deleted: []types.DeletedObject{},
					errors:  []types.Error{},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects over limit successfully",
			args: args{
				ctx: context.Background(),
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: objectsOverLimit,
						Quiet:   *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
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
				output: &partialOutput{
					deleted: []types.DeletedObject{},
					errors:  []types.Error{},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects failure",
			args: args{
				ctx: context.Background(),
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
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
				ctx: context.Background(),
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
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
				ctx: context.Background(),
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
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
				output: &partialOutput{
					deleted: []types.DeletedObject{},
					errors: []types.Error{
						{
							Key:       aws.String("Key"),
							Code:      aws.String("Code"),
							Message:   aws.String("Message"),
							VersionId: aws.String("VersionId"),
						},
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

			output, err := s3Client.deleteObjectsWithRetry(tt.args.ctx, tt.args.input, tt.args.bucketName, sleepTimeSecForS3)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if (output == nil) != (tt.want.output == nil) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
			if (output != nil) && (tt.want.output != nil) {
				got := &partialOutput{
					deleted: output.Deleted,
					errors:  output.Errors,
				}
				if !reflect.DeepEqual(got, tt.want.output) {
					t.Errorf("output = %#v, want %#v", got, tt.want.output)
				}
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
		{
			name: "list objects versions with marker successfully",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextMarker",
							getNextMarkerForS3Initialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsWithMarkerMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								keyMarker := middleware.GetStackValue(ctx, keyMarkerKeyForS3{}).(*string)
								versionIdMarker := middleware.GetStackValue(ctx, versionIdMarkerKeyForS3{}).(*string)

								var nextKeyMarker *string
								var nextVersionIdMarker *string
								var objectVersions []types.ObjectVersion
								var objectDeleteMarkers []types.DeleteMarkerEntry
								if keyMarker == nil && versionIdMarker == nil {
									nextKeyMarker = aws.String("NextMarker")
									nextVersionIdMarker = aws.String("NextMarker")
									objectVersions = []types.ObjectVersion{
										{
											Key:       aws.String("KeyForVersions1"),
											VersionId: aws.String("VersionIdForVersions1"),
										},
									}
									objectDeleteMarkers = []types.DeleteMarkerEntry{
										{
											Key:       aws.String("KeyForDeleteMarkers1"),
											VersionId: aws.String("VersionIdForDeleteMarkers1"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &s3.ListObjectVersionsOutput{
											Versions:            objectVersions,
											DeleteMarkers:       objectDeleteMarkers,
											NextKeyMarker:       nextKeyMarker,
											NextVersionIdMarker: nextVersionIdMarker,
										},
									}, middleware.Metadata{}, nil
								} else {
									objectVersions = []types.ObjectVersion{
										{
											Key:       aws.String("KeyForVersions2"),
											VersionId: aws.String("VersionIdForVersions2"),
										},
									}
									objectDeleteMarkers = []types.DeleteMarkerEntry{
										{
											Key:       aws.String("KeyForDeleteMarkers2"),
											VersionId: aws.String("VersionIdForDeleteMarkers2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &s3.ListObjectVersionsOutput{
											Versions:            objectVersions,
											DeleteMarkers:       objectDeleteMarkers,
											NextKeyMarker:       nextKeyMarker,
											NextVersionIdMarker: nextVersionIdMarker,
										},
									}, middleware.Metadata{}, nil
								}
							},
						),
						middleware.Before,
					)
					return err
				},
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForVersions1"),
						VersionId: aws.String("VersionIdForVersions1"),
					},
					{
						Key:       aws.String("KeyForDeleteMarkers1"),
						VersionId: aws.String("VersionIdForDeleteMarkers1"),
					},
					{
						Key:       aws.String("KeyForVersions2"),
						VersionId: aws.String("VersionIdForVersions2"),
					},
					{
						Key:       aws.String("KeyForDeleteMarkers2"),
						VersionId: aws.String("VersionIdForDeleteMarkers2"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions with marker failure",
			args: args{
				ctx:        context.Background(),
				bucketName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextMarker",
							getNextMarkerForS3Initialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListObjectVersionsWithMarkerErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								keyMarker := middleware.GetStackValue(ctx, keyMarkerKeyForS3{}).(*string)
								versionIdMarker := middleware.GetStackValue(ctx, versionIdMarkerKeyForS3{}).(*string)

								var nextKeyMarker *string
								var nextVersionIdMarker *string
								var objectVersions []types.ObjectVersion
								var objectDeleteMarkers []types.DeleteMarkerEntry
								if keyMarker == nil && versionIdMarker == nil {
									nextKeyMarker = aws.String("NextMarker")
									nextVersionIdMarker = aws.String("NextMarker")
									objectVersions = []types.ObjectVersion{
										{
											Key:       aws.String("KeyForVersions1"),
											VersionId: aws.String("VersionIdForVersions1"),
										},
									}
									objectDeleteMarkers = []types.DeleteMarkerEntry{
										{
											Key:       aws.String("KeyForDeleteMarkers1"),
											VersionId: aws.String("VersionIdForDeleteMarkers1"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &s3.ListObjectVersionsOutput{
											Versions:            objectVersions,
											DeleteMarkers:       objectDeleteMarkers,
											NextKeyMarker:       nextKeyMarker,
											NextVersionIdMarker: nextVersionIdMarker,
										},
									}, middleware.Metadata{}, nil
								} else {
									return middleware.FinalizeOutput{
										Result: &s3.ListObjectVersionsOutput{},
									}, middleware.Metadata{}, fmt.Errorf("ListObjectVersionsError")
								}
							},
						),
						middleware.Before,
					)
					return err
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("operation error S3: ListObjectVersions, ListObjectVersionsError"),
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
