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

	_, err := i.deleteRoleWithRetry(ctx, input, roleName, sleepTimeSec)
	return err
}

func (i *Iam) deleteRoleWithRetry(
	ctx context.Context,
	input *iam.DeleteRoleInput,
	roleName *string,
	sleepTimeSec int,
) (*iam.DeleteRoleOutput, error) {
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		output, err := i.client.DeleteRole(ctx, input)
		if err != nil && strings.Contains(err.Error(), "api error Throttling: Rate exceeded") {
			retryCount++
			if err := WaitForRetry(retryCount, sleepTimeSec, roleName, err); err != nil {
				return nil, err
			}
			continue
		}
		if err != nil {
			return nil, err
		}

		return output, nil
	}
}

func (i *Iam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
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

		output, err := i.client.ListAttachedRolePolicies(ctx, input)
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

	_, err := i.detachRolePolicyWithRetry(ctx, input, roleName, sleepTimeSec)
	return err
}

func (i *Iam) detachRolePolicyWithRetry(
	ctx context.Context,
	input *iam.DetachRolePolicyInput,
	roleName *string,
	sleepTimeSec int,
) (*iam.DetachRolePolicyOutput, error) {
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		output, err := i.client.DetachRolePolicy(ctx, input)
		if err != nil && strings.Contains(err.Error(), "api error Throttling: Rate exceeded") {
			retryCount++
			if err := WaitForRetry(retryCount, sleepTimeSec, roleName, err); err != nil {
				return nil, err
			}
			continue
		}
		if err != nil {
			return nil, err
		}

		return output, nil
	}
}

func (i *Iam) CheckRoleExists(ctx context.Context, roleName *string) (bool, error) {
	input := &iam.GetRoleInput{
		RoleName: roleName,
	}

	_, err := i.client.GetRole(ctx, input)
	if err != nil && strings.Contains(err.Error(), "NoSuchEntity") {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
