package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-to-k/delstack/logger"
)

var _ IIamSDKClient = (*MockIamSDKClient)(nil)
var _ IIamSDKClient = (*ErrorMockIamSDKClient)(nil)
var _ IIamSDKClient = (*ApiErrorMockIamSDKClient)(nil)

var sleepTimeSecForIam = 1

/*
	Mocks for SDK Client
*/
type MockIamSDKClient struct{}

func NewMockIamSDKClient() *MockIamSDKClient {
	return &MockIamSDKClient{}
}

func (m *MockIamSDKClient) DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, nil
}

func (m *MockIamSDKClient) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	output := &iam.ListAttachedRolePoliciesOutput{
		AttachedPolicies: []types.AttachedPolicy{
			{
				PolicyArn:  aws.String("PolicyArn1"),
				PolicyName: aws.String("PolicyName1"),
			},
			{
				PolicyArn:  aws.String("PolicyArn2"),
				PolicyName: aws.String("PolicyName2"),
			},
		},
	}
	return output, nil
}

func (m *MockIamSDKClient) DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	return nil, nil
}

type ErrorMockIamSDKClient struct{}

func NewErrorMockIamSDKClient() *ErrorMockIamSDKClient {
	return &ErrorMockIamSDKClient{}
}

func (m *ErrorMockIamSDKClient) DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, fmt.Errorf("DeleteRoleError")
}

func (m *ErrorMockIamSDKClient) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *ErrorMockIamSDKClient) DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	return nil, fmt.Errorf("DetachRolePolicyError")
}

type ApiErrorMockIamSDKClient struct{}

func NewApiErrorMockIamSDKClient() *ApiErrorMockIamSDKClient {
	return &ApiErrorMockIamSDKClient{}
}

func (m *ApiErrorMockIamSDKClient) DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, fmt.Errorf("api error Throttling: Rate exceeded")
}

func (m *ApiErrorMockIamSDKClient) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return nil, fmt.Errorf("api error Throttling: Rate exceeded")
}

func (m *ApiErrorMockIamSDKClient) DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	return nil, fmt.Errorf("api error Throttling: Rate exceeded")
}

/*
	Test Cases
*/
func TestDeleteRole(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		client   IIamSDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete role successfully",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   errorMock,
			},
			want:    fmt.Errorf("DeleteRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure  for api error",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			err := iamClient.DeleteRole(tt.args.roleName, sleepTimeSecForIam)
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

func TestListAttachedRolePolicies(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		client   IIamSDKClient
	}

	type want struct {
		output []types.AttachedPolicy
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list attached role policies successfully",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   mock,
			},
			want: want{
				output: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list attached role policies failure",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   errorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("ListAttachedRolePoliciesError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			output, err := iamClient.ListAttachedRolePolicies(tt.args.roleName)
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

func TestDetachRolePolicies(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		policies []types.AttachedPolicy
		client   IIamSDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "detach role policies successfully",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "detach empty role policies successfully",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{},
				client:   mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "detach role policies failure",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				client: errorMock,
			},
			want:    fmt.Errorf("DetachRolePolicyError"),
			wantErr: true,
		},
		{
			name: "detach role policies failure for api error",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				client: apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			err := iamClient.DetachRolePolicies(tt.args.roleName, tt.args.policies, sleepTimeSecForIam)
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

func TestDetachRolePolicy(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx       context.Context
		roleName  *string
		PolicyArn *string
		client    IIamSDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete recovery point successfully",
			args: args{
				ctx:       ctx,
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				client:    mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete recovery point failure",
			args: args{
				ctx:       ctx,
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				client:    errorMock,
			},
			want:    fmt.Errorf("DetachRolePolicyError"),
			wantErr: true,
		},
		{
			name: "delete recovery point failure for api error",
			args: args{
				ctx:       ctx,
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				client:    apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			err := iamClient.DetachRolePolicy(tt.args.roleName, tt.args.PolicyArn, sleepTimeSecForIam)
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
