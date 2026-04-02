package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*IamUserOperator)(nil)

type IamUserOperator struct {
	client    client.IIamUser
	resources []*types.StackResourceSummary
}

func NewIamUserOperator(client client.IIamUser) *IamUserOperator {
	return &IamUserOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *IamUserOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *IamUserOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *IamUserOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, user := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteIamUser(ctx, user.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *IamUserOperator) DeleteIamUser(ctx context.Context, userName *string) error {
	exists, err := o.client.CheckUserExists(ctx, userName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	if err := o.client.DetachUserPolicies(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteUserInlinePolicies(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeactivateAndDeleteMFADevices(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteAccessKeys(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteLoginProfile(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteSigningCertificates(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteSSHPublicKeys(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteServiceSpecificCredentials(ctx, userName); err != nil {
		return err
	}

	if err := o.client.RemoveUserFromGroups(ctx, userName); err != nil {
		return err
	}

	if err := o.client.DeleteUser(ctx, userName); err != nil {
		return err
	}

	return nil
}
