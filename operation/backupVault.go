package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
)

func DeleteBackupVaults(config aws.Config, resources []types.StackResourceSummary) error {
	// TODO: Concurrency Delete
	backupVaultClient := client.NewBackup(config)
	for _, backupVault := range resources {
		err := DeleteBackupVault(backupVaultClient, backupVault.PhysicalResourceId)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteBackupVault(backupVaultClient *client.Backup, backupVaultName *string) error {
	recoveryPoints, err := backupVaultClient.ListRecoveryPointsByBackupVault(backupVaultName)
	if err != nil {
		return err
	}

	if err := backupVaultClient.DeleteRecoveryPoints(backupVaultName, recoveryPoints); err != nil {
		return err
	}

	if err := backupVaultClient.DeleteBackupVault(backupVaultName); err != nil {
		return err
	}

	return nil
}
