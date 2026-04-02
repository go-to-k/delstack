package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

// expectAllDependencyRemovalsSucceed sets up mock expectations for all 9 parallel
// dependency removal methods to succeed. Use AnyTimes() because with parallel execution,
// when one method fails, other goroutines may or may not have been called.
func expectAllDependencyRemovalsSucceed(m *client.MockIIam, userName *string) {
	m.EXPECT().DetachUserPolicies(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeleteAccessKeys(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeleteLoginProfile(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeleteSigningCertificates(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().RemoveUserFromGroups(gomock.Any(), userName).Return(nil).AnyTimes()
}

func TestIamUserOperator_DeleteIamUser(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx      context.Context
		userName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIIam)
		want          error
		wantErr       bool
	}{
		{
			name: "delete user successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				expectAllDependencyRemovalsSucceed(m, aws.String("test"))
				m.EXPECT().DeleteUser(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete user failure for CheckUserExists errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("GetUserError"))
			},
			want:    fmt.Errorf("GetUserError"),
			wantErr: true,
		},
		{
			name: "delete user successfully for CheckUserExists (not exists)",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete user failure for DetachUserPolicies errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DetachUserPoliciesError"))
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DetachUserPoliciesError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteUserInlinePolicies errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteUserInlinePoliciesError"))
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeleteUserInlinePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeactivateAndDeleteMFADevices errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeactivateAndDeleteMFADevicesError"))
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeactivateAndDeleteMFADevicesError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteAccessKeys errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteAccessKeysError"))
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeleteAccessKeysError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteLoginProfile errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteLoginProfileError"))
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeleteLoginProfileError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteSigningCertificates errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteSigningCertificatesError"))
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeleteSigningCertificatesError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteSSHPublicKeys errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteSSHPublicKeysError"))
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeleteSSHPublicKeysError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteServiceSpecificCredentials errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteServiceSpecificCredentialsError"))
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
			},
			want:    fmt.Errorf("DeleteServiceSpecificCredentialsError"),
			wantErr: true,
		},
		{
			name: "delete user failure for RemoveUserFromGroups errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DetachUserPolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteUserInlinePolicies(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeactivateAndDeleteMFADevices(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteAccessKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSigningCertificates(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteSSHPublicKeys(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().DeleteServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().RemoveUserFromGroups(gomock.Any(), aws.String("test")).Return(fmt.Errorf("RemoveUserFromGroupsError"))
			},
			want:    fmt.Errorf("RemoveUserFromGroupsError"),
			wantErr: true,
		},
		{
			name: "delete user failure for DeleteUser errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				expectAllDependencyRemovalsSucceed(m, aws.String("test"))
				m.EXPECT().DeleteUser(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteUserError"))
			},
			want:    fmt.Errorf("DeleteUserError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iamUserMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamUserMock)

			iamUserOperator := NewIamUserOperator(iamUserMock)

			err := iamUserOperator.DeleteIamUser(tt.args.ctx, tt.args.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestIamUserOperator_DeleteResourcesForIamUser(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIIam)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				expectAllDependencyRemovalsSucceed(m, aws.String("PhysicalResourceId1"))
				m.EXPECT().DeleteUser(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("GetUserError"))
			},
			want:    fmt.Errorf("GetUserError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iamUserMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamUserMock)

			iamUserOperator := NewIamUserOperator(iamUserMock)
			iamUserOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::IAM::User"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := iamUserOperator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
