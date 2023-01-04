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

type IBackupSDKClient interface {
	ListRecoveryPointsByBackupVault(ctx context.Context, params *backup.ListRecoveryPointsByBackupVaultInput, optFns ...func(*backup.Options)) (*backup.ListRecoveryPointsByBackupVaultOutput, error)
	DeleteRecoveryPoint(ctx context.Context, params *backup.DeleteRecoveryPointInput, optFns ...func(*backup.Options)) (*backup.DeleteRecoveryPointOutput, error)
	DeleteBackupVault(ctx context.Context, params *backup.DeleteBackupVaultInput, optFns ...func(*backup.Options)) (*backup.DeleteBackupVaultOutput, error)
	ListBackupVaults(ctx context.Context, params *backup.ListBackupVaultsInput, optFns ...func(*backup.Options)) (*backup.ListBackupVaultsOutput, error)
}

type Backup struct {
	client IBackupSDKClient
}

func NewBackup(client IBackupSDKClient) *Backup {
	return &Backup{
		client,
	}
}

func (backupClient *Backup) ListRecoveryPointsByBackupVault(ctx context.Context, backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	var nextToken *string
	recoveryPoints := []types.RecoveryPointByBackupVault{}

	for {
		select {
		case <-ctx.Done():
			return recoveryPoints, ctx.Err()
		default:
		}

		input := &backup.ListRecoveryPointsByBackupVaultInput{
			BackupVaultName: backupVaultName,
			NextToken:       nextToken,
		}

		output, err := backupClient.client.ListRecoveryPointsByBackupVault(ctx, input)
		if err != nil {
			return nil, err
		}
		recoveryPoints = append(recoveryPoints, output.RecoveryPoints...)

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}

	return recoveryPoints, nil
}

func (backupClient *Backup) DeleteRecoveryPoints(ctx context.Context, backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	for _, recoveryPoint := range recoveryPoints {
		if err := backupClient.DeleteRecoveryPoint(ctx, backupVaultName, recoveryPoint.RecoveryPointArn); err != nil {
			return err
		}
	}
	return nil
}

func (backupClient *Backup) DeleteRecoveryPoint(ctx context.Context, backupVaultName *string, recoveryPointArn *string) error {
	input := &backup.DeleteRecoveryPointInput{
		BackupVaultName:  backupVaultName,
		RecoveryPointArn: recoveryPointArn,
	}

	_, err := backupClient.client.DeleteRecoveryPoint(ctx, input)

	return err
}

func (backupClient *Backup) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	input := &backup.DeleteBackupVaultInput{
		BackupVaultName: backupVaultName,
	}

	_, err := backupClient.client.DeleteBackupVault(ctx, input)

	return err
}

func (backupClient *Backup) CheckBackupVaultExists(ctx context.Context, backupVaultName *string) (bool, error) {
	var nextToken *string

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		input := &backup.ListBackupVaultsInput{
			NextToken: nextToken,
		}

		output, err := backupClient.client.ListBackupVaults(ctx, input)
		if err != nil {
			return false, err
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
