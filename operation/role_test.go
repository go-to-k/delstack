package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
)

var _ client.IIam = (*mockIam)(nil)
var _ client.IIam = (*allErrorMockIam)(nil)
var _ client.IIam = (*deleteRoleErrorMockIam)(nil)
var _ client.IIam = (*listAttachedRolePoliciesErrorMockIam)(nil)
var _ client.IIam = (*detachRolePoliciesErrorMockIam)(nil)

/*
	Mocks for client
*/
type mockIam struct{}

func NewMockIam() *mockIam {
	return &mockIam{}
}

func (m *mockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *mockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *mockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *mockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

type allErrorMockIam struct{}

func NewAllErrorMockIam() *allErrorMockIam {
	return &allErrorMockIam{}
}

func (m *allErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *allErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *allErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *allErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePolicyError")
}

type deleteRoleErrorMockIam struct{}

func NewDeleteRoleErrorMockIam() *deleteRoleErrorMockIam {
	return &deleteRoleErrorMockIam{}
}

func (m *deleteRoleErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *deleteRoleErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *deleteRoleErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *deleteRoleErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

type listAttachedRolePoliciesErrorMockIam struct{}

func NewListAttachedRolePoliciesErrorMockIam() *listAttachedRolePoliciesErrorMockIam {
	return &listAttachedRolePoliciesErrorMockIam{}
}

func (m *listAttachedRolePoliciesErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *listAttachedRolePoliciesErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *listAttachedRolePoliciesErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *listAttachedRolePoliciesErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

type detachRolePoliciesErrorMockIam struct{}

func NewDetachRolePoliciesErrorMockIam() *detachRolePoliciesErrorMockIam {
	return &detachRolePoliciesErrorMockIam{}
}

func (m *detachRolePoliciesErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *detachRolePoliciesErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{
		{
			PolicyArn:  aws.String("PolicyArn1"),
			PolicyName: aws.String("PolicyName1"),
		},
		{
			PolicyArn:  aws.String("PolicyArn2"),
			PolicyName: aws.String("PolicyName2"),
		},
	}
	return output, nil
}

func (m *detachRolePoliciesErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *detachRolePoliciesErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

/*
	Test Cases
*/
func TestDeleteRole(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockIam()
	allErrorMock := NewAllErrorMockIam()
	deleteRoleErrorMock := NewDeleteRoleErrorMockIam()
	listAttachedRolePoliciesErrorMock := NewListAttachedRolePoliciesErrorMockIam()
	detachRolePoliciesErrorMock := NewDetachRolePoliciesErrorMockIam()

	type args struct {
		ctx      context.Context
		roleName *string
		client   client.IIam
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete role successfully",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure for all errors",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   allErrorMock,
			},
			want:    fmt.Errorf("ListAttachedRolePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete role failure for delete role errors",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   deleteRoleErrorMock,
			},
			want:    fmt.Errorf("DeleteRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for list attached role policies errors",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   listAttachedRolePoliciesErrorMock,
			},
			want:    fmt.Errorf("ListAttachedRolePoliciesError"),
			wantErr: true,
		},
		{
			name: "delete role failure for detach role errors",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   detachRolePoliciesErrorMock,
			},
			want:    fmt.Errorf("DetachRolePoliciesError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamOperator := NewRoleOperator(tt.args.client)

			err := iamOperator.DeleteRole(tt.args.roleName)
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

func TestDeleteResourcesForIam(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockIam()
	allErrorMock := NewAllErrorMockIam()

	type args struct {
		ctx    context.Context
		client client.IIam
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx:    ctx,
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    ctx,
				client: allErrorMock,
			},
			want:    fmt.Errorf("ListAttachedRolePoliciesError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamOperator := NewRoleOperator(tt.args.client)
			iamOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String(resourcetype.IAM_ROLE),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := iamOperator.DeleteResources()
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
