package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
)

var _ IBackupSDKClient = (*MockBackupSDKClient)(nil)
var _ IBackupSDKClient = (*ErrorMockBackupSDKClient)(nil)
var _ IBackupSDKClient = (*NotExistsMockForListBackupVaultsBackupSDKClient)(nil)

/*
	Mocks for SDK Client
*/
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

/*
	Test Cases
*/
func TestBackup_ListRecoveryPointsByBackupVault(t *testing.T) {
	t.Parallel()

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
				ctx:             context.Background(),
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
				ctx:             context.Background(),
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
			t.Parallel()

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
	t.Parallel()

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
				ctx:             context.Background(),
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
				ctx:             context.Background(),
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
				ctx:             context.Background(),
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
			t.Parallel()

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
	t.Parallel()

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
				ctx:              context.Background(),
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
				ctx:              context.Background(),
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
			t.Parallel()

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
	t.Parallel()

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
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				client:          errorMock,
			},
			want:    fmt.Errorf("DeleteBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

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
				ctx:             context.Background(),
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
				ctx:             context.Background(),
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
				ctx:             context.Background(),
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
			t.Parallel()

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
