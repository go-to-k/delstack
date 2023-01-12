package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

var _ IIamSDKClient = (*MockIamSDKClient)(nil)
var _ IIamSDKClient = (*ErrorMockIamSDKClient)(nil)
var _ IIamSDKClient = (*ApiErrorMockIamSDKClient)(nil)
var _ IIamSDKClient = (*NotExistsMockForGetRoleIamSDKClient)(nil)

const sleepTimeSecForIam = 1

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

func (m *MockIamSDKClient) GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	output := &iam.GetRoleOutput{
		Role: &types.Role{
			RoleName: aws.String("RoleName"),
		},
	}
	return output, nil
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

func (m *ErrorMockIamSDKClient) GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return nil, fmt.Errorf("GetRoleError")
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

func (m *ApiErrorMockIamSDKClient) GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return nil, fmt.Errorf("api error Throttling: Rate exceeded")
}

type NotExistsMockForGetRoleIamSDKClient struct{}

func NewNotExistsMockForGetRoleIamSDKClient() *NotExistsMockForGetRoleIamSDKClient {
	return &NotExistsMockForGetRoleIamSDKClient{}
}

func (m *NotExistsMockForGetRoleIamSDKClient) DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return nil, nil
}

func (m *NotExistsMockForGetRoleIamSDKClient) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
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

func (m *NotExistsMockForGetRoleIamSDKClient) DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	return nil, nil
}

func (m *NotExistsMockForGetRoleIamSDKClient) GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return nil, fmt.Errorf("NoSuchEntity")
}
