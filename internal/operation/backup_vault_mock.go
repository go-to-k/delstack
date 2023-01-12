package operation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-to-k/delstack/pkg/client"
)

/*
	Mocks for client
*/

var _ client.IBackup = (*MockBackup)(nil)
var _ client.IBackup = (*AllErrorMockBackup)(nil)
var _ client.IBackup = (*ListRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorAfterZeroLengthMockBackup)(nil)
var _ client.IBackup = (*DeleteBackupVaultErrorMockBackup)(nil)
var _ client.IBackup = (*CheckBackupVaultExistsErrorMockBackup)(nil)
var _ client.IBackup = (*CheckBackupVaultNotExistsMockBackup)(nil)

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
