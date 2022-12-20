package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
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

func (operator *BackupVaultOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, backupVault := range operator.resources {
		backupVault := backupVault
		sem.Acquire(ctx, 1)
		eg.Go(func() error {
			defer sem.Release(1)

			return operator.DeleteBackupVault(ctx, backupVault.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (operator *BackupVaultOperator) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	exists, err := operator.client.CheckBackupVaultExists(ctx, backupVaultName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	recoveryPoints, err := operator.client.ListRecoveryPointsByBackupVault(ctx, backupVaultName)
	if err != nil {
		return err
	}

	if len(recoveryPoints) > 0 {
		if err := operator.client.DeleteRecoveryPoints(ctx, backupVaultName, recoveryPoints); err != nil {
			return err
		}
	}

	if err := operator.client.DeleteBackupVault(ctx, backupVaultName); err != nil {
		return err
	}

	return nil
}
