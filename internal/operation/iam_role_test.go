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
	gomock "github.com/golang/mock/gomock"
)

/*
	Test Cases
*/

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
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test")).Return(
					[]types.AttachedPolicy{
						{
							PolicyArn:  aws.String("PolicyArn1"),
							PolicyName: aws.String("PolicyName1"),
						},
						{
							PolicyArn:  aws.String("PolicyArn2"),
							PolicyName: aws.String("PolicyName2"),
						},
					}, nil)
				m.EXPECT().DetachRolePolicies(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteRole(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure for check role exists errors",
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
			name: "delete role failure for check role not exists",
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
			name: "delete role failure for list attached role policies errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test")).Return(nil, fmt.Errorf("ListAttachedRolePoliciesError"))
			},
			want:    fmt.Errorf("ListAttachedRolePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete role failure for detach role errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test")).Return(
					[]types.AttachedPolicy{
						{
							PolicyArn:  aws.String("PolicyArn1"),
							PolicyName: aws.String("PolicyName1"),
						},
						{
							PolicyArn:  aws.String("PolicyArn2"),
							PolicyName: aws.String("PolicyName2"),
						},
					}, nil)
				m.EXPECT().DetachRolePolicies(gomock.Any(), aws.String("test"), gomock.Any()).Return(fmt.Errorf("DetachRolePoliciesError"))
			},
			want:    fmt.Errorf("DetachRolePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete role successfully for ListAttachedRolePolicies with zero length",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test")).Return([]types.AttachedPolicy{}, nil)
				m.EXPECT().DeleteRole(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure for delete role errors",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckRoleExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("test")).Return(
					[]types.AttachedPolicy{
						{
							PolicyArn:  aws.String("PolicyArn1"),
							PolicyName: aws.String("PolicyName1"),
						},
						{
							PolicyArn:  aws.String("PolicyArn2"),
							PolicyName: aws.String("PolicyName2"),
						},
					}, nil)
				m.EXPECT().DetachRolePolicies(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
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
				m.EXPECT().ListAttachedRolePolicies(gomock.Any(), aws.String("PhysicalResourceId1")).Return(
					[]types.AttachedPolicy{
						{
							PolicyArn:  aws.String("PolicyArn1"),
							PolicyName: aws.String("PolicyName1"),
						},
						{
							PolicyArn:  aws.String("PolicyArn2"),
							PolicyName: aws.String("PolicyName2"),
						},
					}, nil)
				m.EXPECT().DetachRolePolicies(gomock.Any(), aws.String("PhysicalResourceId1"), gomock.Any()).Return(nil)
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
