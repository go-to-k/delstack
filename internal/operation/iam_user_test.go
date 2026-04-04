package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

// expectAllDependencyRemovalsSucceed sets up mock expectations for all 9 parallel
// dependency removal operations to succeed. Use AnyTimes() because with parallel execution,
// when one method fails, other goroutines may or may not have been called.
func expectAllDependencyRemovalsSucceed(m *client.MockIIam, userName *string) {
	m.EXPECT().ListAttachedUserPolicies(gomock.Any(), userName, gomock.Nil()).Return([]types.AttachedPolicy{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListUserPolicies(gomock.Any(), userName, gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListMFADevices(gomock.Any(), userName, gomock.Nil()).Return([]types.MFADevice{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListAccessKeys(gomock.Any(), userName, gomock.Nil()).Return([]types.AccessKeyMetadata{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().DeleteLoginProfile(gomock.Any(), userName).Return(nil).AnyTimes()
	m.EXPECT().ListSigningCertificates(gomock.Any(), userName, gomock.Nil()).Return([]types.SigningCertificate{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListSSHPublicKeys(gomock.Any(), userName, gomock.Nil()).Return([]types.SSHPublicKeyMetadata{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListServiceSpecificCredentials(gomock.Any(), userName).Return([]types.ServiceSpecificCredentialMetadata{}, nil).AnyTimes()
	m.EXPECT().ListGroupsForUser(gomock.Any(), userName, gomock.Nil()).Return([]types.Group{}, (*string)(nil), nil).AnyTimes()
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
			name: "delete user successfully with dependencies",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)

				// Policies
				m.EXPECT().ListAttachedUserPolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.AttachedPolicy{{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DetachUserPolicy(gomock.Any(), aws.String("test"), aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")).Return(nil).AnyTimes()

				// Inline policies
				m.EXPECT().ListUserPolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]string{"InlinePolicy1"}, (*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DeleteUserPolicy(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil).AnyTimes()

				// MFA
				m.EXPECT().ListMFADevices(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.MFADevice{{SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"), UserName: aws.String("test")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DeactivateMFADevice(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil).AnyTimes()
				m.EXPECT().DeleteVirtualMFADevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				// Access keys
				m.EXPECT().ListAccessKeys(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.AccessKeyMetadata{{AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DeleteAccessKey(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil).AnyTimes()

				// Login profile
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()

				// Signing certs
				m.EXPECT().ListSigningCertificates(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.SigningCertificate{}, (*string)(nil), nil).AnyTimes()

				// SSH keys
				m.EXPECT().ListSSHPublicKeys(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.SSHPublicKeyMetadata{}, (*string)(nil), nil).AnyTimes()

				// Service specific credentials
				m.EXPECT().ListServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return([]types.ServiceSpecificCredentialMetadata{}, nil).AnyTimes()

				// Groups
				m.EXPECT().ListGroupsForUser(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.Group{{GroupName: aws.String("Group1")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().RemoveUserFromGroup(gomock.Any(), aws.String("Group1"), aws.String("test")).Return(nil).AnyTimes()

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
			name: "delete user failure for ListAttachedUserPolicies errors",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckUserExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedUserPolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(nil, nil, fmt.Errorf("ListAttachedUserPoliciesError"))
				m.EXPECT().ListUserPolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListMFADevices(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.MFADevice{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListAccessKeys(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.AccessKeyMetadata{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().DeleteLoginProfile(gomock.Any(), aws.String("test")).Return(nil).AnyTimes()
				m.EXPECT().ListSigningCertificates(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.SigningCertificate{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListSSHPublicKeys(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.SSHPublicKeyMetadata{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListServiceSpecificCredentials(gomock.Any(), aws.String("test")).Return([]types.ServiceSpecificCredentialMetadata{}, nil).AnyTimes()
				m.EXPECT().ListGroupsForUser(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.Group{}, (*string)(nil), nil).AnyTimes()
			},
			want:    fmt.Errorf("ListAttachedUserPoliciesError"),
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
			iamMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamMock)

			iamUserOperator := NewIamUserOperator(iamMock)

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
			iamMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamMock)

			iamUserOperator := NewIamUserOperator(iamMock)
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
