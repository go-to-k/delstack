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

func (m *MockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *MockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *MockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type AllErrorMockIam struct{}

func NewAllErrorMockIam() *AllErrorMockIam {
	return &AllErrorMockIam{}
}

func (m *AllErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *AllErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *AllErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *AllErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePolicyError")
}

func (m *AllErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, fmt.Errorf("GetRoleError")
}

type DeleteRoleErrorMockIam struct{}

func NewDeleteRoleErrorMockIam() *DeleteRoleErrorMockIam {
	return &DeleteRoleErrorMockIam{}
}

func (m *DeleteRoleErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return fmt.Errorf("DeleteRoleError")
}

func (m *DeleteRoleErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *DeleteRoleErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *DeleteRoleErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DeleteRoleErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type ListAttachedRolePoliciesErrorMockIam struct{}

func NewListAttachedRolePoliciesErrorMockIam() *ListAttachedRolePoliciesErrorMockIam {
	return &ListAttachedRolePoliciesErrorMockIam{}
}

func (m *ListAttachedRolePoliciesErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	return nil, fmt.Errorf("ListAttachedRolePoliciesError")
}

func (m *ListAttachedRolePoliciesErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *ListAttachedRolePoliciesErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type DetachRolePoliciesErrorMockIam struct{}

func NewDetachRolePoliciesErrorMockIam() *DetachRolePoliciesErrorMockIam {
	return &DetachRolePoliciesErrorMockIam{}
}

func (m *DetachRolePoliciesErrorMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *DetachRolePoliciesErrorMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesError")
}

func (m *DetachRolePoliciesErrorMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type DetachRolePoliciesErrorAfterZeroLengthMockIam struct{}

func NewDetachRolePoliciesErrorAfterZeroLengthMockIam() *DetachRolePoliciesErrorAfterZeroLengthMockIam {
	return &DetachRolePoliciesErrorAfterZeroLengthMockIam{}
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
	output := []types.AttachedPolicy{}
	return output, nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return fmt.Errorf("DetachRolePoliciesErrorAfterZeroLength")
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *DetachRolePoliciesErrorAfterZeroLengthMockIam) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type CheckRoleExistsErrorMockRole struct{}

func NewCheckRoleExistsErrorMockRole() *CheckRoleExistsErrorMockRole {
	return &CheckRoleExistsErrorMockRole{}
}

func (m *CheckRoleExistsErrorMockRole) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *CheckRoleExistsErrorMockRole) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleExistsErrorMockRole) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, fmt.Errorf("GetRoleError")
}

type CheckRoleNotExistsMockRole struct{}

func NewCheckRoleNotExistsMockRole() *CheckRoleNotExistsMockRole {
	return &CheckRoleNotExistsMockRole{}
}

func (m *CheckRoleNotExistsMockRole) DeleteRole(ctx context.Context, roleName *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) ListAttachedRolePolicies(ctx context.Context, roleName *string) ([]types.AttachedPolicy, error) {
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

func (m *CheckRoleNotExistsMockRole) DetachRolePolicies(ctx context.Context, roleName *string, policies []types.AttachedPolicy, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) DetachRolePolicy(ctx context.Context, roleName *string, PolicyArn *string, sleepTimeSec int) error {
	return nil
}

func (m *CheckRoleNotExistsMockRole) CheckRoleExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, nil
}

/*
	Test Cases
*/
func TestRoleOperator_DeleteRole(t *testing.T) {
	io.NewLogger(false)
	ctx := context.Background()
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

			err := iamOperator.DeleteRole(tt.args.ctx, tt.args.roleName)
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
	io.NewLogger(false)
	ctx := context.Background()
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
				ResourceType:       aws.String("AWS::IAM::Role"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := iamOperator.DeleteResources(tt.args.ctx)
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
