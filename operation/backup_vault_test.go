package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
)

var _ client.IBackup = (*MockBackup)(nil)
var _ client.IBackup = (*AllErrorMockBackup)(nil)
var _ client.IBackup = (*ListRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteBackupVaultErrorMockBackup)(nil)

/*
	Mocks for client
*/
type MockBackup struct{}

func NewMockBackup() *MockBackup {
	return &MockBackup{}
}

func (m *MockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
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

func (m *MockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *MockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *MockBackup) DeleteBackupVault(backupVaultName *string) error {
	return nil
}

type AllErrorMockBackup struct{}

func NewAllErrorMockBackup() *AllErrorMockBackup {
	return &AllErrorMockBackup{}
}

func (m *AllErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *AllErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsError")
}

func (m *AllErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return fmt.Errorf("DeleteRecoveryPointError")
}

func (m *AllErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return fmt.Errorf("DeleteBackupVaultError")
}

type ListRecoveryPointsErrorMockBackup struct{}

func NewListRecoveryPointsErrorMockBackup() *ListRecoveryPointsErrorMockBackup {
	return &ListRecoveryPointsErrorMockBackup{}
}

func (m *ListRecoveryPointsErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *ListRecoveryPointsErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *ListRecoveryPointsErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *ListRecoveryPointsErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return nil
}

type DeleteRecoveryPointsErrorMockBackup struct{}

func NewDeleteRecoveryPointsErrorMockBackup() *DeleteRecoveryPointsErrorMockBackup {
	return &DeleteRecoveryPointsErrorMockBackup{}
}

func (m *DeleteRecoveryPointsErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, nil
}

func (m *DeleteRecoveryPointsErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsError")
}

func (m *DeleteRecoveryPointsErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *DeleteRecoveryPointsErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return nil
}

type DeleteBackupVaultErrorMockBackup struct{}

func NewDeleteBackupVaultErrorMockBackup() *DeleteBackupVaultErrorMockBackup {
	return &DeleteBackupVaultErrorMockBackup{}
}

func (m *DeleteBackupVaultErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, nil
}

func (m *DeleteBackupVaultErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *DeleteBackupVaultErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *DeleteBackupVaultErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return fmt.Errorf("DeleteBackupVaultError")
}

/*
	Test Cases
*/
func TestBackupVaultOperator_DeleteBackupVault(t *testing.T) {
	logger.NewLogger(false)
	ctx := context.TODO()
	mock := NewMockBackup()
	allErrorMock := NewAllErrorMockBackup()
	listRecoveryPointsErrorMock := NewListRecoveryPointsErrorMockBackup()
	deleteRecoveryPointsErrorMock := NewDeleteRecoveryPointsErrorMockBackup()
	deleteBackupVaultErrorMock := NewDeleteBackupVaultErrorMockBackup()

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
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for all errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          allErrorMock,
			},
			want:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			wantErr: true,
		},
		{
			name: "delete backup vault failure for list recovery points errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          listRecoveryPointsErrorMock,
			},
			want:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			wantErr: true,
		},
		{
			name: "delete backup vault failure for delete recovery points errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          deleteRecoveryPointsErrorMock,
			},
			want:    fmt.Errorf("DeleteRecoveryPointsError"),
			wantErr: true,
		},
		{
			name: "delete backup vault failure for delete backup vault errors",
			args: args{
				ctx:             ctx,
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

			err := backupOperator.DeleteBackupVault(tt.args.backupVaultName)
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
	logger.NewLogger(false)
	ctx := context.TODO()
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
			want:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupOperator := NewBackupVaultOperator(tt.args.client)
			backupOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String(resourcetype.BACKUP_VAULT),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := backupOperator.DeleteResources()
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
