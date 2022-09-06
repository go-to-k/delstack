package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"golang.org/x/sync/errgroup"
)

var _ Operator = (*BackupVaultOperator)(nil)

type BackupVaultOperator struct {
	client    *client.Backup
	resources []*types.StackResourceSummary
}

func NewBackupVaultOperator(config aws.Config) *BackupVaultOperator {
	client := client.NewBackup(config)
	return &BackupVaultOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *BackupVaultOperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *BackupVaultOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *BackupVaultOperator) DeleteResources() error {
	var eg errgroup.Group
	var semaphore = make(chan struct{}, CONCURRENCY_NUM)

	for _, backupVault := range operator.resources {
		backupVault := backupVault
		eg.Go(func() error {
			semaphore <- struct{}{}

			if err := operator.DeleteBackupVault(backupVault.PhysicalResourceId); err != nil {
				return err
			}
			<-semaphore

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (operator *BackupVaultOperator) DeleteBackupVault(backupVaultName *string) error {
	recoveryPoints, err := operator.client.ListRecoveryPointsByBackupVault(backupVaultName)
	if err != nil {
		return err
	}

	if err := operator.client.DeleteRecoveryPoints(backupVaultName, recoveryPoints); err != nil {
		return err
	}

	if err := operator.client.DeleteBackupVault(backupVaultName); err != nil {
		return err
	}

	return nil
}
