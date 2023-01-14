package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/go-to-k/delstack/pkg/client"
)

type BackupVaultOperatorFactory struct {
	config aws.Config
}

func NewBackupVaultOperatorFactory(config aws.Config) *BackupVaultOperatorFactory {
	return &BackupVaultOperatorFactory{config}
}

func (f *BackupVaultOperatorFactory) CreateBackupVaultOperator() *BackupVaultOperator {
	return NewBackupVaultOperator(
		f.createBackupClient(),
	)
}

func (f *BackupVaultOperatorFactory) createBackupClient() *client.Backup {
	sdkBackupClient := backup.NewFromConfig(f.config)

	return client.NewBackup(
		sdkBackupClient,
	)
}
