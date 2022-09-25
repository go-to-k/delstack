package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*BackupVaultOperator)(nil)

type BackupVaultOperator struct {
	client    client.IBackup
	resources []*types.StackResourceSummary
}

func NewBackupVaultOperator(client client.IBackup) *BackupVaultOperator {
	return &BackupVaultOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *BackupVaultOperator) AddResource(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *BackupVaultOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *BackupVaultOperator) DeleteResources() error {
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(option.ConcurrencyNum))

	for _, backupVault := range operator.resources {
		backupVault := backupVault
		eg.Go(func() error {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			return operator.DeleteBackupVault(backupVault.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (operator *BackupVaultOperator) DeleteBackupVault(backupVaultName *string) error {
	exists, err := operator.client.CheckBackupVaultExists(backupVaultName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	recoveryPoints, err := operator.client.ListRecoveryPointsByBackupVault(backupVaultName)
	if err != nil {
		return err
	}

	if len(recoveryPoints) > 0 {
		if err := operator.client.DeleteRecoveryPoints(backupVaultName, recoveryPoints); err != nil {
			return err
		}
	}

	if err := operator.client.DeleteBackupVault(backupVaultName); err != nil {
		return err
	}

	return nil
}
