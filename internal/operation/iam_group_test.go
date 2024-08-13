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

func TestIamGroupOperator_DeleteIamGroup(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx       context.Context
		GroupName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIIam)
		want          error
		wantErr       bool
	}{
		{
			name: "delete Group successfully",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedGroupPolicies(gomock.Any(), aws.String("test")).Return(
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
				m.EXPECT().DetachGroupPolicies(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteGroup(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete Group failure for check Group exists errors",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("GetGroupError"))
			},
			want:    fmt.Errorf("GetGroupError"),
			wantErr: true,
		},
		{
			name: "delete Group failure for check Group not exists",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete Group failure for list attached Group policies errors",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedGroupPolicies(gomock.Any(), aws.String("test")).Return(nil, fmt.Errorf("ListAttachedGroupPoliciesError"))
			},
			want:    fmt.Errorf("ListAttachedGroupPoliciesError"),
			wantErr: true,
		},
		{
			name: "delete Group failure for detach Group errors",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedGroupPolicies(gomock.Any(), aws.String("test")).Return(
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
				m.EXPECT().DetachGroupPolicies(gomock.Any(), aws.String("test"), gomock.Any()).Return(fmt.Errorf("DetachGroupPoliciesError"))
			},
			want:    fmt.Errorf("DetachGroupPoliciesError"),
			wantErr: true,
		},
		{
			name: "delete Group successfully for ListAttachedGroupPolicies with zero length",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedGroupPolicies(gomock.Any(), aws.String("test")).Return([]types.AttachedPolicy{}, nil)
				m.EXPECT().DeleteGroup(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete Group failure for delete Group errors",
			args: args{
				ctx:       context.Background(),
				GroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIIam) {
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListAttachedGroupPolicies(gomock.Any(), aws.String("test")).Return(
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
				m.EXPECT().DetachGroupPolicies(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteGroup(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteGroupError"))
			},
			want:    fmt.Errorf("DeleteGroupError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iamMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamMock)

			iamGroupOperator := NewIamGroupOperator(iamMock)

			err := iamGroupOperator.DeleteIamGroup(tt.args.ctx, tt.args.GroupName)
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

func TestIamGroupOperator_DeleteResourcesForIamGroup(t *testing.T) {
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
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().ListAttachedGroupPolicies(gomock.Any(), aws.String("PhysicalResourceId1")).Return(
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
				m.EXPECT().DetachGroupPolicies(gomock.Any(), aws.String("PhysicalResourceId1"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteGroup(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
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
				m.EXPECT().CheckGroupExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("GetGroupError"))
			},
			want:    fmt.Errorf("GetGroupError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			iamMock := client.NewMockIIam(ctrl)
			tt.prepareMockFn(iamMock)

			iamGroupOperator := NewIamGroupOperator(iamMock)
			iamGroupOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::IAM::Group"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := iamGroupOperator.DeleteResources(tt.args.ctx)
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
