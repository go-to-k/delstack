//go:generate mockgen -source=$GOFILE -destination=iam_user_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type IIamUser interface {
	CheckUserExists(ctx context.Context, userName *string) (bool, error)
	DetachUserPolicies(ctx context.Context, userName *string) error
	DeleteUserInlinePolicies(ctx context.Context, userName *string) error
	DeactivateAndDeleteMFADevices(ctx context.Context, userName *string) error
	DeleteAccessKeys(ctx context.Context, userName *string) error
	DeleteLoginProfile(ctx context.Context, userName *string) error
	DeleteSigningCertificates(ctx context.Context, userName *string) error
	DeleteSSHPublicKeys(ctx context.Context, userName *string) error
	DeleteServiceSpecificCredentials(ctx context.Context, userName *string) error
	RemoveUserFromGroups(ctx context.Context, userName *string) error
	DeleteUser(ctx context.Context, userName *string) error
}

var _ IIamUser = (*IamUser)(nil)

type IamUser struct {
	client  *iam.Client
	retryer *Retryer
}

func NewIamUser(client *iam.Client) *IamUser {
	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}
	retryer := NewRetryer(retryable, SleepTimeSecForIam)

	return &IamUser{
		client:  client,
		retryer: retryer,
	}
}

func (u *IamUser) CheckUserExists(ctx context.Context, userName *string) (bool, error) {
	input := &iam.GetUserInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.GetUser(ctx, input, optFn)
	if err != nil && strings.Contains(err.Error(), "NoSuchEntity") {
		return false, nil
	}
	if err != nil {
		return false, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return true, nil
}

func (u *IamUser) DetachUserPolicies(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListAttachedUserPoliciesInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListAttachedUserPolicies(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, policy := range output.AttachedPolicies {
			if err := u.detachUserPolicy(ctx, userName, policy.PolicyArn); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) detachUserPolicy(ctx context.Context, userName *string, policyArn *string) error {
	input := &iam.DetachUserPolicyInput{
		UserName:  userName,
		PolicyArn: policyArn,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DetachUserPolicy(ctx, input, optFn)
	return err
}

func (u *IamUser) DeleteUserInlinePolicies(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListUserPoliciesInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListUserPolicies(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, policyName := range output.PolicyNames {
			if err := u.deleteUserPolicy(ctx, userName, &policyName); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) deleteUserPolicy(ctx context.Context, userName *string, policyName *string) error {
	input := &iam.DeleteUserPolicyInput{
		UserName:   userName,
		PolicyName: policyName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteUserPolicy(ctx, input, optFn)
	return err
}

func (u *IamUser) DeactivateAndDeleteMFADevices(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListMFADevicesInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListMFADevices(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, device := range output.MFADevices {
			if err := u.deactivateMFADevice(ctx, userName, device.SerialNumber); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}

			// Virtual MFA devices have ARN format serial numbers
			if strings.Contains(*device.SerialNumber, ":mfa/") {
				if err := u.deleteVirtualMFADevice(ctx, device.SerialNumber); err != nil {
					return &ClientError{
						ResourceName: userName,
						Err:          err,
					}
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) deactivateMFADevice(ctx context.Context, userName *string, serialNumber *string) error {
	input := &iam.DeactivateMFADeviceInput{
		UserName:     userName,
		SerialNumber: serialNumber,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeactivateMFADevice(ctx, input, optFn)
	return err
}

func (u *IamUser) deleteVirtualMFADevice(ctx context.Context, serialNumber *string) error {
	input := &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: serialNumber,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteVirtualMFADevice(ctx, input, optFn)
	return err
}

func (u *IamUser) DeleteAccessKeys(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListAccessKeysInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListAccessKeys(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, key := range output.AccessKeyMetadata {
			if err := u.deleteAccessKey(ctx, userName, key.AccessKeyId); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) deleteAccessKey(ctx context.Context, userName *string, accessKeyId *string) error {
	input := &iam.DeleteAccessKeyInput{
		UserName:    userName,
		AccessKeyId: accessKeyId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteAccessKey(ctx, input, optFn)
	return err
}

func (u *IamUser) DeleteLoginProfile(ctx context.Context, userName *string) error {
	input := &iam.DeleteLoginProfileInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteLoginProfile(ctx, input, optFn)
	if err != nil && strings.Contains(err.Error(), "NoSuchEntity") {
		return nil
	}
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return nil
}

func (u *IamUser) DeleteSigningCertificates(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListSigningCertificatesInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListSigningCertificates(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, cert := range output.Certificates {
			if err := u.deleteSigningCertificate(ctx, userName, cert.CertificateId); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) deleteSigningCertificate(ctx context.Context, userName *string, certificateId *string) error {
	input := &iam.DeleteSigningCertificateInput{
		UserName:      userName,
		CertificateId: certificateId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteSigningCertificate(ctx, input, optFn)
	return err
}

func (u *IamUser) DeleteSSHPublicKeys(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListSSHPublicKeysInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListSSHPublicKeys(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, key := range output.SSHPublicKeys {
			if err := u.deleteSSHPublicKey(ctx, userName, key.SSHPublicKeyId); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) deleteSSHPublicKey(ctx context.Context, userName *string, sshPublicKeyId *string) error {
	input := &iam.DeleteSSHPublicKeyInput{
		UserName:       userName,
		SSHPublicKeyId: sshPublicKeyId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteSSHPublicKey(ctx, input, optFn)
	return err
}

func (u *IamUser) DeleteServiceSpecificCredentials(ctx context.Context, userName *string) error {
	input := &iam.ListServiceSpecificCredentialsInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	output, err := u.client.ListServiceSpecificCredentials(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	for _, cred := range output.ServiceSpecificCredentials {
		if err := u.deleteServiceSpecificCredential(ctx, userName, cred.ServiceSpecificCredentialId); err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}
	}

	return nil
}

func (u *IamUser) deleteServiceSpecificCredential(ctx context.Context, userName *string, credentialId *string) error {
	input := &iam.DeleteServiceSpecificCredentialInput{
		UserName:                    userName,
		ServiceSpecificCredentialId: credentialId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteServiceSpecificCredential(ctx, input, optFn)
	return err
}

func (u *IamUser) RemoveUserFromGroups(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	for {
		select {
		case <-ctx.Done():
			return &ClientError{
				ResourceName: userName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &iam.ListGroupsForUserInput{
			UserName: userName,
			Marker:   marker,
		}

		output, err := u.client.ListGroupsForUser(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, group := range output.Groups {
			if err := u.removeUserFromGroup(ctx, userName, group.GroupName); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}
		}

		marker = output.Marker
		if marker == nil {
			break
		}
	}

	return nil
}

func (u *IamUser) removeUserFromGroup(ctx context.Context, userName *string, groupName *string) error {
	input := &iam.RemoveUserFromGroupInput{
		UserName:  userName,
		GroupName: groupName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.RemoveUserFromGroup(ctx, input, optFn)
	return err
}

func (u *IamUser) DeleteUser(ctx context.Context, userName *string) error {
	input := &iam.DeleteUserInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = u.retryer
	}

	_, err := u.client.DeleteUser(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}
