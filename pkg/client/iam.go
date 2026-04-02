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
	// Group operations
	DeleteGroup(ctx context.Context, groupName *string) error
	CheckGroupExists(ctx context.Context, groupName *string) (bool, error)
	GetGroupUsers(ctx context.Context, groupName *string) ([]types.User, error)
	RemoveUsersFromGroup(ctx context.Context, groupName *string, users []types.User) error

	// User operations
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

var _ IIam = (*Iam)(nil)

type Iam struct {
	client            *iam.Client
	retryer           *Retryer
	outputForGetGroup *iam.GetGroupOutput // for caching
}

func NewIam(client *iam.Client) *Iam {
	retryable := func(err error) bool {
		return strings.Contains(err.Error(), "api error Throttling: Rate exceeded")
	}
	retryer := NewRetryer(retryable, SleepTimeSecForIam)

	return &Iam{
		client,
		retryer,
		nil,
	}
}

// Group operations

func (i *Iam) DeleteGroup(ctx context.Context, groupName *string) error {
	input := &iam.DeleteGroupInput{
		GroupName: groupName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteGroup(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: groupName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) CheckGroupExists(ctx context.Context, groupName *string) (bool, error) {
	_, err := i.getGroup(ctx, groupName)
	if err != nil && strings.Contains(err.Error(), "NoSuchEntity") {
		return false, nil
	}
	if err != nil {
		return false, &ClientError{
			ResourceName: groupName,
			Err:          err,
		}
	}

	return true, nil
}

func (i *Iam) GetGroupUsers(ctx context.Context, groupName *string) ([]types.User, error) {
	output, err := i.getGroup(ctx, groupName)
	if err != nil {
		return nil, &ClientError{
			ResourceName: groupName,
			Err:          err,
		}
	}

	return output.Users, nil
}

func (i *Iam) getGroup(ctx context.Context, groupName *string) (*iam.GetGroupOutput, error) {
	if i.outputForGetGroup != nil {
		return i.outputForGetGroup, nil
	}

	input := &iam.GetGroupInput{
		GroupName: groupName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	// GetGroup returns a Marker, but we don't need to paginate because we get only one group by a full name
	output, err := i.client.GetGroup(ctx, input, optFn)
	if err != nil {
		return nil, err
	}

	i.outputForGetGroup = output
	return output, nil
}

func (i *Iam) RemoveUsersFromGroup(ctx context.Context, groupName *string, users []types.User) error {
	for _, user := range users {
		if err := i.removeUserFromGroup(ctx, groupName, user.UserName); err != nil {
			return &ClientError{
				ResourceName: groupName,
				Err:          err,
			}
		}
	}

	return nil
}

func (i *Iam) removeUserFromGroup(ctx context.Context, groupName *string, userName *string) error {
	input := &iam.RemoveUserFromGroupInput{
		UserName:  userName,
		GroupName: groupName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.RemoveUserFromGroup(ctx, input, optFn)
	if err != nil {
		return err
	}
	return nil
}

// User operations

func (i *Iam) CheckUserExists(ctx context.Context, userName *string) (bool, error) {
	input := &iam.GetUserInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.GetUser(ctx, input, optFn)
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

func (i *Iam) DetachUserPolicies(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListAttachedUserPolicies(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, policy := range output.AttachedPolicies {
			if err := i.detachUserPolicy(ctx, userName, policy.PolicyArn); err != nil {
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

func (i *Iam) detachUserPolicy(ctx context.Context, userName *string, policyArn *string) error {
	input := &iam.DetachUserPolicyInput{
		UserName:  userName,
		PolicyArn: policyArn,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DetachUserPolicy(ctx, input, optFn)
	return err
}

func (i *Iam) DeleteUserInlinePolicies(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListUserPolicies(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, policyName := range output.PolicyNames {
			if err := i.deleteUserPolicy(ctx, userName, &policyName); err != nil {
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

func (i *Iam) deleteUserPolicy(ctx context.Context, userName *string, policyName *string) error {
	input := &iam.DeleteUserPolicyInput{
		UserName:   userName,
		PolicyName: policyName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteUserPolicy(ctx, input, optFn)
	return err
}

func (i *Iam) DeactivateAndDeleteMFADevices(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListMFADevices(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, device := range output.MFADevices {
			if err := i.deactivateMFADevice(ctx, userName, device.SerialNumber); err != nil {
				return &ClientError{
					ResourceName: userName,
					Err:          err,
				}
			}

			// Virtual MFA devices have ARN format serial numbers
			if strings.Contains(*device.SerialNumber, ":mfa/") {
				if err := i.deleteVirtualMFADevice(ctx, device.SerialNumber); err != nil {
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

func (i *Iam) deactivateMFADevice(ctx context.Context, userName *string, serialNumber *string) error {
	input := &iam.DeactivateMFADeviceInput{
		UserName:     userName,
		SerialNumber: serialNumber,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeactivateMFADevice(ctx, input, optFn)
	return err
}

func (i *Iam) deleteVirtualMFADevice(ctx context.Context, serialNumber *string) error {
	input := &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: serialNumber,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteVirtualMFADevice(ctx, input, optFn)
	return err
}

func (i *Iam) DeleteAccessKeys(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListAccessKeys(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, key := range output.AccessKeyMetadata {
			if err := i.deleteAccessKey(ctx, userName, key.AccessKeyId); err != nil {
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

func (i *Iam) deleteAccessKey(ctx context.Context, userName *string, accessKeyId *string) error {
	input := &iam.DeleteAccessKeyInput{
		UserName:    userName,
		AccessKeyId: accessKeyId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteAccessKey(ctx, input, optFn)
	return err
}

func (i *Iam) DeleteLoginProfile(ctx context.Context, userName *string) error {
	input := &iam.DeleteLoginProfileInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteLoginProfile(ctx, input, optFn)
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

func (i *Iam) DeleteSigningCertificates(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListSigningCertificates(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, cert := range output.Certificates {
			if err := i.deleteSigningCertificate(ctx, userName, cert.CertificateId); err != nil {
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

func (i *Iam) deleteSigningCertificate(ctx context.Context, userName *string, certificateId *string) error {
	input := &iam.DeleteSigningCertificateInput{
		UserName:      userName,
		CertificateId: certificateId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteSigningCertificate(ctx, input, optFn)
	return err
}

func (i *Iam) DeleteSSHPublicKeys(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListSSHPublicKeys(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, key := range output.SSHPublicKeys {
			if err := i.deleteSSHPublicKey(ctx, userName, key.SSHPublicKeyId); err != nil {
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

func (i *Iam) deleteSSHPublicKey(ctx context.Context, userName *string, sshPublicKeyId *string) error {
	input := &iam.DeleteSSHPublicKeyInput{
		UserName:       userName,
		SSHPublicKeyId: sshPublicKeyId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteSSHPublicKey(ctx, input, optFn)
	return err
}

func (i *Iam) DeleteServiceSpecificCredentials(ctx context.Context, userName *string) error {
	input := &iam.ListServiceSpecificCredentialsInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListServiceSpecificCredentials(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	for _, cred := range output.ServiceSpecificCredentials {
		if err := i.deleteServiceSpecificCredential(ctx, userName, cred.ServiceSpecificCredentialId); err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}
	}

	return nil
}

func (i *Iam) deleteServiceSpecificCredential(ctx context.Context, userName *string, credentialId *string) error {
	input := &iam.DeleteServiceSpecificCredentialInput{
		UserName:                    userName,
		ServiceSpecificCredentialId: credentialId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteServiceSpecificCredential(ctx, input, optFn)
	return err
}

func (i *Iam) RemoveUserFromGroups(ctx context.Context, userName *string) error {
	var marker *string

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
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

		output, err := i.client.ListGroupsForUser(ctx, input, optFn)
		if err != nil {
			return &ClientError{
				ResourceName: userName,
				Err:          err,
			}
		}

		for _, group := range output.Groups {
			if err := i.removeUserFromGroupByUserName(ctx, userName, group.GroupName); err != nil {
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

func (i *Iam) removeUserFromGroupByUserName(ctx context.Context, userName *string, groupName *string) error {
	input := &iam.RemoveUserFromGroupInput{
		UserName:  userName,
		GroupName: groupName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.RemoveUserFromGroup(ctx, input, optFn)
	return err
}

func (i *Iam) DeleteUser(ctx context.Context, userName *string) error {
	input := &iam.DeleteUserInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteUser(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}
