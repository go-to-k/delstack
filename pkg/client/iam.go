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
	DeleteUser(ctx context.Context, userName *string) error
	DeleteLoginProfile(ctx context.Context, userName *string) error
	ListAttachedUserPolicies(ctx context.Context, userName *string, marker *string) ([]types.AttachedPolicy, *string, error)
	DetachUserPolicy(ctx context.Context, userName *string, policyArn *string) error
	ListUserPolicies(ctx context.Context, userName *string, marker *string) ([]string, *string, error)
	DeleteUserPolicy(ctx context.Context, userName *string, policyName *string) error
	ListMFADevices(ctx context.Context, userName *string, marker *string) ([]types.MFADevice, *string, error)
	DeactivateMFADevice(ctx context.Context, userName *string, serialNumber *string) error
	DeleteVirtualMFADevice(ctx context.Context, serialNumber *string) error
	ListAccessKeys(ctx context.Context, userName *string, marker *string) ([]types.AccessKeyMetadata, *string, error)
	DeleteAccessKey(ctx context.Context, userName *string, accessKeyId *string) error
	ListSigningCertificates(ctx context.Context, userName *string, marker *string) ([]types.SigningCertificate, *string, error)
	DeleteSigningCertificate(ctx context.Context, userName *string, certificateId *string) error
	ListSSHPublicKeys(ctx context.Context, userName *string, marker *string) ([]types.SSHPublicKeyMetadata, *string, error)
	DeleteSSHPublicKey(ctx context.Context, userName *string, sshPublicKeyId *string) error
	ListServiceSpecificCredentials(ctx context.Context, userName *string) ([]types.ServiceSpecificCredentialMetadata, error)
	DeleteServiceSpecificCredential(ctx context.Context, userName *string, credentialId *string) error
	ListGroupsForUser(ctx context.Context, userName *string, marker *string) ([]types.Group, *string, error)
	RemoveUserFromGroup(ctx context.Context, groupName *string, userName *string) error
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
		if err := i.RemoveUserFromGroup(ctx, groupName, user.UserName); err != nil {
			return &ClientError{
				ResourceName: groupName,
				Err:          err,
			}
		}
	}

	return nil
}

func (i *Iam) RemoveUserFromGroup(ctx context.Context, groupName *string, userName *string) error {
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

func (i *Iam) ListAttachedUserPolicies(ctx context.Context, userName *string, marker *string) ([]types.AttachedPolicy, *string, error) {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListAttachedUserPolicies(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.AttachedPolicies, output.Marker, nil
}

func (i *Iam) DetachUserPolicy(ctx context.Context, userName *string, policyArn *string) error {
	input := &iam.DetachUserPolicyInput{
		UserName:  userName,
		PolicyArn: policyArn,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DetachUserPolicy(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListUserPolicies(ctx context.Context, userName *string, marker *string) ([]string, *string, error) {
	input := &iam.ListUserPoliciesInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListUserPolicies(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.PolicyNames, output.Marker, nil
}

func (i *Iam) DeleteUserPolicy(ctx context.Context, userName *string, policyName *string) error {
	input := &iam.DeleteUserPolicyInput{
		UserName:   userName,
		PolicyName: policyName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteUserPolicy(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListMFADevices(ctx context.Context, userName *string, marker *string) ([]types.MFADevice, *string, error) {
	input := &iam.ListMFADevicesInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListMFADevices(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.MFADevices, output.Marker, nil
}

func (i *Iam) DeactivateMFADevice(ctx context.Context, userName *string, serialNumber *string) error {
	input := &iam.DeactivateMFADeviceInput{
		UserName:     userName,
		SerialNumber: serialNumber,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeactivateMFADevice(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) DeleteVirtualMFADevice(ctx context.Context, serialNumber *string) error {
	input := &iam.DeleteVirtualMFADeviceInput{
		SerialNumber: serialNumber,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteVirtualMFADevice(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: serialNumber,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListAccessKeys(ctx context.Context, userName *string, marker *string) ([]types.AccessKeyMetadata, *string, error) {
	input := &iam.ListAccessKeysInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListAccessKeys(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.AccessKeyMetadata, output.Marker, nil
}

func (i *Iam) DeleteAccessKey(ctx context.Context, userName *string, accessKeyId *string) error {
	input := &iam.DeleteAccessKeyInput{
		UserName:    userName,
		AccessKeyId: accessKeyId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteAccessKey(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListSigningCertificates(ctx context.Context, userName *string, marker *string) ([]types.SigningCertificate, *string, error) {
	input := &iam.ListSigningCertificatesInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListSigningCertificates(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.Certificates, output.Marker, nil
}

func (i *Iam) DeleteSigningCertificate(ctx context.Context, userName *string, certificateId *string) error {
	input := &iam.DeleteSigningCertificateInput{
		UserName:      userName,
		CertificateId: certificateId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteSigningCertificate(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListSSHPublicKeys(ctx context.Context, userName *string, marker *string) ([]types.SSHPublicKeyMetadata, *string, error) {
	input := &iam.ListSSHPublicKeysInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListSSHPublicKeys(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.SSHPublicKeys, output.Marker, nil
}

func (i *Iam) DeleteSSHPublicKey(ctx context.Context, userName *string, sshPublicKeyId *string) error {
	input := &iam.DeleteSSHPublicKeyInput{
		UserName:       userName,
		SSHPublicKeyId: sshPublicKeyId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteSSHPublicKey(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListServiceSpecificCredentials(ctx context.Context, userName *string) ([]types.ServiceSpecificCredentialMetadata, error) {
	input := &iam.ListServiceSpecificCredentialsInput{
		UserName: userName,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListServiceSpecificCredentials(ctx, input, optFn)
	if err != nil {
		return nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.ServiceSpecificCredentials, nil
}

func (i *Iam) DeleteServiceSpecificCredential(ctx context.Context, userName *string, credentialId *string) error {
	input := &iam.DeleteServiceSpecificCredentialInput{
		UserName:                    userName,
		ServiceSpecificCredentialId: credentialId,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	_, err := i.client.DeleteServiceSpecificCredential(ctx, input, optFn)
	if err != nil {
		return &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}
	return nil
}

func (i *Iam) ListGroupsForUser(ctx context.Context, userName *string, marker *string) ([]types.Group, *string, error) {
	input := &iam.ListGroupsForUserInput{
		UserName: userName,
		Marker:   marker,
	}

	optFn := func(o *iam.Options) {
		o.Retryer = i.retryer
	}

	output, err := i.client.ListGroupsForUser(ctx, input, optFn)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: userName,
			Err:          err,
		}
	}

	return output.Groups, output.Marker, nil
}
