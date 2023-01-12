package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
)

/*
	Test Cases
*/

func TestBackup_ListRecoveryPointsByBackupVault(t *testing.T) {
	ctx := context.Background()
	mock := NewMockBackupSDKClient()
	errorMock := NewErrorMockBackupSDKClient()

	type args struct {
		ctx             context.Context
		backupVaultName *string
		client          IBackupSDKClient
	}

	type want struct {
		output []types.RecoveryPointByBackupVault
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "check backup vault exists successfully",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          mock,
			},
			want: want{
				output: []types.RecoveryPointByBackupVault{
					{
						BackupVaultArn:  aws.String("BackupVaultArn1"),
						BackupVaultName: aws.String("BackupVaultName1"),
					},
					{
						BackupVaultArn:  aws.String("BackupVaultArn2"),
						BackupVaultName: aws.String("BackupVaultName2"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "check backup vault exists failure",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          errorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("ListRecoveryPointsByBackupVaultError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupClient := NewBackup(tt.args.client)

			output, err := backupClient.ListRecoveryPointsByBackupVault(tt.args.ctx, tt.args.backupVaultName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}

func TestBackup_DeleteRecoveryPoints(t *testing.T) {
	ctx := context.Background()
	mock := NewMockBackupSDKClient()
	errorMock := NewErrorMockBackupSDKClient()

	type args struct {
		ctx             context.Context
		backupVaultName *string
		recoveryPoints  []types.RecoveryPointByBackupVault
		client          IBackupSDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete recovery points successfully",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				recoveryPoints: []types.RecoveryPointByBackupVault{
					{
						BackupVaultArn:  aws.String("BackupVaultArn1"),
						BackupVaultName: aws.String("BackupVaultName1"),
					},
					{
						BackupVaultArn:  aws.String("BackupVaultArn2"),
						BackupVaultName: aws.String("BackupVaultName2"),
					},
				},
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete empty recovery points successfully",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				recoveryPoints:  []types.RecoveryPointByBackupVault{},
				client:          mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete recovery points failure",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				recoveryPoints: []types.RecoveryPointByBackupVault{
					{
						BackupVaultArn:  aws.String("BackupVaultArn1"),
						BackupVaultName: aws.String("BackupVaultName1"),
					},
					{
						BackupVaultArn:  aws.String("BackupVaultArn2"),
						BackupVaultName: aws.String("BackupVaultName2"),
					},
				},
				client: errorMock,
			},
			want:    fmt.Errorf("DeleteRecoveryPointError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupClient := NewBackup(tt.args.client)

			err := backupClient.DeleteRecoveryPoints(tt.args.ctx, tt.args.backupVaultName, tt.args.recoveryPoints)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestBackup_DeleteRecoveryPoint(t *testing.T) {
	ctx := context.Background()
	mock := NewMockBackupSDKClient()
	errorMock := NewErrorMockBackupSDKClient()

	type args struct {
		ctx              context.Context
		backupVaultName  *string
		recoveryPointArn *string
		client           IBackupSDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete recovery point successfully",
			args: args{
				ctx:              ctx,
				backupVaultName:  aws.String("test"),
				recoveryPointArn: aws.String("test"),
				client:           mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete recovery point failure",
			args: args{
				ctx:              ctx,
				backupVaultName:  aws.String("test"),
				recoveryPointArn: aws.String("test"),
				client:           errorMock,
			},
			want:    fmt.Errorf("DeleteRecoveryPointError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupClient := NewBackup(tt.args.client)

			err := backupClient.DeleteRecoveryPoint(tt.args.ctx, tt.args.backupVaultName, tt.args.recoveryPointArn)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestBackup_DeleteBackupVault(t *testing.T) {
	ctx := context.Background()
	mock := NewMockBackupSDKClient()
	errorMock := NewErrorMockBackupSDKClient()

	type args struct {
		ctx             context.Context
		backupVaultName *string
		client          IBackupSDKClient
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
			name: "delete backup vault failure",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          errorMock,
			},
			want:    fmt.Errorf("DeleteBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupClient := NewBackup(tt.args.client)

			err := backupClient.DeleteBackupVault(tt.args.ctx, tt.args.backupVaultName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestBackup_CheckBackupVaultExists(t *testing.T) {
	ctx := context.Background()
	mock := NewMockBackupSDKClient()
	errorMock := NewErrorMockBackupSDKClient()
	notExitsMock := NewNotExistsMockForListBackupVaultsBackupSDKClient()

	type args struct {
		ctx             context.Context
		backupVaultName *string
		client          IBackupSDKClient
	}

	type want struct {
		exists bool
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "check backup vault for backup vault exists",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          mock,
			},
			want: want{
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check backup vault for backup vault do not exist",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          notExitsMock,
			},
			want: want{
				exists: false,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check bucket exists failure",
			args: args{
				ctx:             ctx,
				backupVaultName: aws.String("test"),
				client:          errorMock,
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("ListBackupVaultsError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			backupClient := NewBackup(tt.args.client)

			output, err := backupClient.CheckBackupVaultExists(tt.args.ctx, tt.args.backupVaultName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.exists) {
				t.Errorf("output = %#v, want %#v", output, tt.want.exists)
			}
		})
	}
}
