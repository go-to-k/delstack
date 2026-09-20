package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*IamRoleOperator)(nil)

type IamRoleOperator struct {
	client    client.IIam
	resources []*types.StackResourceSummary
}

func NewIamRoleOperator(iamClient client.IIam) *IamRoleOperator {
	return &IamRoleOperator{
		client:    iamClient,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *IamRoleOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *IamRoleOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *IamRoleOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, role := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteIamRole(ctx, role.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *IamRoleOperator) DeleteIamRole(ctx context.Context, roleName *string) error {
	exists, err := o.client.CheckRoleExists(ctx, roleName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error { return o.detachRolePolicies(egCtx, roleName) })
	eg.Go(func() error { return o.deleteRoleInlinePolicies(egCtx, roleName) })
	eg.Go(func() error { return o.removeRoleFromInstanceProfiles(egCtx, roleName) })

	if err := eg.Wait(); err != nil {
		return err
	}

	return o.client.DeleteRole(ctx, roleName)
}

func (o *IamRoleOperator) detachRolePolicies(ctx context.Context, roleName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		policies, nextMarker, err := o.client.ListAttachedRolePolicies(ctx, roleName, marker)
		if err != nil {
			return err
		}

		for _, policy := range policies {
			if err := o.client.DetachRolePolicy(ctx, roleName, policy.PolicyArn); err != nil {
				return err
			}
		}

		marker = nextMarker
		if marker == nil {
			break
		}
	}

	return nil
}

func (o *IamRoleOperator) deleteRoleInlinePolicies(ctx context.Context, roleName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		policyNames, nextMarker, err := o.client.ListRolePolicies(ctx, roleName, marker)
		if err != nil {
			return err
		}

		for _, policyName := range policyNames {
			if err := o.client.DeleteRolePolicy(ctx, roleName, &policyName); err != nil {
				return err
			}
		}

		marker = nextMarker
		if marker == nil {
			break
		}
	}

	return nil
}

// removeRoleFromInstanceProfiles removes the role from its instance profiles, which blocks
// DeleteRole. The instance profiles themselves are left as is because they can live outside the stack.
func (o *IamRoleOperator) removeRoleFromInstanceProfiles(ctx context.Context, roleName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		profiles, nextMarker, err := o.client.ListInstanceProfilesForRole(ctx, roleName, marker)
		if err != nil {
			return err
		}

		for _, profile := range profiles {
			if err := o.client.RemoveRoleFromInstanceProfile(ctx, profile.InstanceProfileName, roleName); err != nil {
				return err
			}
		}

		marker = nextMarker
		if marker == nil {
			break
		}
	}

	return nil
}
