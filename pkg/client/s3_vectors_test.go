package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/aws/smithy-go/middleware"
)

/*
	Test Cases
*/

func TestS3Vectors_DeleteVectorBucket(t *testing.T) {
	type args struct {
		ctx                context.Context
		vectorBucketName   *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete vector bucket successfully",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteVectorBucketMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3vectors.DeleteVectorBucketOutput{},
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
			name: "delete vector bucket failure",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteVectorBucketErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("DeleteVectorBucketError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test-vector-bucket"),
				Err:          fmt.Errorf("operation error S3Vectors: DeleteVectorBucket, DeleteVectorBucketError"),
			},
			wantErr: true,
		},
		{
			name: "delete vector bucket failure for api error SlowDown",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteVectorBucketApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: nil,
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxRetryCount,
										Err:     fmt.Errorf("api error SlowDown"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test-vector-bucket"),
				Err:          fmt.Errorf("operation error S3Vectors: DeleteVectorBucket, exceeded maximum number of attempts, 10, api error SlowDown"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := s3vectors.NewFromConfig(cfg)
			s3VectorsClient := NewS3Vectors(client)

			err = s3VectorsClient.DeleteVectorBucket(tt.args.ctx, tt.args.vectorBucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
			}
		})
	}
}

func TestS3Vectors_DeleteIndex(t *testing.T) {
	type args struct {
		ctx                context.Context
		indexName          *string
		vectorBucketName   *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete index successfully",
			args: args{
				ctx:              context.Background(),
				indexName:        aws.String("test-index"),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteIndexMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3vectors.DeleteIndexOutput{},
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
			name: "delete index failure",
			args: args{
				ctx:              context.Background(),
				indexName:        aws.String("test-index"),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteIndexErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("DeleteIndexError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test-vector-bucket/test-index"),
				Err:          fmt.Errorf("operation error S3Vectors: DeleteIndex, DeleteIndexError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := s3vectors.NewFromConfig(cfg)
			s3VectorsClient := NewS3Vectors(client)

			err = s3VectorsClient.DeleteIndex(tt.args.ctx, tt.args.indexName, tt.args.vectorBucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
			}
		})
	}
}

func TestS3Vectors_ListIndexesByPage(t *testing.T) {
	type args struct {
		ctx                context.Context
		vectorBucketName   *string
		nextToken          *string
		keyPrefix          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output *ListIndexesByPageOutput
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list indexes successfully",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				nextToken:        nil,
				keyPrefix:        nil,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListIndexesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3vectors.ListIndexesOutput{
										Indexes: []types.IndexSummary{
											{
												IndexName: aws.String("index1"),
											},
											{
												IndexName: aws.String("index2"),
											},
										},
										NextToken: aws.String("token1"),
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: &ListIndexesByPageOutput{
					Indexes: []types.IndexSummary{
						{
							IndexName: aws.String("index1"),
						},
						{
							IndexName: aws.String("index2"),
						},
					},
					NextToken: aws.String("token1"),
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list indexes failure",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				nextToken:        nil,
				keyPrefix:        nil,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListIndexesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("ListIndexesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err: &ClientError{
					ResourceName: aws.String("test-vector-bucket"),
					Err:          fmt.Errorf("operation error S3Vectors: ListIndexes, ListIndexesError"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := s3vectors.NewFromConfig(cfg)
			s3VectorsClient := NewS3Vectors(client)

			output, err := s3VectorsClient.ListIndexesByPage(tt.args.ctx, tt.args.vectorBucketName, tt.args.nextToken, tt.args.keyPrefix)
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

func TestS3Vectors_CheckVectorBucketExists(t *testing.T) {
	type args struct {
		ctx                context.Context
		vectorBucketName   *string
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
			name: "check vector bucket exists",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListVectorBucketsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3vectors.ListVectorBucketsOutput{
										VectorBuckets: []types.VectorBucketSummary{
											{
												VectorBucketName: aws.String("test-vector-bucket"),
											},
											{
												VectorBucketName: aws.String("test-vector-bucket2"),
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
			name: "check vector bucket do not exist",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListVectorBucketsNotExistMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3vectors.ListVectorBucketsOutput{
										VectorBuckets: []types.VectorBucketSummary{
											{
												VectorBucketName: aws.String("test-vector-bucket0"),
											},
											{
												VectorBucketName: aws.String("test-vector-bucket1"),
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
			name: "check vector bucket exists failure",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListVectorBucketsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("ListVectorBucketsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test-vector-bucket"),
					Err:          fmt.Errorf("operation error S3Vectors: ListVectorBuckets, ListVectorBucketsError"),
				},
			},
			wantErr: true,
		},
		{
			name: "check vector bucket exists failure for api error SlowDown",
			args: args{
				ctx:              context.Background(),
				vectorBucketName: aws.String("test-vector-bucket"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListVectorBucketsApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: nil,
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxRetryCount,
										Err:     fmt.Errorf("api error SlowDown"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test-vector-bucket"),
					Err:          fmt.Errorf("operation error S3Vectors: ListVectorBuckets, exceeded maximum number of attempts, 10, api error SlowDown"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := s3vectors.NewFromConfig(cfg)
			s3VectorsClient := NewS3Vectors(client)

			output, err := s3VectorsClient.CheckVectorBucketExists(tt.args.ctx, tt.args.vectorBucketName)
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
