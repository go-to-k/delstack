//go:generate mockgen -source=$GOFILE -destination=iam_mock.go -package=$GOPACKAGE -write_package_comment=false
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
	CheckRoleExists(ctx context.Context, roleName *string) (bool, error)
}

var _ IIam = (*Iam)(nil)

type Iam struct {
	client  *iam.Client
	retryer *Retryer
}

func NewIam(client *iam.Client) *Iam {
	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}
	retryer := NewRetryer(retryable, SleepTimeSecForIam)

	return &Iam{
		client,
		retryer,
	}
}

func (i *Iam) DeleteRole(ctx context.Context, roleName *string) error {
	input := &iam.DeleteRoleInput{
		RoleName: roleName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteRole(ctx, input, optFn)
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

		optFn := func(o *iam.Options) {
			o.Retryer = i.retryer
		}

		output, err := i.client.ListAttachedRolePolicies(ctx, input, optFn)
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
		if err := i.detachRolePolicy(ctx, roleName, policy.PolicyArn); err != nil {
			return &ClientError{
				ResourceName: roleName,
				Err:          err,
			}
		}
	}

	return nil
}

func (i *Iam) detachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string) error {
	input := &iam.DetachRolePolicyInput{
		PolicyArn: PolicyArn,
		RoleName:  roleName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DetachRolePolicy(ctx, input, optFn)
	if err != nil {
		return err
	}
	return nil
}

func (i *Iam) CheckRoleExists(ctx context.Context, roleName *string) (bool, error) {
	input := &iam.GetRoleInput{
		RoleName: roleName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.GetRole(ctx, input, optFn)

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
