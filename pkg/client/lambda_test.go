package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsMiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/goleak"
)

func TestLambdaClient_GetFunction(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		functionName       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    *lambda.GetFunctionOutput
		wantErr bool
	}{
		{
			name: "get function successfully",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFunctionMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.GetFunctionOutput{
										Configuration: &types.FunctionConfiguration{
											FunctionName: aws.String("test-function"),
										},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: &lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
				},
			},
			wantErr: false,
		},
		{
			name: "get function failure",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFunctionErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.GetFunctionOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetFunctionError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    nil,
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

			client := lambda.NewFromConfig(cfg)
			waiter := lambda.NewFunctionUpdatedV2Waiter(client)
			lambdaClient := NewLambdaClient(client, waiter)

			got, err := lambdaClient.GetFunction(tt.args.ctx, tt.args.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Configuration == nil || *got.Configuration.FunctionName != *tt.want.Configuration.FunctionName {
					t.Errorf("got = %#v, want %#v", got, tt.want)
				}
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !strings.Contains(err.Error(), "GetFunctionError") {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.functionName {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.functionName)
				}
			}
		})
	}
}

func TestLambdaClient_UpdateFunctionConfiguration(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		input              *lambda.UpdateFunctionConfigurationInput
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update function configuration successfully",
			args: args{
				ctx: context.Background(),
				input: &lambda.UpdateFunctionConfigurationInput{
					FunctionName: aws.String("test-function"),
					VpcConfig: &types.VpcConfig{
						SubnetIds:        []string{},
						SecurityGroupIds: []string{},
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateFunctionConfigurationOrGetFunctionMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "UpdateFunctionConfiguration" {
									return middleware.FinalizeOutput{
										Result: &lambda.UpdateFunctionConfigurationOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "GetFunction" {
									return middleware.FinalizeOutput{
										Result: &lambda.GetFunctionOutput{
											Configuration: &types.FunctionConfiguration{
												State:            types.StateActive,
												LastUpdateStatus: types.LastUpdateStatusSuccessful,
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
			wantErr: false,
		},
		{
			name: "update function configuration failure",
			args: args{
				ctx: context.Background(),
				input: &lambda.UpdateFunctionConfigurationInput{
					FunctionName: aws.String("test-function"),
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateFunctionConfigurationErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.UpdateFunctionConfigurationOutput{},
								}, middleware.Metadata{}, fmt.Errorf("UpdateFunctionConfigurationError")
							},
						),
						middleware.Before,
					)
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

			client := lambda.NewFromConfig(cfg)
			waiter := lambda.NewFunctionUpdatedV2Waiter(client)
			lambdaClient := NewLambdaClient(client, waiter)

			err = lambdaClient.UpdateFunctionConfiguration(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.input.FunctionName {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.input.FunctionName)
				}
			}
		})
	}
}

func TestLambdaClient_DeleteFunction(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		functionName       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete function successfully",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteFunctionMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.DeleteFunctionOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: false,
		},
		{
			name: "delete function failure",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteFunctionErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.DeleteFunctionOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteFunctionError")
							},
						),
						middleware.Before,
					)
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

			client := lambda.NewFromConfig(cfg)
			waiter := lambda.NewFunctionUpdatedV2Waiter(client)
			lambdaClient := NewLambdaClient(client, waiter)

			err = lambdaClient.DeleteFunction(tt.args.ctx, tt.args.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.functionName {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.functionName)
				}
			}
		})
	}
}

func TestLambdaClient_CheckLambdaFunctionExists(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		functionName       *string
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
			name: "check function exists successfully",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFunctionExistsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.GetFunctionOutput{
										Configuration: &types.FunctionConfiguration{
											FunctionName: aws.String("test-function"),
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
			name: "check function not exists successfully",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFunctionNotExistsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.GetFunctionOutput{},
								}, middleware.Metadata{}, fmt.Errorf("Function not found")
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
			name: "check function exists failure",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetFunctionErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &lambda.GetFunctionOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetFunctionError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test-function"),
					Err:          fmt.Errorf("operation error Lambda: GetFunction, GetFunctionError"),
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

			client := lambda.NewFromConfig(cfg)
			waiter := lambda.NewFunctionUpdatedV2Waiter(client)
			lambdaClient := NewLambdaClient(client, waiter)

			output, err := lambdaClient.CheckLambdaFunctionExists(tt.args.ctx, tt.args.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want.err)
				return
			}
			if output != tt.want.exists {
				t.Errorf("output = %#v, want %#v", output, tt.want.exists)
			}
		})
	}
}
