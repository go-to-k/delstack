package preprocessor

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestAgentCoreVPCDetacher_Preprocess(t *testing.T) {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	io.Logger = &logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		stackName *string
		resources []cfntypes.StackResourceSummary
	}

	cases := []struct {
		name    string
		args    args
		setup   func(*client.MockIAgentCore, *client.MockIEC2)
		wantErr bool
	}{
		{
			name: "no agentcore runtimes",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []cfntypes.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("test-bucket"),
					},
				},
			},
			setup:   func(m *client.MockIAgentCore, e *client.MockIEC2) {},
			wantErr: false,
		},
		{
			name: "runtime not attached to VPC",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []cfntypes.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::BedrockAgentCore::Runtime"),
						PhysicalResourceId: aws.String("test-runtime-abc1234567"),
					},
				},
			},
			setup: func(m *client.MockIAgentCore, e *client.MockIEC2) {
				m.EXPECT().GetAgentRuntime(gomock.Any(), aws.String("test-runtime-abc1234567")).Return(
					&bedrockagentcorecontrol.GetAgentRuntimeOutput{
						NetworkConfiguration: &types.NetworkConfiguration{
							NetworkMode: types.NetworkModePublic,
						},
					},
					nil,
				)
			},
			wantErr: false,
		},
		{
			name: "runtime with VPC",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []cfntypes.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::BedrockAgentCore::Runtime"),
						PhysicalResourceId: aws.String("test-runtime-abc1234567"),
					},
				},
			},
			setup: func(m *client.MockIAgentCore, e *client.MockIEC2) {
				m.EXPECT().GetAgentRuntime(gomock.Any(), aws.String("test-runtime-abc1234567")).Return(
					&bedrockagentcorecontrol.GetAgentRuntimeOutput{
						AgentRuntimeArtifact: &types.AgentRuntimeArtifactMemberContainerConfiguration{},
						RoleArn:              aws.String("arn:aws:iam::123456789012:role/test-role"),
						NetworkConfiguration: &types.NetworkConfiguration{
							NetworkMode: types.NetworkModeVpc,
							NetworkModeConfig: &types.VpcConfig{
								SecurityGroups: []string{"sg-12345"},
								Subnets:        []string{"subnet-12345"},
							},
						},
					},
					nil,
				)
				m.EXPECT().UpdateAgentRuntime(gomock.Any(), gomock.Any()).Return(nil)
				e.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "runtime with VPC and ENIs",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []cfntypes.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::BedrockAgentCore::Runtime"),
						PhysicalResourceId: aws.String("test-runtime-abc1234567"),
					},
				},
			},
			setup: func(m *client.MockIAgentCore, e *client.MockIEC2) {
				m.EXPECT().GetAgentRuntime(gomock.Any(), aws.String("test-runtime-abc1234567")).Return(
					&bedrockagentcorecontrol.GetAgentRuntimeOutput{
						AgentRuntimeArtifact: &types.AgentRuntimeArtifactMemberContainerConfiguration{},
						RoleArn:              aws.String("arn:aws:iam::123456789012:role/test-role"),
						NetworkConfiguration: &types.NetworkConfiguration{
							NetworkMode: types.NetworkModeVpc,
							NetworkModeConfig: &types.VpcConfig{
								SecurityGroups: []string{"sg-12345"},
								Subnets:        []string{"subnet-12345"},
							},
						},
					},
					nil,
				)
				m.EXPECT().UpdateAgentRuntime(gomock.Any(), gomock.Any()).Return(nil)
				e.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return(
					[]ec2types.NetworkInterface{
						{NetworkInterfaceId: aws.String("eni-111")},
						{NetworkInterfaceId: aws.String("eni-222")},
					},
					nil,
				)
				e.EXPECT().DeleteNetworkInterface(gomock.Any(), aws.String("eni-111")).Return(nil)
				e.EXPECT().DeleteNetworkInterface(gomock.Any(), aws.String("eni-222")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "get runtime error continues processing",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []cfntypes.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::BedrockAgentCore::Runtime"),
						PhysicalResourceId: aws.String("test-runtime-abc1234567"),
					},
				},
			},
			setup: func(m *client.MockIAgentCore, e *client.MockIEC2) {
				m.EXPECT().GetAgentRuntime(gomock.Any(), aws.String("test-runtime-abc1234567")).Return(
					nil,
					fmt.Errorf("runtime not found"),
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockAgentCore := client.NewMockIAgentCore(ctrl)
			mockEC2 := client.NewMockIEC2(ctrl)
			tt.setup(mockAgentCore, mockEC2)

			detacher := NewAgentCoreVPCDetacher(mockAgentCore, mockEC2)
			err := detacher.Preprocess(tt.args.ctx, tt.args.stackName, tt.args.resources)

			if (err != nil) != tt.wantErr {
				t.Errorf("Preprocess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentCoreVPCDetacher_isAttachedToVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentCore := client.NewMockIAgentCore(ctrl)
	mockEC2 := client.NewMockIEC2(ctrl)
	detacher := NewAgentCoreVPCDetacher(mockAgentCore, mockEC2)

	cases := []struct {
		name   string
		output *bedrockagentcorecontrol.GetAgentRuntimeOutput
		want   bool
	}{
		{
			name: "nil network configuration",
			output: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				NetworkConfiguration: nil,
			},
			want: false,
		},
		{
			name: "public network mode",
			output: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				NetworkConfiguration: &types.NetworkConfiguration{
					NetworkMode: types.NetworkModePublic,
				},
			},
			want: false,
		},
		{
			name: "VPC network mode",
			output: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				NetworkConfiguration: &types.NetworkConfiguration{
					NetworkMode: types.NetworkModeVpc,
				},
			},
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := detacher.isAttachedToVPC(tt.output)
			if got != tt.want {
				t.Errorf("isAttachedToVPC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentCoreVPCDetacher_getSecurityGroupIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentCore := client.NewMockIAgentCore(ctrl)
	mockEC2 := client.NewMockIEC2(ctrl)
	detacher := NewAgentCoreVPCDetacher(mockAgentCore, mockEC2)

	cases := []struct {
		name   string
		output *bedrockagentcorecontrol.GetAgentRuntimeOutput
		want   []string
	}{
		{
			name: "nil network configuration",
			output: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				NetworkConfiguration: nil,
			},
			want: nil,
		},
		{
			name: "nil network mode config",
			output: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				NetworkConfiguration: &types.NetworkConfiguration{
					NetworkMode: types.NetworkModeVpc,
				},
			},
			want: nil,
		},
		{
			name: "security groups present",
			output: &bedrockagentcorecontrol.GetAgentRuntimeOutput{
				NetworkConfiguration: &types.NetworkConfiguration{
					NetworkMode: types.NetworkModeVpc,
					NetworkModeConfig: &types.VpcConfig{
						SecurityGroups: []string{"sg-111", "sg-222"},
					},
				},
			},
			want: []string{"sg-111", "sg-222"},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := detacher.getSecurityGroupIDs(tt.output)
			if len(got) != len(tt.want) {
				t.Errorf("getSecurityGroupIDs() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("getSecurityGroupIDs()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
