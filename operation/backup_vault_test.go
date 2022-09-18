package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
)

var _ client.IBackup = (*mockBackup)(nil)
var _ client.IBackup = (*allErrorMockBackup)(nil)
var _ client.IBackup = (*listRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*deleteRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*deleteBackupVaultErrorMockBackup)(nil)

/*
	Mocks for client
*/
type mockBackup struct{}

func NewMockBackup() *mockBackup {
	return &mockBackup{}
}

func (m *mockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
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

func (m *mockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *mockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *mockBackup) DeleteBackupVault(backupVaultName *string) error {
	return nil
}

type allErrorMockBackup struct{}

func NewAllErrorMockBackup() *allErrorMockBackup {
	return &allErrorMockBackup{}
}

func (m *allErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *allErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsError")
}

func (m *allErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return fmt.Errorf("DeleteRecoveryPointError")
}

func (m *allErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return fmt.Errorf("DeleteBackupVaultError")
}

type listRecoveryPointsErrorMockBackup struct{}

func NewListRecoveryPointsErrorMockBackup() *listRecoveryPointsErrorMockBackup {
	return &listRecoveryPointsErrorMockBackup{}
}

func (m *listRecoveryPointsErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *listRecoveryPointsErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *listRecoveryPointsErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *listRecoveryPointsErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return nil
}

type deleteRecoveryPointsErrorMockBackup struct{}

func NewDeleteRecoveryPointsErrorMockBackup() *deleteRecoveryPointsErrorMockBackup {
	return &deleteRecoveryPointsErrorMockBackup{}
}

func (m *deleteRecoveryPointsErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, nil
}

func (m *deleteRecoveryPointsErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return fmt.Errorf("DeleteRecoveryPointsError")
}

func (m *deleteRecoveryPointsErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *deleteRecoveryPointsErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return nil
}

type deleteBackupVaultErrorMockBackup struct{}

func NewDeleteBackupVaultErrorMockBackup() *deleteBackupVaultErrorMockBackup {
	return &deleteBackupVaultErrorMockBackup{}
}

func (m *deleteBackupVaultErrorMockBackup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	return nil, nil
}

func (m *deleteBackupVaultErrorMockBackup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	return nil
}

func (m *deleteBackupVaultErrorMockBackup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	return nil
}

func (m *deleteBackupVaultErrorMockBackup) DeleteBackupVault(backupVaultName *string) error {
	return fmt.Errorf("DeleteBackupVaultError")
}

/*
	Test Cases
*/
func TestDeleteBackupVault(t *testing.T) {
	logger.NewLogger()
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
