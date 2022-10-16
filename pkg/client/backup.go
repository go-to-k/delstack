package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
)

type IBackup interface {
	ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error)
	DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error
	DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error
	DeleteBackupVault(backupVaultName *string) error
	CheckBackupVaultExists(backupVaultName *string) (bool, error)
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

func (backupClient *Backup) ListRecoveryPointsByBackupVault(backupVaultName *string) ([]types.RecoveryPointByBackupVault, error) {
	var nextToken *string
	recoveryPoints := []types.RecoveryPointByBackupVault{}

	for {
		input := &backup.ListRecoveryPointsByBackupVaultInput{
			BackupVaultName: backupVaultName,
			NextToken:       nextToken,
		}

		output, err := backupClient.client.ListRecoveryPointsByBackupVault(context.TODO(), input)
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

func (backupClient *Backup) DeleteRecoveryPoints(backupVaultName *string, recoveryPoints []types.RecoveryPointByBackupVault) error {
	for _, recoveryPoint := range recoveryPoints {
		if err := backupClient.DeleteRecoveryPoint(backupVaultName, recoveryPoint.RecoveryPointArn); err != nil {
			return err
		}
	}
	return nil
}

func (backupClient *Backup) DeleteRecoveryPoint(backupVaultName *string, recoveryPointArn *string) error {
	input := &backup.DeleteRecoveryPointInput{
		BackupVaultName:  backupVaultName,
		RecoveryPointArn: recoveryPointArn,
	}

	_, err := backupClient.client.DeleteRecoveryPoint(context.TODO(), input)

	return err
}

func (backupClient *Backup) DeleteBackupVault(backupVaultName *string) error {
	input := &backup.DeleteBackupVaultInput{
		BackupVaultName: backupVaultName,
	}

	_, err := backupClient.client.DeleteBackupVault(context.TODO(), input)

	return err
}

func (backupClient *Backup) CheckBackupVaultExists(backupVaultName *string) (bool, error) {
	var nextToken *string

	for {
		input := &backup.ListBackupVaultsInput{
			NextToken: nextToken,
		}

		output, err := backupClient.client.ListBackupVaults(context.TODO(), input)
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
