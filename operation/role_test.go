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

var _ client.IIam = (*MockIam)(nil)
var _ client.IIam = (*AllErrorMockIam)(nil)
var _ client.IIam = (*DeleteRoleErrorMockIam)(nil)
var _ client.IIam = (*ListAttachedRolePoliciesErrorMockIam)(nil)
var _ client.IIam = (*DetachRolePoliciesErrorMockIam)(nil)
var _ client.IIam = (*DetachRolePoliciesErrorAfterZeroLengthMockIam)(nil)
var _ client.IIam = (*CheckRoleExistsErrorMockRole)(nil)
var _ client.IIam = (*CheckRoleNotExistsMockRole)(nil)

/*
	Mocks for client
*/
type MockIam struct{}

func NewMockIam() *MockIam {
	return &MockIam{}
}

func (m *MockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *MockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) CheckRoleExists(repositoryName *string) (bool, error) {
	return true, nil
}

type AllErrorMockIam struct{}

func NewAllErrorMockIam() *AllErrorMockIam {
	return &AllErrorMockIam{}
}

func (m *AllErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *AllErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *AllErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *AllErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePolicyError")
}

func (m *AllErrorMockIam) CheckRoleExists(repositoryName *string) (bool, error) {
	return false, fmt.Errorf("GetRoleError")
}

type DeleteRoleErrorMockIam struct{}

func NewDeleteRoleErrorMockIam() *DeleteRoleErrorMockIam {
	return &DeleteRoleErrorMockIam{}
}

func (m *DeleteRoleErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *DeleteRoleErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *DeleteRoleErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *DeleteRoleErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DeleteRoleErrorMockIam) CheckRoleExists(repositoryName *string) (bool, error) {
	return true, nil
}

type ListAttachedRolePoliciesErrorMockIam struct{}

func NewListAttachedRolePoliciesErrorMockIam() *ListAttachedRolePoliciesErrorMockIam {
	return &ListAttachedRolePoliciesErrorMockIam{}
}

func (m *ListAttachedRolePoliciesErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *ListAttachedRolePoliciesErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) CheckRoleExists(repositoryName *string) (bool, error) {
	return true, nil
}

type DetachRolePoliciesErrorMockIam struct{}

func NewDetachRolePoliciesErrorMockIam() *DetachRolePoliciesErrorMockIam {
	return &DetachRolePoliciesErrorMockIam{}
}

func (m *DetachRolePoliciesErrorMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *DetachRolePoliciesErrorMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *DetachRolePoliciesErrorMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorMockIam) CheckRoleExists(repositoryName *string) (bool, error) {
	return true, nil
}

type DetachRolePoliciesErrorAfterZeroLengthMockIam struct{}

func NewDetachRolePoliciesErrorAfterZeroLengthMockIam() *DetachRolePoliciesErrorAfterZeroLengthMockIam {
	return &DetachRolePoliciesErrorAfterZeroLengthMockIam{}
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{}
	return output, nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesErrorAfterZeroLength")
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) CheckRoleExists(repositoryName *string) (bool, error) {
	return true, nil
}

type CheckRoleExistsErrorMockRole struct{}

func NewCheckRoleExistsErrorMockRole() *CheckRoleExistsErrorMockRole {
	return &CheckRoleExistsErrorMockRole{}
}

func (m *CheckRoleExistsErrorMockRole) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *CheckRoleExistsErrorMockRole) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) CheckRoleExists(repositoryName *string) (bool, error) {
	return false, fmt.Errorf("GetRoleError")
}

type CheckRoleNotExistsMockRole struct{}

func NewCheckRoleNotExistsMockRole() *CheckRoleNotExistsMockRole {
	return &CheckRoleNotExistsMockRole{}
}

func (m *CheckRoleNotExistsMockRole) DeleteRole(roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) ListAttachedRolePolicies(roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *CheckRoleNotExistsMockRole) DetachRolePolicies(roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) DetachRolePolicy(roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) CheckRoleExists(repositoryName *string) (bool, error) {
	return false, nil
}

/*
	Test Cases
*/
func TestRoleOperator_DeleteRole(t *testing.T) {
	logger.NewLogger(false)
	ctx := context.TODO()
	mock := NewMockIam()
	allErrorMock := NewAllErrorMockIam()
	deleteRoleErrorMock := NewDeleteRoleErrorMockIam()
	listAttachedRolePoliciesErrorMock := NewListAttachedRolePoliciesErrorMockIam()
	detachRolePoliciesErrorMock := NewDetachRolePoliciesErrorMockIam()
	detachRolePoliciesErrorAfterZeroLengthMock := NewDetachRolePoliciesErrorAfterZeroLengthMockIam()
	checkRoleExistsErrorMock := NewCheckRoleExistsErrorMockRole()
	checkRoleNotExistsMock := NewCheckRoleNotExistsMockRole()

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
			want:    fmt.Errorf("GetRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for check role exists errors",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   checkRoleExistsErrorMock,
			},
			want:    fmt.Errorf("GetRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for check role not exists",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   checkRoleNotExistsMock,
			},
			want:    nil,
			wantErr: false,
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
		{
			name: "delete role successfully for detach role errors after zero length",
			args: args{
				ctx:      ctx,
				roleName: aws.String("test"),
				client:   detachRolePoliciesErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
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

func TestRoleOperator_DeleteResourcesForIam(t *testing.T) {
	logger.NewLogger(false)
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
			want:    fmt.Errorf("GetRoleError"),
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
