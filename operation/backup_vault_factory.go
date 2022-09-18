package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/go-to-k/delstack/client"
)

type BackupVaultOperatorFactory struct {
	config aws.Config
}

func NewBackupVaultOperatorFactory(config aws.Config) *BackupVaultOperatorFactory {
	return &BackupVaultOperatorFactory{config}
}

func (factory *BackupVaultOperatorFactory) CreateBackupVaultOperator() *BackupVaultOperator {
	return NewBackupVaultOperator(
		factory.createBackupClient(),
	)
}

func (factory *BackupVaultOperatorFactory) createBackupClient() *client.Backup {
	sdkBackupClient := backup.NewFromConfig(factory.config)

	return client.NewBackup(
		factory.config,
		sdkBackupClient,
	)
}
