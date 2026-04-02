package operation

import (
	"context"
	"runtime"
	"strings"

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

	eg.Go(func() error { return o.detachUserPolicies(egCtx, userName) })
	eg.Go(func() error { return o.deleteUserInlinePolicies(egCtx, userName) })
	eg.Go(func() error { return o.deactivateAndDeleteMFADevices(egCtx, userName) })
	eg.Go(func() error { return o.deleteAccessKeys(egCtx, userName) })
	eg.Go(func() error { return o.client.DeleteLoginProfile(egCtx, userName) })
	eg.Go(func() error { return o.deleteSigningCertificates(egCtx, userName) })
	eg.Go(func() error { return o.deleteSSHPublicKeys(egCtx, userName) })
	eg.Go(func() error { return o.deleteServiceSpecificCredentials(egCtx, userName) })
	eg.Go(func() error { return o.removeUserFromGroups(egCtx, userName) })

	if err := eg.Wait(); err != nil {
		return err
	}

	return o.client.DeleteUser(ctx, userName)
}

func (o *IamUserOperator) detachUserPolicies(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		policies, nextMarker, err := o.client.ListAttachedUserPolicies(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, policy := range policies {
			if err := o.client.DetachUserPolicy(ctx, userName, policy.PolicyArn); err != nil {
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

func (o *IamUserOperator) deleteUserInlinePolicies(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		policyNames, nextMarker, err := o.client.ListUserPolicies(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, policyName := range policyNames {
			if err := o.client.DeleteUserPolicy(ctx, userName, &policyName); err != nil {
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

func (o *IamUserOperator) deactivateAndDeleteMFADevices(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		devices, nextMarker, err := o.client.ListMFADevices(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, device := range devices {
			if err := o.client.DeactivateMFADevice(ctx, userName, device.SerialNumber); err != nil {
				return err
			}

			// Virtual MFA devices have ARN format serial numbers
			if strings.Contains(*device.SerialNumber, ":mfa/") {
				if err := o.client.DeleteVirtualMFADevice(ctx, device.SerialNumber); err != nil {
					return err
				}
			}
		}

		marker = nextMarker
		if marker == nil {
			break
		}
	}

	return nil
}

func (o *IamUserOperator) deleteAccessKeys(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		keys, nextMarker, err := o.client.ListAccessKeys(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if err := o.client.DeleteAccessKey(ctx, userName, key.AccessKeyId); err != nil {
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

func (o *IamUserOperator) deleteSigningCertificates(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		certs, nextMarker, err := o.client.ListSigningCertificates(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, cert := range certs {
			if err := o.client.DeleteSigningCertificate(ctx, userName, cert.CertificateId); err != nil {
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

func (o *IamUserOperator) deleteSSHPublicKeys(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		keys, nextMarker, err := o.client.ListSSHPublicKeys(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, key := range keys {
			if err := o.client.DeleteSSHPublicKey(ctx, userName, key.SSHPublicKeyId); err != nil {
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

func (o *IamUserOperator) deleteServiceSpecificCredentials(ctx context.Context, userName *string) error {
	creds, err := o.client.ListServiceSpecificCredentials(ctx, userName)
	if err != nil {
		return err
	}

	for _, cred := range creds {
		if err := o.client.DeleteServiceSpecificCredential(ctx, userName, cred.ServiceSpecificCredentialId); err != nil {
			return err
		}
	}

	return nil
}

func (o *IamUserOperator) removeUserFromGroups(ctx context.Context, userName *string) error {
	var marker *string

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		groups, nextMarker, err := o.client.ListGroupsForUser(ctx, userName, marker)
		if err != nil {
			return err
		}

		for _, group := range groups {
			if err := o.client.RemoveUserFromGroup(ctx, group.GroupName, userName); err != nil {
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
