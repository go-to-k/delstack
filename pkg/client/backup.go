//go:generate mockgen -source=./backup.go -destination=./backup_mock.go -package=client
package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
)

type IBackup interface {
	ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error)
	DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error
	DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error
	DeleteBackupVault(ctx context.Context, backupVaultName *string) error
	CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error)
}

var _ IBackup = (*Backup)(nil)

type Backup struct {
	client *backup.Client
}

func NewBackup(client *backup.Client) *Backup {
	return &Backup{
		client,
	}
}

func (b *Backup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	var nextToken *string
	recoveryPoints := []types.RecoveryPointByBackupVault{}

	for {
		select {
		case <-ctx.Done():
			return recoveryPoints, &ClientError{
				ResourceName: backupVaultName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &backup.ListRecoveryPointsByBackupVaultInput{
			BackupVaultName: backupVaultName,
			NextToken:       nextToken,
		}

		output, err := b.client.ListRecoveryPointsByBackupVault(ctx, input)
		if err != nil {
			return nil, &ClientError{
				ResourceName: backupVaultName,
				Err:          err,
			}
		}
		recoveryPoints = append(recoveryPoints, output.RecoveryPoints...)

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	return recoveryPoints, nil
}

func (b *Backup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	for _, recoveryPoint := range recoveryPoints {
		if err := b.DeleteRecoveryPoint(ctx, backupVaultName, recoveryPoint.RecoveryPointArn); err != nil {
			return err // return non wrapping error because already wrapped error in DeleteRecoveryPoint
		}
	}
	return nil
}

func (b *Backup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	input := &backup.DeleteRecoveryPointInput{
		BackupVaultName:  backupVaultName,
		RecoveryPointArn: recoveryPointArn,
	}

	_, err := b.client.DeleteRecoveryPoint(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: backupVaultName,
			Err:          err,
		}
	}
	return nil
}

func (b *Backup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	input := &backup.DeleteBackupVaultInput{
		BackupVaultName: backupVaultName,
	}

	_, err := b.client.DeleteBackupVault(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: backupVaultName,
			Err:          err,
		}
	}
	return nil
}

func (b *Backup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	var nextToken *string

	for {
		select {
		case <-ctx.Done():
			return false, &ClientError{
				ResourceName: backupVaultName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &backup.ListBackupVaultsInput{
			NextToken: nextToken,
		}

		output, err := b.client.ListBackupVaults(ctx, input)
		if err != nil {
			return false, &ClientError{
				ResourceName: backupVaultName,
				Err:          err,
			}
		}

		for _, vault := range output.BackupVaultList {
			if *vault.BackupVaultName == *backupVaultName {
				return true, nil
			}
		}

		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return false, nil
}
