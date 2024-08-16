package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestBackupVaultOperator_DeleteBackupVault(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx             context.Context
		backupVaultName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIBackup)
		want          error
		wantErr       bool
	}{
		{
			name: "delete backup vault successfully",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListRecoveryPointsByBackupVault(gomock.Any(), aws.String("test")).Return(
					[]types.RecoveryPointByBackupVault{
						{
							BackupVaultName: aws.String("BackupVaultName1"),
							BackupVaultArn:  aws.String("BackupVaultArn1"),
						},
						{
							BackupVaultName: aws.String("BackupVaultName2"),
							BackupVaultArn:  aws.String("BackupVaultArn2"),
						},
					}, nil)
				m.EXPECT().DeleteRecoveryPoints(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteBackupVault(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for check backup vault exists errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("ListBackupVaultsError"))
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
		{
			name: "delete backup vault successfully for backup vault not exists",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for list recovery points errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListRecoveryPointsByBackupVault(gomock.Any(), aws.String("test")).Return(nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError"))
			},
			want:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			wantErr: true,
		},
		{
			name: "delete backup vault failure for delete recovery points errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListRecoveryPointsByBackupVault(gomock.Any(), aws.String("test")).Return(
					[]types.RecoveryPointByBackupVault{
						{
							BackupVaultName: aws.String("BackupVaultName1"),
							BackupVaultArn:  aws.String("BackupVaultArn1"),
						},
						{
							BackupVaultName: aws.String("BackupVaultName2"),
							BackupVaultArn:  aws.String("BackupVaultArn2"),
						},
					}, nil)
				m.EXPECT().DeleteRecoveryPoints(gomock.Any(), aws.String("test"), gomock.Any()).Return(fmt.Errorf("DeleteRecoveryPointsError"))
			},
			want:    fmt.Errorf("DeleteRecoveryPointsError"),
			wantErr: true,
		},
		{
			name: "delete backup vault successfully for delete recovery points errors after zero length",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListRecoveryPointsByBackupVault(gomock.Any(), aws.String("test")).Return([]types.RecoveryPointByBackupVault{}, nil)
				m.EXPECT().DeleteBackupVault(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for delete backup vault errors",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().ListRecoveryPointsByBackupVault(gomock.Any(), aws.String("test")).Return(
					[]types.RecoveryPointByBackupVault{
						{
							BackupVaultName: aws.String("BackupVaultName1"),
							BackupVaultArn:  aws.String("BackupVaultArn1"),
						},
						{
							BackupVaultName: aws.String("BackupVaultName2"),
							BackupVaultArn:  aws.String("BackupVaultArn2"),
						},
					}, nil)
				m.EXPECT().DeleteRecoveryPoints(gomock.Any(), aws.String("test"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteBackupVault(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteBackupVaultError"))
			},
			want:    fmt.Errorf("DeleteBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			backupMock := client.NewMockIBackup(ctrl)
			tt.prepareMockFn(backupMock)

			backupOperator := NewBackupVaultOperator(backupMock)

			err := backupOperator.DeleteBackupVault(tt.args.ctx, tt.args.backupVaultName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestBackupVaultOperator_DeleteResourcesForBackupVault(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIBackup)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().ListRecoveryPointsByBackupVault(gomock.Any(), aws.String("PhysicalResourceId1")).Return(
					[]types.RecoveryPointByBackupVault{
						{
							BackupVaultName: aws.String("BackupVaultName1"),
							BackupVaultArn:  aws.String("BackupVaultArn1"),
						},
						{
							BackupVaultName: aws.String("BackupVaultName2"),
							BackupVaultArn:  aws.String("BackupVaultArn2"),
						},
					}, nil)
				m.EXPECT().DeleteRecoveryPoints(gomock.Any(), aws.String("PhysicalResourceId1"), gomock.Any()).Return(nil)
				m.EXPECT().DeleteBackupVault(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIBackup) {
				m.EXPECT().CheckBackupVaultExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("ListBackupVaultsError"))
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			backupMock := client.NewMockIBackup(ctrl)
			tt.prepareMockFn(backupMock)

			backupOperator := NewBackupVaultOperator(backupMock)

			backupOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::Backup::BackupVault"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := backupOperator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
