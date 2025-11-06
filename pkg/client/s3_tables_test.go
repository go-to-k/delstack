package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/aws/smithy-go/middleware"
)

/*
	Test Cases
*/

func TestS3Tables_DeleteTableBucket(t *testing.T) {
	type args struct {
		ctx                context.Context
		tableBucketARN     *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete table bucket successfully",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteTableBucketMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.DeleteTableBucketOutput{},
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
			name: "delete table bucket failure",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteTableBucketErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("DeleteTableBucketError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				Err:          fmt.Errorf("operation error S3Tables: DeleteTableBucket, DeleteTableBucketError"),
			},
			wantErr: true,
		},
		{
			name: "delete table bucket failure for api error SlowDown",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteTableBucketApiErrorMock",
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
				ResourceName: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				Err:          fmt.Errorf("operation error S3Tables: DeleteTableBucket, exceeded maximum number of attempts, 10, api error SlowDown"),
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			err = s3TablesClient.DeleteTableBucket(tt.args.ctx, tt.args.tableBucketARN)
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

func TestS3Tables_DeleteNamespace(t *testing.T) {
	type args struct {
		ctx                context.Context
		namespace          *string
		tableBucketARN     *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete namespace successfully",
			args: args{
				ctx:            context.Background(),
				namespace:      aws.String("namespace1"),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteNamespaceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.DeleteNamespaceOutput{},
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
			name: "delete namespace failure",
			args: args{
				ctx:            context.Background(),
				namespace:      aws.String("namespace1"),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteNamespaceErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("DeleteNamespaceError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test/namespace1"),
				Err:          fmt.Errorf("operation error S3Tables: DeleteNamespace, DeleteNamespaceError"),
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			err = s3TablesClient.DeleteNamespace(tt.args.ctx, tt.args.namespace, tt.args.tableBucketARN)
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

func TestS3Tables_DeleteTable(t *testing.T) {
	type args struct {
		ctx                context.Context
		tableName          *string
		namespace          *string
		tableBucketARN     *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete table successfully",
			args: args{
				ctx:            context.Background(),
				tableName:      aws.String("table1"),
				namespace:      aws.String("namespace1"),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteTableMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.DeleteTableOutput{},
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
			name: "delete table failure",
			args: args{
				ctx:            context.Background(),
				tableName:      aws.String("table1"),
				namespace:      aws.String("namespace1"),
				tableBucketARN: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteTableErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("DeleteTableError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test/namespace1/table1"),
				Err:          fmt.Errorf("operation error S3Tables: DeleteTable, DeleteTableError"),
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			err = s3TablesClient.DeleteTable(tt.args.ctx, tt.args.tableName, tt.args.namespace, tt.args.tableBucketARN)
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

func TestS3Tables_ListNamespacesByPage(t *testing.T) {
	type args struct {
		ctx                context.Context
		tableBucketARN     *string
		continuationToken  *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output *ListNamespacesByPageOutput
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list namespaces successfully",
			args: args{
				ctx:               context.Background(),
				tableBucketARN:    aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				continuationToken: nil,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListNamespacesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.ListNamespacesOutput{
										Namespaces: []types.NamespaceSummary{
											{
												Namespace: []string{"namespace1", "namespace2"},
											},
										},
										ContinuationToken: aws.String("token1"),
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: &ListNamespacesByPageOutput{
					Namespaces: []types.NamespaceSummary{
						{
							Namespace: []string{"namespace1", "namespace2"},
						},
					},
					ContinuationToken: aws.String("token1"),
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list namespaces failure",
			args: args{
				ctx:               context.Background(),
				tableBucketARN:    aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				continuationToken: nil,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListNamespacesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("ListNamespacesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err: &ClientError{
					ResourceName: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
					Err:          fmt.Errorf("operation error S3Tables: ListNamespaces, ListNamespacesError"),
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			output, err := s3TablesClient.ListNamespacesByPage(tt.args.ctx, tt.args.tableBucketARN, tt.args.continuationToken)
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

func TestS3Tables_ListTablesByPage(t *testing.T) {
	type args struct {
		ctx                context.Context
		tableBucketARN     *string
		namespace          *string
		continuationToken  *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output *ListTablesByPageOutput
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list tables successfully",
			args: args{
				ctx:               context.Background(),
				tableBucketARN:    aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				namespace:         aws.String("namespace1"),
				continuationToken: nil,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListTablesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.ListTablesOutput{
										Tables: []types.TableSummary{
											{
												Name: aws.String("table1"),
											},
											{
												Name: aws.String("table2"),
											},
										},
										ContinuationToken: aws.String("token1"),
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: &ListTablesByPageOutput{
					Tables: []types.TableSummary{
						{
							Name: aws.String("table1"),
						},
						{
							Name: aws.String("table2"),
						},
					},
					ContinuationToken: aws.String("token1"),
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list tables failure",
			args: args{
				ctx:               context.Background(),
				tableBucketARN:    aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test"),
				namespace:         aws.String("namespace1"),
				continuationToken: nil,
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListTablesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("ListTablesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err: &ClientError{
					ResourceName: aws.String("arn:aws:s3:us-east-1:123456789012:table-bucket/test/namespace1"),
					Err:          fmt.Errorf("operation error S3Tables: ListTables, ListTablesError"),
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			output, err := s3TablesClient.ListTablesByPage(tt.args.ctx, tt.args.tableBucketARN, tt.args.namespace, tt.args.continuationToken)
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

func TestS3Tables_CheckTableBucketExists(t *testing.T) {
	type args struct {
		ctx                context.Context
		tableBucketARN     *string
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
			name: "check table bucket exists",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListTableBucketsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.ListTableBucketsOutput{
										TableBuckets: []types.TableBucketSummary{
											{
												Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
											},
											{
												Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test2"),
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
			name: "check table bucket do not exist",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListTableBucketsNotExistMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.ListTableBucketsOutput{
										TableBuckets: []types.TableBucketSummary{
											{
												Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test0"),
											},
											{
												Arn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test1"),
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
			name: "check table bucket exists failure",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListTableBucketsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: nil,
								}, middleware.Metadata{}, fmt.Errorf("ListTableBucketsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
					Err:          fmt.Errorf("operation error S3Tables: ListTableBuckets, ListTableBucketsError"),
				},
			},
			wantErr: true,
		},
		{
			name: "check table bucket exists failure for api error SlowDown",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListTableBucketsApiErrorMock",
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
					ResourceName: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
					Err:          fmt.Errorf("operation error S3Tables: ListTableBuckets, exceeded maximum number of attempts, 10, api error SlowDown"),
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			output, err := s3TablesClient.CheckTableBucketExists(tt.args.ctx, tt.args.tableBucketARN)
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

func TestS3Tables_CheckNamespaceExists(t *testing.T) {
	type args struct {
		ctx                context.Context
		tableBucketARN     *string
		namespace          *string
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
			name: "namespace exists",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
				namespace:      aws.String("test-namespace"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListNamespacesWithNamespaceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.ListNamespacesOutput{
										Namespaces: []types.NamespaceSummary{
											{
												Namespace: []string{"test-namespace"},
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
			name: "namespace does not exist",
			args: args{
				ctx:            context.Background(),
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
				namespace:      aws.String("test-namespace"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListNamespacesWithoutNamespaceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &s3tables.ListNamespacesOutput{
										Namespaces: []types.NamespaceSummary{},
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

			client := s3tables.NewFromConfig(cfg)
			s3TablesClient := NewS3Tables(client)

			output, err := s3TablesClient.CheckNamespaceExists(tt.args.ctx, tt.args.tableBucketARN, tt.args.namespace)
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

func TestParseS3TablesNamespaceArn(t *testing.T) {
	type args struct {
		namespaceArn *string
	}

	type want struct {
		tableBucketARN *string
		namespace      *string
		err            error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "parse namespace ARN successfully",
			args: args{
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test-bucket/namespace/test-namespace"),
			},
			want: want{
				tableBucketARN: aws.String("arn:aws:s3tables:us-east-1:123456789012:test-bucket"),
				namespace:      aws.String("test-namespace"),
				err:            nil,
			},
			wantErr: false,
		},
		{
			name: "parse namespace ARN with nil",
			args: args{
				namespaceArn: nil,
			},
			want: want{
				tableBucketARN: nil,
				namespace:      nil,
				err:            fmt.Errorf("namespace ARN is nil"),
			},
			wantErr: true,
		},
		{
			name: "parse namespace ARN with invalid format",
			args: args{
				namespaceArn: aws.String("invalid-arn"),
			},
			want: want{
				tableBucketARN: nil,
				namespace:      nil,
				err:            fmt.Errorf("invalid namespace ARN format: invalid-arn"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			tableBucketARN, namespace, err := ParseS3TablesNamespaceArn(tt.args.namespaceArn)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(tableBucketARN, tt.want.tableBucketARN) {
					t.Errorf("tableBucketARN = %#v, want %#v", tableBucketARN, tt.want.tableBucketARN)
				}
				if !reflect.DeepEqual(namespace, tt.want.namespace) {
					t.Errorf("namespace = %#v, want %#v", namespace, tt.want.namespace)
				}
			}
		})
	}
}
