package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsMiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go/middleware"
)

/*
	Test Cases
*/

func TestCloudFormation_DeleteStack(t *testing.T) {
	type args struct {
		ctx                context.Context
		stackName          *string
		retainResources    []string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete stack successfully",
			args: args{
				ctx:             context.Background(),
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteStackOrDescribeStacksForWaiterMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "DeleteStack" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DeleteStackOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DescribeStacks" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DescribeStacksOutput{
											Stacks: []types.Stack{
												{
													StackName:   aws.String("StackName"),
													StackStatus: "DELETE_COMPLETE",
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete stack successfully including non retainResources",
			args: args{
				ctx:             context.Background(),
				stackName:       aws.String("test"),
				retainResources: []string{},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteStackOrDescribeStacksForWaiterIncludingNonRetainResourcesMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "DeleteStack" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DeleteStackOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DescribeStacks" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DescribeStacksOutput{
											Stacks: []types.Stack{
												{
													StackName:   aws.String("StackName"),
													StackStatus: "DELETE_COMPLETE",
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete stack successfully for transitioned to Failure",
			args: args{
				ctx:             context.Background(),
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteStackOrDescribeStacksForWaiterStateTransitionedToFailureMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "DeleteStack" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DeleteStackOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DescribeStacks" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DescribeStacksOutput{},
									}, middleware.Metadata{}, fmt.Errorf("waiter state transitioned to Failure")
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete stack failure",
			args: args{
				ctx:             context.Background(),
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteStackErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DeleteStackOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteStackError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error CloudFormation: DeleteStack, DeleteStackError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for wait errors",
			args: args{
				ctx:             context.Background(),
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteStackOrDescribeStacksForWaiterErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "DeleteStack" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DeleteStackOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DescribeStacks" {
									return middleware.FinalizeOutput{
										Result: &cloudformation.DescribeStacksOutput{},
									}, middleware.Metadata{}, fmt.Errorf("WaitError")
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("expected err to be of type smithy.APIError, got %w", fmt.Errorf("operation error CloudFormation: DescribeStacks, WaitError")),
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

			client := cloudformation.NewFromConfig(cfg)
			cfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(client)
			cfnClient := NewCloudFormation(
				client,
				cfnWaiter,
			)

			err = cfnClient.DeleteStack(tt.args.ctx, tt.args.stackName, tt.args.retainResources)
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

func TestCloudFormation_DescribeStacks(t *testing.T) {
	type args struct {
		ctx                context.Context
		stackName          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output *cloudformation.DescribeStacksOutput
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
			name: "describe stacks successfully",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeStacksMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackName:   aws.String("StackName"),
												StackStatus: "DELETE_FAILED",
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
				output: &cloudformation.DescribeStacksOutput{
					Stacks: []types.Stack{
						{
							StackName:   aws.String("StackName"),
							StackStatus: "DELETE_FAILED",
						},
					},
				},
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "describe stacks failure",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeStacksErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DescribeStacksOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeStacksError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: &cloudformation.DescribeStacksOutput{},
				exists: true,
				err:    fmt.Errorf("operation error CloudFormation: DescribeStacks, DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "describe stacks but not exist",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeStacksMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DescribeStacksOutput{},
								}, middleware.Metadata{}, fmt.Errorf("does not exist")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: &cloudformation.DescribeStacksOutput{},
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
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := cloudformation.NewFromConfig(cfg)
			cfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(client)
			cfnClient := NewCloudFormation(
				client,
				cfnWaiter,
			)

			output, exists, err := cfnClient.DescribeStacks(tt.args.ctx, tt.args.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
				return
			}
			if !tt.wantErr && exists != tt.want.exists {
				t.Errorf("exists = %#v, want %#v", exists, tt.want.exists)
			}
			if !tt.wantErr && exists && !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
				return
			}
		})
	}
}

func TestCloudFormation_waitDeleteStack(t *testing.T) {
	type args struct {
		ctx                context.Context
		stackName          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "wait successfully",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeStacksForWaiterMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackName:   aws.String("StackName"),
												StackStatus: "DELETE_COMPLETE",
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
			want:    nil,
			wantErr: false,
		},
		{
			name: "wait failure for wait error",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeStacksForWaiterErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DescribeStacksOutput{},
								}, middleware.Metadata{}, fmt.Errorf("WaitError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("expected err to be of type smithy.APIError, got %w", fmt.Errorf("operation error CloudFormation: DescribeStacks, WaitError")),
			wantErr: true,
		},
		{
			name: "wait failure for transitioned to Failure",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeStacksForWaiterStateTransitionedToFailureMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.DescribeStacksOutput{},
								}, middleware.Metadata{}, fmt.Errorf("waiter state transitioned to Failure")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    nil,
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

			client := cloudformation.NewFromConfig(cfg)
			cfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(client)
			cfnClient := NewCloudFormation(
				client,
				cfnWaiter,
			)

			err = cfnClient.waitDeleteStack(tt.args.ctx, tt.args.stackName)
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

func TestCloudFormation_ListStackResources(t *testing.T) {
	type args struct {
		ctx                context.Context
		stackName          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output []types.StackResourceSummary
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list stack resources successfully",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListStackResourcesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.ListStackResourcesOutput{
										StackResourceSummaries: []types.StackResourceSummary{
											{
												LogicalResourceId:  aws.String("LogicalResourceId1"),
												ResourceStatus:     "DELETE_FAILED",
												ResourceType:       aws.String("AWS::CloudFormation::Stack"),
												PhysicalResourceId: aws.String("PhysicalResourceId1"),
											},
											{
												LogicalResourceId:  aws.String("LogicalResourceId2"),
												ResourceStatus:     "DELETE_FAILED",
												ResourceType:       aws.String("AWS::S3::Bucket"),
												PhysicalResourceId: aws.String("PhysicalResourceId2"),
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
				output: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stack resources failure",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListStackResourcesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudformation.ListStackResourcesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListStackResourcesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.StackResourceSummary{},
				err:    fmt.Errorf("operation error CloudFormation: ListStackResources, ListStackResourcesError"),
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

			client := cloudformation.NewFromConfig(cfg)
			cfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(client)
			cfnClient := NewCloudFormation(
				client,
				cfnWaiter,
			)

			output, err := cfnClient.ListStackResources(tt.args.ctx, tt.args.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}

// func TestCloudFormation_ListStacks(t *testing.T) {
// 	type args struct {
// 		ctx                context.Context
// 		withAPIOptionsFunc func(*middleware.Stack) error
// 	}

// 	type want struct {
// 		output []types.StackSummary
// 		err    error
// 	}

// 	cases := []struct {
// 		name    string
// 		args    args
// 		want    want
// 		wantErr bool
// 	}{
// 		{
// 			name: "list stacks successfully",
// 			args: args{
// 				ctx: context.Background(),
// 				withAPIOptionsFunc: func(stack *middleware.Stack) error {
// 					return stack.Finalize.Add(
// 						middleware.FinalizeMiddlewareFunc(
// 							"ListStacksMock",
// 							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
// 								return middleware.FinalizeOutput{
// 									Result: &cloudformation.ListStacksOutput{
// 										StackSummaries: []types.StackSummary{
// 											{
// 												StackName:   aws.String("TestStack1"),
// 												StackStatus: types.StackStatusCreateComplete,
// 											},
// 											{
// 												StackName:   aws.String("TestStack2"),
// 												StackStatus: types.StackStatusCreateComplete,
// 											},
// 										},
// 									},
// 								}, middleware.Metadata{}, nil
// 							},
// 						),
// 						middleware.Before,
// 					)
// 				},
// 			},
// 			want: want{
// 				output: []types.StackSummary{
// 					{
// 						StackName:   aws.String("TestStack1"),
// 						StackStatus: types.StackStatusCreateComplete,
// 					},
// 					{
// 						StackName:   aws.String("TestStack2"),
// 						StackStatus: types.StackStatusCreateComplete,
// 					},
// 				},
// 				err: nil,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "list stacks but empty successfully",
// 			args: args{
// 				ctx: context.Background(),
// 				withAPIOptionsFunc: func(stack *middleware.Stack) error {
// 					return stack.Finalize.Add(
// 						middleware.FinalizeMiddlewareFunc(
// 							"ListStacksEmptyMock",
// 							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
// 								return middleware.FinalizeOutput{
// 									Result: &cloudformation.ListStacksOutput{
// 										StackSummaries: []types.StackSummary{},
// 									},
// 								}, middleware.Metadata{}, nil
// 							},
// 						),
// 						middleware.Before,
// 					)
// 				},
// 			},
// 			want: want{
// 				output: []types.StackSummary{},
// 				err:    nil,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "list stacks failure",
// 			args: args{
// 				ctx: context.Background(),
// 				withAPIOptionsFunc: func(stack *middleware.Stack) error {
// 					return stack.Finalize.Add(
// 						middleware.FinalizeMiddlewareFunc(
// 							"ListStacksErrorMock",
// 							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
// 								return middleware.FinalizeOutput{
// 									Result: &cloudformation.ListStacksOutput{},
// 								}, middleware.Metadata{}, fmt.Errorf("ListStacksError")
// 							},
// 						),
// 						middleware.Before,
// 					)
// 				},
// 			},
// 			want: want{
// 				output: []types.StackSummary{},
// 				err:    fmt.Errorf("operation error CloudFormation: ListStacks, DeleteRoleError"),
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range cases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cfg, err := config.LoadDefaultConfig(
// 				tt.args.ctx,
// 				config.WithRegion("ap-northeast-1"),
// 				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
// 			)
// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			client := cloudformation.NewFromConfig(cfg)
// 			cfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(client)
// 			cfnClient := NewCloudFormation(
// 				client,
// 				cfnWaiter,
// 			)

// 			output, err := cfnClient.ListStacks(tt.args.ctx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.wantErr && err.Error() != tt.want.err.Error() {
// 				t.Errorf("err = %#v, want %#v", err, tt.want)
// 				return
// 			}
// 			if !reflect.DeepEqual(output, tt.want.output) {
// 				t.Errorf("output = %#v, want %#v", output, tt.want.output)
// 			}
// 		})
// 	}
// }
