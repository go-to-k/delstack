package client

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type IIam interface {
	DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error
	ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error)
	DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error
	DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error
	CheckRoleExists(ctx context.Context, roleName *string) (bool, error)
}

var _ IIam = (*Iam)(nil)

type IIamSDKClient interface {
	DeleteRole(ctx context.Context, params *iam.DeleteRoleInput, optFns ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	DetachRolePolicy(ctx context.Context, params *iam.DetachRolePolicyInput, optFns ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error)
	GetRole(ctx context.Context, params *iam.GetRoleInput, optFns ...func(*iam.Options)) (*iam.GetRoleOutput, error)
}

type Iam struct {
	client IIamSDKClient
}

func NewIam(client IIamSDKClient) *Iam {
	return &Iam{
		client,
	}
}

func (iamClient *Iam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	input := &iam.DeleteRoleInput{
		RoleName: roleName,
	}

	retryCount := 0
	for {
		_, err := iamClient.client.DeleteRole(ctx, input)
		if err != nil && strings.Contains(err.Error(), "api error Throttling: Rate exceeded") {
			retryCount++
			if err := WaitForRetry(retryCount, sleepTimeSec, roleName, err); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (iamClient *Iam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	var marker *string
	policies := []types.AttachedPolicy{}

	for {
		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: roleName,
			Marker:   marker,
		}

		output, err := iamClient.client.ListAttachedRolePolicies(ctx, input)
		if err != nil {
			return nil, err
		}

		policies = append(policies, output.AttachedPolicies...)

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return policies, nil
}

func (iamClient *Iam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	for _, policy := range policies {
		if err := iamClient.DetachRolePolicy(ctx, roleName, policy.PolicyArn, sleepTimeSec); err != nil {
			return err
		}
	}

	return nil
}

func (iamClient *Iam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	input := &iam.DetachRolePolicyInput{
		PolicyArn: PolicyArn,
		RoleName:  roleName,
	}

	retryCount := 0
	for {
		_, err := iamClient.client.DetachRolePolicy(ctx, input)
		if err != nil && strings.Contains(err.Error(), "api error Throttling: Rate exceeded") {
			retryCount++
			if err := WaitForRetry(retryCount, sleepTimeSec, roleName, err); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func (iamClient *Iam) CheckRoleExists(ctx context.Context, roleName *string) (bool, error) {
	input := &iam.GetRoleInput{
		RoleName: roleName,
	}

	_, err := iamClient.client.GetRole(ctx, input)
	if err != nil && strings.Contains(err.Error(), "NoSuchEntity") {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
