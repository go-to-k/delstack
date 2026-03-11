package preprocessor

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestLambdaVPCDetacher_Preprocess(t *testing.T) {
	// Initialize logger for tests
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	io.Logger = &logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		stackName *string
		resources []types.StackResourceSummary
	}

	cases := []struct {
		name    string
		args    args
		setup   func(*client.MockILambda, *client.MockICloudFormation)
		wantErr bool
	}{
		{
			name: "no lambda functions",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("test-bucket"),
					},
				},
			},
			setup:   func(m *client.MockILambda, c *client.MockICloudFormation) {},
			wantErr: false,
		},
		{
			name: "lambda function not attached to VPC",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function"),
					},
				},
			},
			setup: func(m *client.MockILambda, c *client.MockICloudFormation) {
				m.EXPECT().GetFunction(gomock.Any(), aws.String("test-function")).Return(
					&lambda.GetFunctionOutput{
						Configuration: &lambdatypes.FunctionConfiguration{
							VpcConfig: &lambdatypes.VpcConfigResponse{
								VpcId: nil,
							},
						},
					},
					nil,
				)
			},
			wantErr: false,
		},
		{
			name: "lambda function with VPC (no IPv6)",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function"),
					},
				},
			},
			setup: func(m *client.MockILambda, c *client.MockICloudFormation) {
				m.EXPECT().GetFunction(gomock.Any(), aws.String("test-function")).Return(
					&lambda.GetFunctionOutput{
						Configuration: &lambdatypes.FunctionConfiguration{
							VpcConfig: &lambdatypes.VpcConfigResponse{
								VpcId:                   aws.String("vpc-12345"),
								Ipv6AllowedForDualStack: aws.Bool(false),
							},
						},
					},
					nil,
				)
				m.EXPECT().UpdateFunctionConfiguration(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "lambda function with VPC and IPv6",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function"),
					},
				},
			},
			setup: func(m *client.MockILambda, c *client.MockICloudFormation) {
				m.EXPECT().GetFunction(gomock.Any(), aws.String("test-function")).Return(
					&lambda.GetFunctionOutput{
						Configuration: &lambdatypes.FunctionConfiguration{
							VpcConfig: &lambdatypes.VpcConfigResponse{
								VpcId:                   aws.String("vpc-12345"),
								Ipv6AllowedForDualStack: aws.Bool(true),
							},
						},
					},
					nil,
				)
				m.EXPECT().UpdateFunctionConfiguration(gomock.Any(), gomock.Any()).Return(nil).Times(2)
			},
			wantErr: false,
		},
		{
			name: "get function error continues processing",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						PhysicalResourceId: aws.String("test-function"),
					},
				},
			},
			setup: func(m *client.MockILambda, c *client.MockICloudFormation) {
				m.EXPECT().GetFunction(gomock.Any(), aws.String("test-function")).Return(
					nil,
					fmt.Errorf("function not found"),
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockLambda := client.NewMockILambda(ctrl)
			mockCfn := client.NewMockICloudFormation(ctrl)
			tt.setup(mockLambda, mockCfn)

			detacher := NewLambdaVPCDetacher(mockLambda, mockCfn)
			err := detacher.Preprocess(tt.args.ctx, tt.args.stackName, tt.args.resources)

			if (err != nil) != tt.wantErr {
				t.Errorf("Preprocess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLambdaVPCDetacher_isAttachedToVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLambda := client.NewMockILambda(ctrl)
	mockCfn := client.NewMockICloudFormation(ctrl)
	detacher := NewLambdaVPCDetacher(mockLambda, mockCfn)

	cases := []struct {
		name   string
		output *lambda.GetFunctionOutput
		want   bool
	}{
		{
			name: "nil configuration",
			output: &lambda.GetFunctionOutput{
				Configuration: nil,
			},
			want: false,
		},
		{
			name: "nil VPC config",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: nil,
				},
			},
			want: false,
		},
		{
			name: "nil VPC ID",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: &lambdatypes.VpcConfigResponse{
						VpcId: nil,
					},
				},
			},
			want: false,
		},
		{
			name: "empty VPC ID",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: &lambdatypes.VpcConfigResponse{
						VpcId: aws.String(""),
					},
				},
			},
			want: false,
		},
		{
			name: "valid VPC ID",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: &lambdatypes.VpcConfigResponse{
						VpcId: aws.String("vpc-12345"),
					},
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

func TestLambdaVPCDetacher_isIPv6Enabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLambda := client.NewMockILambda(ctrl)
	mockCfn := client.NewMockICloudFormation(ctrl)
	detacher := NewLambdaVPCDetacher(mockLambda, mockCfn)

	cases := []struct {
		name   string
		output *lambda.GetFunctionOutput
		want   bool
	}{
		{
			name: "nil configuration",
			output: &lambda.GetFunctionOutput{
				Configuration: nil,
			},
			want: false,
		},
		{
			name: "nil VPC config",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: nil,
				},
			},
			want: false,
		},
		{
			name: "nil Ipv6AllowedForDualStack",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: &lambdatypes.VpcConfigResponse{
						Ipv6AllowedForDualStack: nil,
					},
				},
			},
			want: false,
		},
		{
			name: "IPv6 disabled",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: &lambdatypes.VpcConfigResponse{
						Ipv6AllowedForDualStack: aws.Bool(false),
					},
				},
			},
			want: false,
		},
		{
			name: "IPv6 enabled",
			output: &lambda.GetFunctionOutput{
				Configuration: &lambdatypes.FunctionConfiguration{
					VpcConfig: &lambdatypes.VpcConfigResponse{
						Ipv6AllowedForDualStack: aws.Bool(true),
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := detacher.isIPv6Enabled(tt.output)
			if got != tt.want {
				t.Errorf("isIPv6Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
