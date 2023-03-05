//go:generate mockgen -source=./iam.go -destination=./iam_mock.go -package=client
package client

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

var SleepTimeSecForIam = 5

type IIam interface {
	DeleteRole(ctx context.Context, roleName *string) error
	ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error)
	DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy) error
	DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string) error
	CheckRoleExists(ctx context.Context, roleName *string) (bool, error)
}

var _ IIam = (*Iam)(nil)

type Iam struct {
	client *iam.Client
}

func NewIam(client *iam.Client) *Iam {
	return &Iam{
		client,
	}
}

func (i *Iam) DeleteRole(ctx context.Context, roleName *string) error {
	input := &iam.DeleteRoleInput{
		RoleName: roleName,
	}

	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}

	_, err := Retry(
		&RetryInput[iam.DeleteRoleInput, iam.DeleteRoleOutput, iam.Options]{
			Ctx:              ctx,
			SleepTimeSec:     SleepTimeSecForIam,
			TargetResource:   roleName,
			Input:            input,
			ApiCaller:        i.client.DeleteRole,
			RetryableChecker: retryable,
		},
	)
	if err != nil {
		return &ClientError{
			ResourceName: roleName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	var marker *string
	policies := []types.AttachedPolicy{}

	for {
		select {
		case <-ctx.Done():
			return policies, &ClientError{
				ResourceName: roleName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListAttachedRolePoliciesInput{
			RoleName: roleName,
			Marker:   marker,
		}

		retryable := func(err error) bool {
			return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
		}

		output, err := Retry(
			&RetryInput[iam.ListAttachedRolePoliciesInput, iam.ListAttachedRolePoliciesOutput, iam.Options]{
				Ctx:              ctx,
				SleepTimeSec:     SleepTimeSecForIam,
				TargetResource:   roleName,
				Input:            input,
				ApiCaller:        i.client.ListAttachedRolePolicies,
				RetryableChecker: retryable,
			},
		)
		if err != nil {
			return nil, &ClientError{
				ResourceName: roleName,
				Err:          err,
			}
		}

		policies = append(policies, output.AttachedPolicies...)

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return policies, nil
}

func (i *Iam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy) error {
	for _, policy := range policies {
		if err := i.DetachRolePolicy(ctx, roleName, policy.PolicyArn); err != nil {
			return err // return non wrapping error because already wrapped error in DetachRolePolicy
		}
	}

	return nil
}

func (i *Iam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string) error {
	input := &iam.DetachRolePolicyInput{
		PolicyArn: PolicyArn,
		RoleName:  roleName,
	}

	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}

	_, err := Retry(
		&RetryInput[iam.DetachRolePolicyInput, iam.DetachRolePolicyOutput, iam.Options]{
			Ctx:              ctx,
			SleepTimeSec:     SleepTimeSecForIam,
			TargetResource:   roleName,
			Input:            input,
			ApiCaller:        i.client.DetachRolePolicy,
			RetryableChecker: retryable,
		},
	)
	if err != nil {
		return &ClientError{
			ResourceName: roleName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) CheckRoleExists(ctx context.Context, roleName *string) (bool, error) {
	input := &iam.GetRoleInput{
		RoleName: roleName,
	}

	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}

	_, err := Retry(
		&RetryInput[iam.GetRoleInput, iam.GetRoleOutput, iam.Options]{
			Ctx:              ctx,
			SleepTimeSec:     SleepTimeSecForIam,
			TargetResource:   roleName,
			Input:            input,
			ApiCaller:        i.client.GetRole,
			RetryableChecker: retryable,
		},
	)

	if err != nil && strings.Contains(err.Error(), "NoSuchEntity") {
		return false, nil
	}
	if err != nil {
		return false, &ClientError{
			ResourceName: roleName,
			Err:          err,
		}
	}

	return true, nil
}
