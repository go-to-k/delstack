package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

/*
	Mocks for client
*/

var _ IIam = (*MockIam)(nil)
var _ IIam = (*AllErrorMockIam)(nil)
var _ IIam = (*DeleteRoleErrorMockIam)(nil)
var _ IIam = (*ListAttachedRolePoliciesErrorMockIam)(nil)
var _ IIam = (*DetachRolePoliciesErrorMockIam)(nil)
var _ IIam = (*DetachRolePoliciesErrorAfterZeroLengthMockIam)(nil)
var _ IIam = (*CheckRoleExistsErrorMockRole)(nil)
var _ IIam = (*CheckRoleNotExistsMockRole)(nil)

type MockIam struct{}

func NewMockIam() *MockIam {
	return &MockIam{}
}

func (m *MockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *MockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type AllErrorMockIam struct{}

func NewAllErrorMockIam() *AllErrorMockIam {
	return &AllErrorMockIam{}
}

func (m *AllErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *AllErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *AllErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *AllErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePolicyError")
}

func (m *AllErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, fmt.Errorf("GetRoleError")
}

type DeleteRoleErrorMockIam struct{}

func NewDeleteRoleErrorMockIam() *DeleteRoleErrorMockIam {
	return &DeleteRoleErrorMockIam{}
}

func (m *DeleteRoleErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *DeleteRoleErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *DeleteRoleErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *DeleteRoleErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DeleteRoleErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type ListAttachedRolePoliciesErrorMockIam struct{}

func NewListAttachedRolePoliciesErrorMockIam() *ListAttachedRolePoliciesErrorMockIam {
	return &ListAttachedRolePoliciesErrorMockIam{}
}

func (m *ListAttachedRolePoliciesErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *ListAttachedRolePoliciesErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type DetachRolePoliciesErrorMockIam struct{}

func NewDetachRolePoliciesErrorMockIam() *DetachRolePoliciesErrorMockIam {
	return &DetachRolePoliciesErrorMockIam{}
}

func (m *DetachRolePoliciesErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *DetachRolePoliciesErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *DetachRolePoliciesErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type DetachRolePoliciesErrorAfterZeroLengthMockIam struct{}

func NewDetachRolePoliciesErrorAfterZeroLengthMockIam() *DetachRolePoliciesErrorAfterZeroLengthMockIam {
	return &DetachRolePoliciesErrorAfterZeroLengthMockIam{}
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{}
	return output, nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesErrorAfterZeroLength")
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type CheckRoleExistsErrorMockRole struct{}

func NewCheckRoleExistsErrorMockRole() *CheckRoleExistsErrorMockRole {
	return &CheckRoleExistsErrorMockRole{}
}

func (m *CheckRoleExistsErrorMockRole) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *CheckRoleExistsErrorMockRole) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, fmt.Errorf("GetRoleError")
}

type CheckRoleNotExistsMockRole struct{}

func NewCheckRoleNotExistsMockRole() *CheckRoleNotExistsMockRole {
	return &CheckRoleNotExistsMockRole{}
}

func (m *CheckRoleNotExistsMockRole) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *CheckRoleNotExistsMockRole) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, nil
}
