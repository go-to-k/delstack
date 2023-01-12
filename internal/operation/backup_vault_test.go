package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
)

var _ client.IBackup = (*MockBackup)(nil)
var _ client.IBackup = (*AllErrorMockBackup)(nil)
var _ client.IBackup = (*ListRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorAfterZeroLengthMockBackup)(nil)
var _ client.IBackup = (*DeleteBackupVaultErrorMockBackup)(nil)
var _ client.IBackup = (*CheckBackupVaultExistsErrorMockBackup)(nil)
var _ client.IBackup = (*CheckBackupVaultNotExistsMockBackup)(nil)

/*
	Mocks for client
*/
type MockBackup struct{}

func NewMockBackup() *MockBackup {
	return &MockBackup{}
}

func (m *MockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	output := []types.RecoveryPointByBackupVault{
		{
			BackupVaultName: aws.String("BackupVaultName1"),
			BackupVaultArn:  aws.String("BackupVaultArn1"),
		},
		{
			BackupVaultName: aws.String("BackupVaultName2"),
			BackupVaultArn:  aws.String("BackupVaultArn2"),
		},
	}
	return output, nil
}

func (m *MockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *MockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *MockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return nil
}

func (m *MockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return true, nil
}

type AllErrorMockBackup struct{}

func NewAllErrorMockBackup() *AllErrorMockBackup {
	return &AllErrorMockBackup{}
}

func (m *AllErrorMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *AllErrorMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsError")
}

func (m *AllErrorMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return fmt.Errorf("DeleteRecoveryPointError")
}

func (m *AllErrorMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return fmt.Errorf("DeleteBackupVaultError")
}

func (m *AllErrorMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return false, fmt.Errorf("ListBackupVaultsError")
}

type ListRecoveryPointsErrorMockBackup struct{}

func NewListRecoveryPointsErrorMockBackup() *ListRecoveryPointsErrorMockBackup {
	return &ListRecoveryPointsErrorMockBackup{}
}

func (m *ListRecoveryPointsErrorMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *ListRecoveryPointsErrorMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *ListRecoveryPointsErrorMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *ListRecoveryPointsErrorMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return nil
}

func (m *ListRecoveryPointsErrorMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return true, nil
}

type DeleteRecoveryPointsErrorMockBackup struct{}

func NewDeleteRecoveryPointsErrorMockBackup() *DeleteRecoveryPointsErrorMockBackup {
	return &DeleteRecoveryPointsErrorMockBackup{}
}

func (m *DeleteRecoveryPointsErrorMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	output := []types.RecoveryPointByBackupVault{
		{
			BackupVaultName: aws.String("BackupVaultName1"),
			BackupVaultArn:  aws.String("BackupVaultArn1"),
		},
		{
			BackupVaultName: aws.String("BackupVaultName2"),
			BackupVaultArn:  aws.String("BackupVaultArn2"),
		},
	}
	return output, nil
}

func (m *DeleteRecoveryPointsErrorMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsError")
}

func (m *DeleteRecoveryPointsErrorMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *DeleteRecoveryPointsErrorMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return nil
}

func (m *DeleteRecoveryPointsErrorMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return true, nil
}

type DeleteRecoveryPointsErrorAfterZeroLengthMockBackup struct{}

func NewDeleteRecoveryPointsErrorAfterZeroLengthMockBackup() *DeleteRecoveryPointsErrorAfterZeroLengthMockBackup {
	return &DeleteRecoveryPointsErrorAfterZeroLengthMockBackup{}
}

func (m *DeleteRecoveryPointsErrorAfterZeroLengthMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	output := []types.RecoveryPointByBackupVault{}
	return output, nil
}

func (m *DeleteRecoveryPointsErrorAfterZeroLengthMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsErrorAfterZeroLength")
}

func (m *DeleteRecoveryPointsErrorAfterZeroLengthMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *DeleteRecoveryPointsErrorAfterZeroLengthMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return nil
}

func (m *DeleteRecoveryPointsErrorAfterZeroLengthMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return true, nil
}

type DeleteBackupVaultErrorMockBackup struct{}

func NewDeleteBackupVaultErrorMockBackup() *DeleteBackupVaultErrorMockBackup {
	return &DeleteBackupVaultErrorMockBackup{}
}

func (m *DeleteBackupVaultErrorMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	output := []types.RecoveryPointByBackupVault{
		{
			BackupVaultName: aws.String("BackupVaultName1"),
			BackupVaultArn:  aws.String("BackupVaultArn1"),
		},
		{
			BackupVaultName: aws.String("BackupVaultName2"),
			BackupVaultArn:  aws.String("BackupVaultArn2"),
		},
	}
	return output, nil
}

func (m *DeleteBackupVaultErrorMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *DeleteBackupVaultErrorMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *DeleteBackupVaultErrorMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return fmt.Errorf("DeleteBackupVaultError")
}

func (m *DeleteBackupVaultErrorMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return true, nil
}

type CheckBackupVaultExistsErrorMockBackup struct{}

func NewCheckBackupVaultExistsErrorMockBackup() *CheckBackupVaultExistsErrorMockBackup {
	return &CheckBackupVaultExistsErrorMockBackup{}
}

func (m *CheckBackupVaultExistsErrorMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	output := []types.RecoveryPointByBackupVault{
		{
			BackupVaultName: aws.String("BackupVaultName1"),
			BackupVaultArn:  aws.String("BackupVaultArn1"),
		},
		{
			BackupVaultName: aws.String("BackupVaultName2"),
			BackupVaultArn:  aws.String("BackupVaultArn2"),
		},
	}
	return output, nil
}

func (m *CheckBackupVaultExistsErrorMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *CheckBackupVaultExistsErrorMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *CheckBackupVaultExistsErrorMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return nil
}

func (m *CheckBackupVaultExistsErrorMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return false, fmt.Errorf("ListBackupVaultsError")
}

type CheckBackupVaultNotExistsMockBackup struct{}

func NewCheckBackupVaultNotExistsMockBackup() *CheckBackupVaultNotExistsMockBackup {
	return &CheckBackupVaultNotExistsMockBackup{}
}

func (m *CheckBackupVaultNotExistsMockBackup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	output := []types.RecoveryPointByBackupVault{
		{
			BackupVaultName: aws.String("BackupVaultName1"),
			BackupVaultArn:  aws.String("BackupVaultArn1"),
		},
		{
			BackupVaultName: aws.String("BackupVaultName2"),
			BackupVaultArn:  aws.String("BackupVaultArn2"),
		},
	}
	return output, nil
}

func (m *CheckBackupVaultNotExistsMockBackup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *CheckBackupVaultNotExistsMockBackup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *CheckBackupVaultNotExistsMockBackup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	return nil
}

func (m *CheckBackupVaultNotExistsMockBackup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	return false, nil
}

/*
	Test Cases
*/
func TestBackupVaultOperator_DeleteBackupVault(t *testing.T) {
	io.NewLogger(false)
	mock := NewMockBackup()
	allErrorMock := NewAllErrorMockBackup()
	listRecoveryPointsErrorMock := NewListRecoveryPointsErrorMockBackup()
	deleteRecoveryPointsErrorMock := NewDeleteRecoveryPointsErrorMockBackup()
	deleteRecoveryPointsErrorAfterZeroLengthMock := NewDeleteRecoveryPointsErrorAfterZeroLengthMockBackup()
	deleteBackupVaultErrorMock := NewDeleteBackupVaultErrorMockBackup()
	checkBackupVaultExistsErrorMock := NewCheckBackupVaultExistsErrorMockBackup()
	checkBackupVaultNotExistsMock := NewCheckBackupVaultNotExistsMockBackup()

	type args struct {
		ctx             context.Context
		backupVaultName *string
		client          client.IBackup
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete backup vault successfully",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for all errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          allErrorMock,
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for check bucket exists errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          checkBackupVaultExistsErrorMock,
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for bucket not exists",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          checkBackupVaultNotExistsMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for list recovery points errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          listRecoveryPointsErrorMock,
			},
			want:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			wantErr: true,
		},
		{
			name: "delete backup vault failure for delete recovery points errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          deleteRecoveryPointsErrorMock,
			},
			want:    fmt.Errorf("DeleteRecoveryPointsError"),
			wantErr: true,
		},
		{
			name: "delete backup vault successfully for delete recovery points errors after zero length",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          deleteRecoveryPointsErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for delete backup vault errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          deleteBackupVaultErrorMock,
			},
			want:    fmt.Errorf("DeleteBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupOperator := NewBackupVaultOperator(tt.args.client)

			err := backupOperator.DeleteBackupVault(tt.args.ctx, tt.args.backupVaultName)
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

func TestBackupVaultOperator_DeleteResourcesForBackupVault(t *testing.T) {
	io.NewLogger(false)
	mock := NewMockBackup()
	allErrorMock := NewAllErrorMockBackup()

	type args struct {
		ctx    context.Context
		client client.IBackup
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
				ctx:    context.Background(),
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    context.Background(),
				client: allErrorMock,
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupOperator := NewBackupVaultOperator(tt.args.client)
			backupOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::Backup::BackupVault"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := backupOperator.DeleteResources(tt.args.ctx)
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
