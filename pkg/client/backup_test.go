package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/aws/smithy-go/middleware"
)

/*
	Test Cases
*/

func TestBackup_ListRecoveryPointsByBackupVault(t *testing.T) {
	type args struct {
		ctx                context.Context
		backupVaultName    *string
		withAPIOptionsFunc func(*middleware.Stack) error
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListRecoveryPointsByBackupVaultMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.ListRecoveryPointsByBackupVaultOutput{
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
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListRecoveryPointsByBackupVaultErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.ListRecoveryPointsByBackupVaultOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("operation error Backup: ListRecoveryPointsByBackupVault, ListRecoveryPointsByBackupVaultError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := backup.NewFromConfig(cfg)
			backupClient := NewBackup(client)

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
	type args struct {
		ctx                context.Context
		backupVaultName    *string
		recoveryPoints     []types.RecoveryPointByBackupVault
		withAPIOptionsFunc func(*middleware.Stack) error
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
						RecoveryPointArn: aws.String("RecoveryPointArn1"),
					},
					{
						RecoveryPointArn: aws.String("RecoveryPointArn2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRecoveryPointMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteRecoveryPointOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRecoveryPointEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteRecoveryPointOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
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
						RecoveryPointArn: aws.String("RecoveryPointArn1"),
					},
					{
						RecoveryPointArn: aws.String("RecoveryPointArn2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRecoveryPointErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteRecoveryPointOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteRecoveryPointError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error Backup: DeleteRecoveryPoint, DeleteRecoveryPointError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := backup.NewFromConfig(cfg)
			backupClient := NewBackup(client)

			err = backupClient.DeleteRecoveryPoints(tt.args.ctx, tt.args.backupVaultName, tt.args.recoveryPoints)
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
	type args struct {
		ctx                context.Context
		backupVaultName    *string
		recoveryPointArn   *string
		withAPIOptionsFunc func(*middleware.Stack) error
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRecoveryPointMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteRecoveryPointOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRecoveryPointErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteRecoveryPointOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteRecoveryPointError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error Backup: DeleteRecoveryPoint, DeleteRecoveryPointError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := backup.NewFromConfig(cfg)
			backupClient := NewBackup(client)

			err = backupClient.DeleteRecoveryPoint(tt.args.ctx, tt.args.backupVaultName, tt.args.recoveryPointArn)
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
	type args struct {
		ctx                context.Context
		backupVaultName    *string
		withAPIOptionsFunc func(*middleware.Stack) error
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteBackupVaultMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteBackupVaultOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete backup vault failure",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteBackupVaultErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.DeleteBackupVaultOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteBackupVaultError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error Backup: DeleteBackupVault, DeleteBackupVaultError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := backup.NewFromConfig(cfg)
			backupClient := NewBackup(client)

			err = backupClient.DeleteBackupVault(tt.args.ctx, tt.args.backupVaultName)
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

	type args struct {
		ctx                context.Context
		backupVaultName    *string
		withAPIOptionsFunc func(*middleware.Stack) error
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBackupVaultsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.ListBackupVaultsOutput{
										BackupVaultList: []types.BackupVaultListMember{
											{
												BackupVaultName: aws.String("test"),
											},
											{
												BackupVaultName: aws.String("test2"),
											},
										},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBackupVaultsNotExistMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.ListBackupVaultsOutput{
										BackupVaultList: []types.BackupVaultListMember{
											{
												BackupVaultName: aws.String("test0"),
											},
											{
												BackupVaultName: aws.String("test2"),
											},
										},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
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
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBackupVaultsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &backup.ListBackupVaultsOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListBackupVaultsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("operation error Backup: ListBackupVaults, ListBackupVaultsError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("ap-northeast-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			client := backup.NewFromConfig(cfg)
			backupClient := NewBackup(client)

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
