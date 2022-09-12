package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-to-k/delstack/logger"
)

type Backup struct {
	client *backup.Client
}

func NewBackup(config aws.Config) *Backup {
	client := backup.NewFromConfig(config)
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
			logger.Logger.Fatal().Msgf("Error: failed list the Recovery Points: %v", err.Error())
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
	if err != nil {
		logger.Logger.Fatal().Msgf("Error: failed delete the Recovery Point, %v", err.Error())
		return err
	}

	return nil
}

func (backupClient *Backup) DeleteBackupVault(backupVaultName *string) error {
	input := &backup.DeleteBackupVaultInput{
		BackupVaultName: backupVaultName,
	}

	_, err := backupClient.client.DeleteBackupVault(context.TODO(), input)
	if err != nil {
		logger.Logger.Fatal().Msgf("Error: failed delete the Backup Vault, %v", err.Error())
		return err
	}

	return nil
}
