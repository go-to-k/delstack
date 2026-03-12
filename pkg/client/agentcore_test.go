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
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/goleak"
)

func TestAgentCoreClient_GetAgentRuntime(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		runtimeId          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    *bedrockagentcorecontrol.GetAgentRuntimeOutput
		wantErr bool
	}{
		{
			name: "get agent runtime successfully",
			args: args{
				ctx:       context.Background(),
				runtimeId: aws.String("test-runtime-abc1234567"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetAgentRuntimeMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
										AgentRuntimeId: aws.String("test-runtime-abc1234567"),
										Status:         types.AgentRuntimeStatusReady,
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				AgentRuntimeId: aws.String("test-runtime-abc1234567"),
				Status:         types.AgentRuntimeStatusReady,
			},
			wantErr: false,
		},
		{
			name: "get agent runtime failure",
			args: args{
				ctx:       context.Background(),
				runtimeId: aws.String("test-runtime-abc1234567"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetAgentRuntimeErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &bedrockagentcorecontrol.GetAgentRuntimeOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetAgentRuntimeError")
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

			sdkClient := bedrockagentcorecontrol.NewFromConfig(cfg)
			agentCoreClient := NewAgentCoreClient(sdkClient)

			got, err := agentCoreClient.GetAgentRuntime(tt.args.ctx, tt.args.runtimeId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.AgentRuntimeId == nil || *got.AgentRuntimeId != *tt.want.AgentRuntimeId {
					t.Errorf("got = %#v, want %#v", got, tt.want)
				}
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !strings.Contains(err.Error(), "GetAgentRuntimeError") {
					t.Errorf("expected ClientError with GetAgentRuntimeError, got = %#v", err)
				}
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.runtimeId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.runtimeId)
				}
			}
		})
	}
}

func TestAgentCoreClient_UpdateAgentRuntime(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		input              *bedrockagentcorecontrol.UpdateAgentRuntimeInput
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "update agent runtime successfully",
			args: args{
				ctx: context.Background(),
				input: &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
					AgentRuntimeId:       aws.String("test-runtime-abc1234567"),
					AgentRuntimeArtifact: &types.AgentRuntimeArtifactMemberContainerConfiguration{
						Value: types.ContainerConfiguration{
							ContainerUri: aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/test:latest"),
						},
					},
					RoleArn:              aws.String("arn:aws:iam::123456789012:role/test-role"),
					NetworkConfiguration: &types.NetworkConfiguration{
						NetworkMode: types.NetworkModePublic,
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateAgentRuntimeOrGetAgentRuntimeMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "UpdateAgentRuntime" {
									return middleware.FinalizeOutput{
										Result: &bedrockagentcorecontrol.UpdateAgentRuntimeOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "GetAgentRuntime" {
									return middleware.FinalizeOutput{
										Result: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
											Status: types.AgentRuntimeStatusReady,
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
			name: "update agent runtime failure",
			args: args{
				ctx: context.Background(),
				input: &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
					AgentRuntimeId:       aws.String("test-runtime-abc1234567"),
					AgentRuntimeArtifact: &types.AgentRuntimeArtifactMemberContainerConfiguration{
						Value: types.ContainerConfiguration{
							ContainerUri: aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/test:latest"),
						},
					},
					RoleArn:              aws.String("arn:aws:iam::123456789012:role/test-role"),
					NetworkConfiguration: &types.NetworkConfiguration{
						NetworkMode: types.NetworkModePublic,
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateAgentRuntimeErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &bedrockagentcorecontrol.UpdateAgentRuntimeOutput{},
								}, middleware.Metadata{}, fmt.Errorf("UpdateAgentRuntimeError")
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: true,
		},
		{
			name: "update agent runtime with failed status still succeeds",
			args: args{
				ctx: context.Background(),
				input: &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
					AgentRuntimeId:       aws.String("test-runtime-abc1234567"),
					AgentRuntimeArtifact: &types.AgentRuntimeArtifactMemberContainerConfiguration{
						Value: types.ContainerConfiguration{
							ContainerUri: aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/test:latest"),
						},
					},
					RoleArn:              aws.String("arn:aws:iam::123456789012:role/test-role"),
					NetworkConfiguration: &types.NetworkConfiguration{
						NetworkMode: types.NetworkModePublic,
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateAgentRuntimeWithFailedStatusMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "UpdateAgentRuntime" {
									return middleware.FinalizeOutput{
										Result: &bedrockagentcorecontrol.UpdateAgentRuntimeOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "GetAgentRuntime" {
									return middleware.FinalizeOutput{
										Result: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
											Status: types.AgentRuntimeStatusUpdateFailed,
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
			name: "update agent runtime with poll error",
			args: args{
				ctx: context.Background(),
				input: &bedrockagentcorecontrol.UpdateAgentRuntimeInput{
					AgentRuntimeId:       aws.String("test-runtime-abc1234567"),
					AgentRuntimeArtifact: &types.AgentRuntimeArtifactMemberContainerConfiguration{
						Value: types.ContainerConfiguration{
							ContainerUri: aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/test:latest"),
						},
					},
					RoleArn:              aws.String("arn:aws:iam::123456789012:role/test-role"),
					NetworkConfiguration: &types.NetworkConfiguration{
						NetworkMode: types.NetworkModePublic,
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateAgentRuntimePollErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "UpdateAgentRuntime" {
									return middleware.FinalizeOutput{
										Result: &bedrockagentcorecontrol.UpdateAgentRuntimeOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "GetAgentRuntime" {
									return middleware.FinalizeOutput{
										Result: &bedrockagentcorecontrol.GetAgentRuntimeOutput{},
									}, middleware.Metadata{}, fmt.Errorf("GetAgentRuntimePollError")
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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

			sdkClient := bedrockagentcorecontrol.NewFromConfig(cfg)
			agentCoreClient := NewAgentCoreClient(sdkClient)

			err = agentCoreClient.UpdateAgentRuntime(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.input.AgentRuntimeId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.input.AgentRuntimeId)
				}
			}
		})
	}
}
