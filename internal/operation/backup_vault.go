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

func (o *BackupVaultOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *BackupVaultOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *BackupVaultOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, backupVault := range o.resources {
		backupVault := backupVault
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteBackupVault(ctx, backupVault.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *BackupVaultOperator) DeleteBackupVault(ctx context.Context, backupVaultName *string) error {
	exists, err := o.client.CheckBackupVaultExists(ctx, backupVaultName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	recoveryPoints, err := o.client.ListRecoveryPointsByBackupVault(ctx, backupVaultName)
	if err != nil {
		return err
	}

	if len(recoveryPoints) > 0 {
		if err := o.client.DeleteRecoveryPoints(ctx, backupVaultName, recoveryPoints); err != nil {
			return err
		}
	}

	if err := o.client.DeleteBackupVault(ctx, backupVaultName); err != nil {
		return err
	}

	return nil
}
