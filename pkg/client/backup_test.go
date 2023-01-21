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

type tokenKey struct{}

func getNextTokenForInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	switch v := in.Parameters.(type) {
	case *backup.ListRecoveryPointsByBackupVaultInput:
		ctx = middleware.WithStackValue(ctx, tokenKey{}, v.NextToken)
	case *backup.ListBackupVaultsInput:
		ctx = middleware.WithStackValue(ctx, tokenKey{}, v.NextToken)
	}
	return next.HandleInitialize(ctx, in)
}

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
			name: "list recovery points by backup vault successfully",
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
			name: "list recovery points by backup vault failure",
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
		{
			name: "list recovery points by backup vault with next token successfully",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextToken",
							getNextTokenForInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListRecoveryPointsByBackupVaultWithNextTokenMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								token := middleware.GetStackValue(ctx, tokenKey{}).(*string)

								var nextToken *string
								var recoveryPoints []types.RecoveryPointByBackupVault
								if token == nil {
									nextToken = aws.String("NextToken")
									recoveryPoints = []types.RecoveryPointByBackupVault{
										{
											BackupVaultName: aws.String("BackupVaultName1"),
											BackupVaultArn:  aws.String("BackupVaultArn1"),
										},
										{
											BackupVaultName: aws.String("BackupVaultName2"),
											BackupVaultArn:  aws.String("BackupVaultArn2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListRecoveryPointsByBackupVaultOutput{
											NextToken:      nextToken,
											RecoveryPoints: recoveryPoints,
										},
									}, middleware.Metadata{}, nil
								} else {
									recoveryPoints = []types.RecoveryPointByBackupVault{
										{
											BackupVaultName: aws.String("BackupVaultName3"),
											BackupVaultArn:  aws.String("BackupVaultArn3"),
										},
										{
											BackupVaultName: aws.String("BackupVaultName4"),
											BackupVaultArn:  aws.String("BackupVaultArn4"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListRecoveryPointsByBackupVaultOutput{
											NextToken:      nextToken,
											RecoveryPoints: recoveryPoints,
										},
									}, middleware.Metadata{}, nil
								}
							},
						),
						middleware.Before,
					)
					return err
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
					{
						BackupVaultArn:  aws.String("BackupVaultArn3"),
						BackupVaultName: aws.String("BackupVaultName3"),
					},
					{
						BackupVaultArn:  aws.String("BackupVaultArn4"),
						BackupVaultName: aws.String("BackupVaultName4"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list recovery points by backup vault with next token failure",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextToken",
							getNextTokenForInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListRecoveryPointsByBackupVaultWithNextTokenErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								token := middleware.GetStackValue(ctx, tokenKey{}).(*string)

								var nextToken *string
								var recoveryPoints []types.RecoveryPointByBackupVault
								if token == nil {
									nextToken = aws.String("NextToken")
									recoveryPoints = []types.RecoveryPointByBackupVault{
										{
											BackupVaultName: aws.String("BackupVaultName1"),
											BackupVaultArn:  aws.String("BackupVaultArn1"),
										},
										{
											BackupVaultName: aws.String("BackupVaultName2"),
											BackupVaultArn:  aws.String("BackupVaultArn2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListRecoveryPointsByBackupVaultOutput{
											NextToken:      nextToken,
											RecoveryPoints: recoveryPoints,
										},
									}, middleware.Metadata{}, nil
								} else {
									return middleware.FinalizeOutput{
										Result: &backup.ListRecoveryPointsByBackupVaultOutput{},
									}, middleware.Metadata{}, fmt.Errorf("ListRecoveryPointsByBackupVaultError")
								}
							},
						),
						middleware.Before,
					)
					return err
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
			name: "check backup vault exists",
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
			name: "check backup vault does not exist",
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
			name: "check backup vault exists failure",
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
		{
			name: "check backup vault exists with next token",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextToken",
							getNextTokenForInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBackupVaultsWithNextTokenMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								token := middleware.GetStackValue(ctx, tokenKey{}).(*string)

								var nextToken *string
								var backupVaultListMember []types.BackupVaultListMember
								if token == nil {
									nextToken = aws.String("NextToken")
									backupVaultListMember = []types.BackupVaultListMember{
										{
											BackupVaultName: aws.String("test0"),
										},
										{
											BackupVaultName: aws.String("test2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListBackupVaultsOutput{
											NextToken:       nextToken,
											BackupVaultList: backupVaultListMember,
										},
									}, middleware.Metadata{}, nil
								} else {
									backupVaultListMember = []types.BackupVaultListMember{
										{
											BackupVaultName: aws.String("test"),
										},
										{
											BackupVaultName: aws.String("test3"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListBackupVaultsOutput{
											NextToken:       nextToken,
											BackupVaultList: backupVaultListMember,
										},
									}, middleware.Metadata{}, nil
								}
							},
						),
						middleware.Before,
					)
					return err
				},
			},
			want: want{
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check backup vault does not exist with next token",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextToken",
							getNextTokenForInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBackupVaultsWithNextTokenMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								token := middleware.GetStackValue(ctx, tokenKey{}).(*string)

								var nextToken *string
								var backupVaultListMember []types.BackupVaultListMember
								if token == nil {
									nextToken = aws.String("NextToken")
									backupVaultListMember = []types.BackupVaultListMember{
										{
											BackupVaultName: aws.String("test0"),
										},
										{
											BackupVaultName: aws.String("test2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListBackupVaultsOutput{
											NextToken:       nextToken,
											BackupVaultList: backupVaultListMember,
										},
									}, middleware.Metadata{}, nil
								} else {
									backupVaultListMember = []types.BackupVaultListMember{
										{
											BackupVaultName: aws.String("test1"),
										},
										{
											BackupVaultName: aws.String("test3"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListBackupVaultsOutput{
											NextToken:       nextToken,
											BackupVaultList: backupVaultListMember,
										},
									}, middleware.Metadata{}, nil
								}
							},
						),
						middleware.Before,
					)
					return err
				},
			},
			want: want{
				exists: false,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check backup vault exists failure",
			args: args{
				ctx:             context.Background(),
				backupVaultName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextToken",
							getNextTokenForInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListBackupVaultsWithNextTokenErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								token := middleware.GetStackValue(ctx, tokenKey{}).(*string)

								var nextToken *string
								var backupVaultListMember []types.BackupVaultListMember
								if token == nil {
									nextToken = aws.String("NextToken")
									backupVaultListMember = []types.BackupVaultListMember{
										{
											BackupVaultName: aws.String("test0"),
										},
										{
											BackupVaultName: aws.String("test2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &backup.ListBackupVaultsOutput{
											NextToken:       nextToken,
											BackupVaultList: backupVaultListMember,
										},
									}, middleware.Metadata{}, nil
								} else {
									return middleware.FinalizeOutput{
										Result: &backup.ListBackupVaultsOutput{},
									}, middleware.Metadata{}, fmt.Errorf("ListBackupVaultsError")
								}
							},
						),
						middleware.Before,
					)
					return err
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
