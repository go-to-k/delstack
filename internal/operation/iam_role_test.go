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

// expectAllRoleDependencyRemovalsSucceed sets up mock expectations for all 3 parallel
// dependency removal operations to succeed. Use AnyTimes() because with parallel execution,
// when one method fails, other goroutines may or may not have been called.
func expectAllRoleDependencyRemovalsSucceed(m *client.MockIIam, roleName *string) {
	m.EXPECT().ListAttachedRolePolicies(gomock.Any(), roleName, gomock.Nil()).Return([]types.AttachedPolicy{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListRolePolicies(gomock.Any(), roleName, gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()
	m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), roleName, gomock.Nil()).Return([]types.InstanceProfile{}, (*string)(nil), nil).AnyTimes()
}

func TestIamRoleOperator_DeleteIamRole(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx      context.Context
		roleName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIIam)
		want          error
		wantErr       bool
	}{
		{
			name: "delete role successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				expectAllRoleDependencyRemovalsSucceed(m, aws.String("test"))
				m.EXPECT().DeleteRole(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role successfully with dependencies",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)

				// Attached policies
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.AttachedPolicy{{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DetachRolePolicy(gomock.Any(), aws.String("test"), aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")).Return(nil).AnyTimes()

				// Inline policies
				m.EXPECT().ListRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]string{"InlinePolicy1"}, (*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DeleteRolePolicy(gomock.Any(), aws.String("test"), aws.String("InlinePolicy1")).Return(nil).AnyTimes()

				// Instance profiles
				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.InstanceProfile{{InstanceProfileName: aws.String("testProfile")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().RemoveRoleFromInstanceProfile(gomock.Any(), aws.String("testProfile"), aws.String("test")).Return(nil).AnyTimes()

				m.EXPECT().DeleteRole(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role successfully with pagination",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)

				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.AttachedPolicy{{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")}},
					aws.String("next"), nil,
				).AnyTimes()
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), aws.String("next")).Return(
					[]types.AttachedPolicy{{PolicyArn: aws.String("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().DetachRolePolicy(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil).AnyTimes()

				m.EXPECT().ListRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()

				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.InstanceProfile{{InstanceProfileName: aws.String("testProfile1")}},
					aws.String("next"), nil,
				).AnyTimes()
				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), aws.String("next")).Return(
					[]types.InstanceProfile{{InstanceProfileName: aws.String("testProfile2")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().RemoveRoleFromInstanceProfile(gomock.Any(), gomock.Any(), aws.String("test")).Return(nil).AnyTimes()

				m.EXPECT().DeleteRole(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure for CheckRoleExists errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("GetRoleError"))
			},
			want:    fmt.Errorf("GetRoleError"),
			wantErr: true,
		},
		{
			name: "delete role successfully for CheckRoleExists (not exists)",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure for ListAttachedRolePolicies errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(nil, nil, fmt.Errorf("ListAttachedRolePoliciesError"))
				m.EXPECT().ListRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.InstanceProfile{}, (*string)(nil), nil).AnyTimes()
			},
			want:    fmt.Errorf("ListAttachedRolePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete role failure for ListRolePolicies errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.AttachedPolicy{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return(nil, nil, fmt.Errorf("ListRolePoliciesError"))
				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.InstanceProfile{}, (*string)(nil), nil).AnyTimes()
			},
			want:    fmt.Errorf("ListRolePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete role failure for ListInstanceProfilesForRole errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.AttachedPolicy{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), gomock.Nil()).Return(nil, nil, fmt.Errorf("ListInstanceProfilesForRoleError"))
			},
			want:    fmt.Errorf("ListInstanceProfilesForRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for RemoveRoleFromInstanceProfile errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]types.AttachedPolicy{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListRolePolicies(gomock.Any(), aws.String("test"), gomock.Nil()).Return([]string{}, (*string)(nil), nil).AnyTimes()
				m.EXPECT().ListInstanceProfilesForRole(gomock.Any(), aws.String("test"), gomock.Nil()).Return(
					[]types.InstanceProfile{{InstanceProfileName: aws.String("testProfile")}},
					(*string)(nil), nil,
				).AnyTimes()
				m.EXPECT().RemoveRoleFromInstanceProfile(gomock.Any(), aws.String("testProfile"), aws.String("test")).Return(fmt.Errorf("RemoveRoleFromInstanceProfileError"))
			},
			want:    fmt.Errorf("RemoveRoleFromInstanceProfileError"),
			wantErr: true,
		},
		{
			name: "delete role failure for DeleteRole errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				expectAllRoleDependencyRemovalsSucceed(m, aws.String("test"))
				m.EXPECT().DeleteRole(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteRoleError"))
			},
			want:    fmt.Errorf("DeleteRoleError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iamMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamMock)

			iamRoleOperator := NewIamRoleOperator(iamMock)

			err := iamRoleOperator.DeleteIamRole(tt.args.ctx, tt.args.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestIamRoleOperator_DeleteResourcesForIamRole(t *testing.T) {
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
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				expectAllRoleDependencyRemovalsSucceed(m, aws.String("PhysicalResourceId1"))
				m.EXPECT().DeleteRole(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
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
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("GetRoleError"))
			},
			want:    fmt.Errorf("GetRoleError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iamMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamMock)

			iamRoleOperator := NewIamRoleOperator(iamMock)
			iamRoleOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::IAM::Role"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := iamRoleOperator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
