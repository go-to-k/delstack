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
	client    client.IIam
	resources []*types.StackResourceSummary
}

func NewIamUserOperator(iamClient client.IIam) *IamUserOperator {
	return &IamUserOperator{
		client:    iamClient,
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

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error { return o.client.DetachUserPolicies(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteUserInlinePolicies(egCtx, userName) })
	eg.Go(func() error { return o.client.DeactivateAndDeleteMFADevices(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteAccessKeys(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteLoginProfile(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteSigningCertificates(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteSSHPublicKeys(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteServiceSpecificCredentials(egCtx, userName) })
	eg.Go(func() error { return o.client.RemoveUserFromGroups(egCtx, userName) })

	if err := eg.Wait(); err != nil {
		return err
	}

	return o.client.DeleteUser(ctx, userName)
}
