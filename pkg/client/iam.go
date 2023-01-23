package client

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type IIam interface {
	DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error
	ListAttachedRolePolicies(ctx context.Context, roleName *string, sleepTimeSec int) ([]types.AttachedPolicy, error)
	DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error
	DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error
	CheckRoleExists(ctx context.Context, roleName *string, sleepTimeSec int) (bool, error)
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

func (i *Iam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	input := &iam.DeleteRoleInput{
		RoleName: roleName,
	}

	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}

	_, err := Retry(
		&RetryInput[iam.DeleteRoleInput, iam.DeleteRoleOutput, iam.Options]{
			Ctx:              ctx,
			SleepTimeSec:     sleepTimeSec,
			TargetResource:   roleName,
			Input:            input,
			ApiCaller:        i.client.DeleteRole,
			RetryableChecker: retryable,
		},
	)
	return err
}

func (i *Iam) ListAttachedRolePolicies(ctx context.Context, roleName *string, sleepTimeSec int) ([]types.AttachedPolicy, error) {
	var marker *string
	policies := []types.AttachedPolicy{}

	for {
		select {
		case <-ctx.Done():
			return policies, ctx.Err()
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
				SleepTimeSec:     sleepTimeSec,
				TargetResource:   roleName,
				Input:            input,
				ApiCaller:        i.client.ListAttachedRolePolicies,
				RetryableChecker: retryable,
			},
		)

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

func (i *Iam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	for _, policy := range policies {
		if err := i.DetachRolePolicy(ctx, roleName, policy.PolicyArn, sleepTimeSec); err != nil {
			return err
		}
	}

	return nil
}

func (i *Iam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
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
			SleepTimeSec:     sleepTimeSec,
			TargetResource:   roleName,
			Input:            input,
			ApiCaller:        i.client.DetachRolePolicy,
			RetryableChecker: retryable,
		},
	)
	return err
}

func (i *Iam) CheckRoleExists(ctx context.Context, roleName *string, sleepTimeSec int) (bool, error) {
	input := &iam.GetRoleInput{
		RoleName: roleName,
	}

	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}

	_, err := Retry(
		&RetryInput[iam.GetRoleInput, iam.GetRoleOutput, iam.Options]{
			Ctx:              ctx,
			SleepTimeSec:     sleepTimeSec,
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
		return false, err
	}

	return true, nil
}
