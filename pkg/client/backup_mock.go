package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
)

/*
	Mocks for SDK Client
*/

var _ IBackupSDKClient = (*MockBackupSDKClient)(nil)
var _ IBackupSDKClient = (*ErrorMockBackupSDKClient)(nil)
var _ IBackupSDKClient = (*NotExistsMockForListBackupVaultsBackupSDKClient)(nil)

type MockBackupSDKClient struct{}

func NewMockBackupSDKClient() *MockBackupSDKClient {
	return &MockBackupSDKClient{}
}

func (m *MockBackupSDKClient) ListRecoveryPointsByBackupVault(ctx context.Context, params *backup.ListRecoveryPointsByBackupVaultInput, optFns ...func(*backup.Options)) (*backup.ListRecoveryPointsByBackupVaultOutput, error) {
	output := &backup.ListRecoveryPointsByBackupVaultOutput{
		RecoveryPoints: []types.RecoveryPointByBackupVault{
			{
				BackupVaultName: aws.String("BackupVaultName1"),
				BackupVaultArn:  aws.String("BackupVaultArn1"),
			},
			{
				BackupVaultName: aws.String("BackupVaultName2"),
				BackupVaultArn:  aws.String("BackupVaultArn2"),
			},
		},
	}
	return output, nil
}

func (m *MockBackupSDKClient) DeleteRecoveryPoint(ctx context.Context, params *backup.DeleteRecoveryPointInput, optFns ...func(*backup.Options)) (*backup.DeleteRecoveryPointOutput, error) {
	return nil, nil
}

func (m *MockBackupSDKClient) DeleteBackupVault(ctx context.Context, params *backup.DeleteBackupVaultInput, optFns ...func(*backup.Options)) (*backup.DeleteBackupVaultOutput, error) {
	return nil, nil
}

func (m *MockBackupSDKClient) ListBackupVaults(ctx context.Context, params *backup.ListBackupVaultsInput, optFns ...func(*backup.Options)) (*backup.ListBackupVaultsOutput, error) {
	output := &backup.ListBackupVaultsOutput{
		BackupVaultList: []types.BackupVaultListMember{
			{
				BackupVaultName: aws.String("test"),
			},
			{
				BackupVaultName: aws.String("test2"),
			},
		},
	}
	return output, nil
}

type ErrorMockBackupSDKClient struct{}

func NewErrorMockBackupSDKClient() *ErrorMockBackupSDKClient {
	return &ErrorMockBackupSDKClient{}
}

func (m *ErrorMockBackupSDKClient) ListRecoveryPointsByBackupVault(ctx context.Context, params *backup.ListRecoveryPointsByBackupVaultInput, optFns ...func(*backup.Options)) (*backup.ListRecoveryPointsByBackupVaultOutput, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *ErrorMockBackupSDKClient) DeleteRecoveryPoint(ctx context.Context, params *backup.DeleteRecoveryPointInput, optFns ...func(*backup.Options)) (*backup.DeleteRecoveryPointOutput, error) {
	return nil, fmt.Errorf("DeleteRecoveryPointError")
}

func (m *ErrorMockBackupSDKClient) DeleteBackupVault(ctx context.Context, params *backup.DeleteBackupVaultInput, optFns ...func(*backup.Options)) (*backup.DeleteBackupVaultOutput, error) {
	return nil, fmt.Errorf("DeleteBackupVaultError")
}

func (m *ErrorMockBackupSDKClient) ListBackupVaults(ctx context.Context, params *backup.ListBackupVaultsInput, optFns ...func(*backup.Options)) (*backup.ListBackupVaultsOutput, error) {
	return nil, fmt.Errorf("ListBackupVaultsError")
}

type NotExistsMockForListBackupVaultsBackupSDKClient struct{}

func NewNotExistsMockForListBackupVaultsBackupSDKClient() *NotExistsMockForListBackupVaultsBackupSDKClient {
	return &NotExistsMockForListBackupVaultsBackupSDKClient{}
}

func (m *NotExistsMockForListBackupVaultsBackupSDKClient) ListRecoveryPointsByBackupVault(ctx context.Context, params *backup.ListRecoveryPointsByBackupVaultInput, optFns ...func(*backup.Options)) (*backup.ListRecoveryPointsByBackupVaultOutput, error) {
	output := &backup.ListRecoveryPointsByBackupVaultOutput{
		RecoveryPoints: []types.RecoveryPointByBackupVault{
			{
				BackupVaultName: aws.String("BackupVaultName1"),
				BackupVaultArn:  aws.String("BackupVaultArn1"),
			},
			{
				BackupVaultName: aws.String("BackupVaultName2"),
				BackupVaultArn:  aws.String("BackupVaultArn2"),
			},
		},
	}
	return output, nil
}

func (m *NotExistsMockForListBackupVaultsBackupSDKClient) DeleteRecoveryPoint(ctx context.Context, params *backup.DeleteRecoveryPointInput, optFns ...func(*backup.Options)) (*backup.DeleteRecoveryPointOutput, error) {
	return nil, nil
}

func (m *NotExistsMockForListBackupVaultsBackupSDKClient) DeleteBackupVault(ctx context.Context, params *backup.DeleteBackupVaultInput, optFns ...func(*backup.Options)) (*backup.DeleteBackupVaultOutput, error) {
	return nil, nil
}

func (m *NotExistsMockForListBackupVaultsBackupSDKClient) ListBackupVaults(ctx context.Context, params *backup.ListBackupVaultsInput, optFns ...func(*backup.Options)) (*backup.ListBackupVaultsOutput, error) {
	output := &backup.ListBackupVaultsOutput{
		BackupVaultList: []types.BackupVaultListMember{
			{
				BackupVaultName: aws.String("test0"),
			},
			{
				BackupVaultName: aws.String("test2"),
			},
		},
	}
	return output, nil
}
