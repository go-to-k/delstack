package operation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestLambdaFunctionOperator_DeleteLambdaFunction(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx          context.Context
		functionName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockILambda)
		want          error
		wantErr       bool
	}{
		{
			name: "delete lambda function successfully",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-function")).Return(true, nil)
				m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-function")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete lambda function failure",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-function")).Return(true, nil)
				m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-function")).Return(fmt.Errorf("DeleteFunctionError"))
			},
			want:    fmt.Errorf("DeleteFunctionError"),
			wantErr: true,
		},
		{
			name: "delete lambda function failure for check lambda function exists errors",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-function")).Return(false, fmt.Errorf("GetFunctionError"))
			},
			want:    fmt.Errorf("GetFunctionError"),
			wantErr: true,
		},
		{
			name: "delete lambda function successfully for function not exists",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-function")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete lambda@edge function successfully after retry",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-edge-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-edge-function")).Return(true, nil)
				gomock.InOrder(
					m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-edge-function")).Return(fmt.Errorf("Lambda was unable to delete because it is a replicated function")),
					m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-edge-function")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete lambda@edge function failure with non-replica error during retry",
			args: args{
				ctx:          context.Background(),
				functionName: aws.String("test-edge-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-edge-function")).Return(true, nil)
				gomock.InOrder(
					m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-edge-function")).Return(fmt.Errorf("Lambda was unable to delete because it is a replicated function")),
					m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-edge-function")).Return(fmt.Errorf("SomeOtherError")),
				)
			},
			want:    fmt.Errorf("SomeOtherError"),
			wantErr: true,
		},
		{
			name: "delete lambda@edge function failure with context cancelled during retry",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				functionName: aws.String("test-edge-function"),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("test-edge-function")).Return(true, nil)
				m.EXPECT().DeleteFunction(gomock.Any(), aws.String("test-edge-function")).Return(fmt.Errorf("Lambda was unable to delete because it is a replicated function"))
			},
			want:    fmt.Errorf("[resource test-edge-function] context canceled"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			lambdaMock := client.NewMockILambda(ctrl)
			tt.prepareMockFn(lambdaMock)

			lambdaFunctionOperator := NewLambdaFunctionOperator(lambdaMock)
			lambdaFunctionOperator.retryInterval = 10 * time.Millisecond
			lambdaFunctionOperator.retryTimeout = 1 * time.Second

			err := lambdaFunctionOperator.DeleteLambdaFunction(tt.args.ctx, tt.args.functionName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestLambdaFunctionOperator_DeleteResourcesForLambdaFunction(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockILambda)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().DeleteFunction(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockILambda) {
				m.EXPECT().CheckLambdaFunctionExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("GetFunctionError"))
			},
			want:    fmt.Errorf("GetFunctionError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			lambdaMock := client.NewMockILambda(ctrl)
			tt.prepareMockFn(lambdaMock)

			lambdaFunctionOperator := NewLambdaFunctionOperator(lambdaMock)
			lambdaFunctionOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::Lambda::Function"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := lambdaFunctionOperator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
