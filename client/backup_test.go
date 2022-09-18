package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/go-to-k/delstack/logger"
)

var _ IBackupSDKClient = (*mockBackupSDKClient)(nil)
var _ IBackupSDKClient = (*errorMockBackupSDKClient)(nil)

/*
	Mocks for SDK Client
*/
type mockBackupSDKClient struct{}

func NewMockBackupSDKClient() *mockBackupSDKClient {
	return &mockBackupSDKClient{}
}

func (m *mockBackupSDKClient) ListRecoveryPointsByBackupVault(ctx context.Context, params *backup.ListRecoveryPointsByBackupVaultInput, optFns ...func(*backup.Options)) (*backup.ListRecoveryPointsByBackupVaultOutput, error) {
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

func (m *mockBackupSDKClient) DeleteRecoveryPoint(ctx context.Context, params *backup.DeleteRecoveryPointInput, optFns ...func(*backup.Options)) (*backup.DeleteRecoveryPointOutput, error) {
	return nil, nil
}

func (m *mockBackupSDKClient) DeleteBackupVault(ctx context.Context, params *backup.DeleteBackupVaultInput, optFns ...func(*backup.Options)) (*backup.DeleteBackupVaultOutput, error) {
	return nil, nil
}

type errorMockBackupSDKClient struct{}

func NewErrorMockBackupSDKClient() *errorMockBackupSDKClient {
	return &errorMockBackupSDKClient{}
}

func (m *errorMockBackupSDKClient) ListRecoveryPointsByBackupVault(ctx context.Context, params *backup.ListRecoveryPointsByBackupVaultInput, optFns ...func(*backup.Options)) (*backup.ListRecoveryPointsByBackupVaultOutput, error) {
	return nil, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
}

func (m *errorMockBackupSDKClient) DeleteRecoveryPoint(ctx context.Context, params *backup.DeleteRecoveryPointInput, optFns ...func(*backup.Options)) (*backup.DeleteRecoveryPointOutput, error) {
	return nil, fmt.Errorf("DeleteRecoveryPointError")
}

func (m *errorMockBackupSDKClient) DeleteBackupVault(ctx context.Context, params *backup.DeleteBackupVaultInput, optFns ...func(*backup.Options)) (*backup.DeleteBackupVaultOutput, error) {
	return nil, fmt.Errorf("DeleteBackupVaultError")
}

/*
	Test Cases
*/
func TestListRecoveryPointsByBackupVault(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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
			name: "list recovery points successfully",
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
			name: "list recovery points failure",
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

			output, err := backupClient.ListRecoveryPointsByBackupVault(tt.args.backupVaultName)
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

func TestDeleteRecoveryPoints(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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

			err := backupClient.DeleteRecoveryPoints(tt.args.backupVaultName, tt.args.recoveryPoints)
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

func TestDeleteRecoveryPoint(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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

			err := backupClient.DeleteRecoveryPoint(tt.args.backupVaultName, tt.args.recoveryPointArn)
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

func TestDeleteBackupVault(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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

			err := backupClient.DeleteBackupVault(tt.args.backupVaultName)
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
