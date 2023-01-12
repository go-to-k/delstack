package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
)

var _ client.IBackup = (*MockBackup)(nil)
var _ client.IBackup = (*AllErrorMockBackup)(nil)
var _ client.IBackup = (*ListRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorMockBackup)(nil)
var _ client.IBackup = (*DeleteRecoveryPointsErrorAfterZeroLengthMockBackup)(nil)
var _ client.IBackup = (*DeleteBackupVaultErrorMockBackup)(nil)
var _ client.IBackup = (*CheckBackupVaultExistsErrorMockBackup)(nil)
var _ client.IBackup = (*CheckBackupVaultNotExistsMockBackup)(nil)

/*
	Test Cases
*/

func TestBackupVaultOperator_DeleteBackupVault(t *testing.T) {
	io.NewLogger(false)
	ctx := context.Background()
	mock := NewMockBackup()
	allErrorMock := NewAllErrorMockBackup()
	listRecoveryPointsErrorMock := NewListRecoveryPointsErrorMockBackup()
	deleteRecoveryPointsErrorMock := NewDeleteRecoveryPointsErrorMockBackup()
	deleteRecoveryPointsErrorAfterZeroLengthMock := NewDeleteRecoveryPointsErrorAfterZeroLengthMockBackup()
	deleteBackupVaultErrorMock := NewDeleteBackupVaultErrorMockBackup()
	checkBackupVaultExistsErrorMock := NewCheckBackupVaultExistsErrorMockBackup()
	checkBackupVaultNotExistsMock := NewCheckBackupVaultNotExistsMockBackup()

	type args struct {
		ctx             context.Context
		backupVaultName *string
		client          client.IBackup
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete backup vault successfully",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for all errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          allErrorMock,
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
		{
			name: "delete bucket failure for check bucket exists errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          checkBackupVaultExistsErrorMock,
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for bucket not exists",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          checkBackupVaultNotExistsMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for list recovery points errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          listRecoveryPointsErrorMock,
			},
			want:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			wantErr: true,
		},
		{
			name: "delete backup vault failure for delete recovery points errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          deleteRecoveryPointsErrorMock,
			},
			want:    fmt.Errorf("DeleteRecoveryPointsError"),
			wantErr: true,
		},
		{
			name: "delete backup vault successfully for delete recovery points errors after zero length",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          deleteRecoveryPointsErrorAfterZeroLengthMock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure for delete backup vault errors",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          deleteBackupVaultErrorMock,
			},
			want:    fmt.Errorf("DeleteBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupOperator := NewBackupVaultOperator(tt.args.client)

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
	ctx := context.Background()
	mock := NewMockBackup()
	allErrorMock := NewAllErrorMockBackup()

	type args struct {
		ctx    context.Context
		client client.IBackup
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx:    ctx,
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    ctx,
				client: allErrorMock,
			},
			want:    fmt.Errorf("ListBackupVaultsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupOperator := NewBackupVaultOperator(tt.args.client)
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
