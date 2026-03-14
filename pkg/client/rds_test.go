package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/goleak"
)

func TestRDS_CheckDBInstanceDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		dbInstanceId       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check db instance deletion protection enabled",
			args: args{
				ctx:          context.Background(),
				dbInstanceId: aws.String("db-instance-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBInstancesProtectionEnabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBInstancesOutput{
										DBInstances: []types.DBInstance{
											{
												DBInstanceIdentifier: aws.String("db-instance-1"),
												DeletionProtection:   aws.Bool(true),
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
			want:    true,
			wantErr: false,
		},
		{
			name: "check db instance deletion protection disabled",
			args: args{
				ctx:          context.Background(),
				dbInstanceId: aws.String("db-instance-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBInstancesProtectionDisabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBInstancesOutput{
										DBInstances: []types.DBInstance{
											{
												DBInstanceIdentifier: aws.String("db-instance-1"),
												DeletionProtection:   aws.Bool(false),
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
			want:    false,
			wantErr: false,
		},
		{
			name: "check db instance deletion protection with no instances",
			args: args{
				ctx:          context.Background(),
				dbInstanceId: aws.String("db-instance-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBInstancesEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBInstancesOutput{
										DBInstances: []types.DBInstance{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "check db instance deletion protection failure",
			args: args{
				ctx:          context.Background(),
				dbInstanceId: aws.String("db-instance-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBInstancesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBInstancesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeDBInstancesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			sdkClient := rds.NewFromConfig(cfg)
			rdsClient := NewRDS(sdkClient)

			got, err := rdsClient.CheckDBInstanceDeletionProtection(tt.args.ctx, tt.args.dbInstanceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
			}
		})
	}
}

func TestRDS_DisableDBInstanceDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		dbInstanceId       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "disable db instance deletion protection successfully",
			args: args{
				ctx:          context.Background(),
				dbInstanceId: aws.String("db-instance-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyDBInstanceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.ModifyDBInstanceOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: false,
		},
		{
			name: "disable db instance deletion protection failure",
			args: args{
				ctx:          context.Background(),
				dbInstanceId: aws.String("db-instance-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyDBInstanceErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.ModifyDBInstanceOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ModifyDBInstanceError")
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			sdkClient := rds.NewFromConfig(cfg)
			rdsClient := NewRDS(sdkClient)

			err = rdsClient.DisableDBInstanceDeletionProtection(tt.args.ctx, tt.args.dbInstanceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
			}
		})
	}
}

func TestRDS_CheckDBClusterDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		dbClusterId        *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check db cluster deletion protection enabled",
			args: args{
				ctx:         context.Background(),
				dbClusterId: aws.String("db-cluster-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBClustersProtectionEnabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBClustersOutput{
										DBClusters: []types.DBCluster{
											{
												DBClusterIdentifier: aws.String("db-cluster-1"),
												DeletionProtection:  aws.Bool(true),
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
			want:    true,
			wantErr: false,
		},
		{
			name: "check db cluster deletion protection disabled",
			args: args{
				ctx:         context.Background(),
				dbClusterId: aws.String("db-cluster-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBClustersProtectionDisabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBClustersOutput{
										DBClusters: []types.DBCluster{
											{
												DBClusterIdentifier: aws.String("db-cluster-1"),
												DeletionProtection:  aws.Bool(false),
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
			want:    false,
			wantErr: false,
		},
		{
			name: "check db cluster deletion protection with no clusters",
			args: args{
				ctx:         context.Background(),
				dbClusterId: aws.String("db-cluster-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBClustersEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBClustersOutput{
										DBClusters: []types.DBCluster{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "check db cluster deletion protection failure",
			args: args{
				ctx:         context.Background(),
				dbClusterId: aws.String("db-cluster-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeDBClustersErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.DescribeDBClustersOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeDBClustersError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			sdkClient := rds.NewFromConfig(cfg)
			rdsClient := NewRDS(sdkClient)

			got, err := rdsClient.CheckDBClusterDeletionProtection(tt.args.ctx, tt.args.dbClusterId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
			}
		})
	}
}

func TestRDS_DisableDBClusterDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		dbClusterId        *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "disable db cluster deletion protection successfully",
			args: args{
				ctx:         context.Background(),
				dbClusterId: aws.String("db-cluster-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyDBClusterMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.ModifyDBClusterOutput{},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: false,
		},
		{
			name: "disable db cluster deletion protection failure",
			args: args{
				ctx:         context.Background(),
				dbClusterId: aws.String("db-cluster-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyDBClusterErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &rds.ModifyDBClusterOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ModifyDBClusterError")
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadDefaultConfig(
				tt.args.ctx,
				config.WithRegion("us-east-1"),
				config.WithAPIOptions([]func(*middleware.Stack) error{tt.args.withAPIOptionsFunc}),
			)
			if err != nil {
				t.Fatal(err)
			}

			sdkClient := rds.NewFromConfig(cfg)
			rdsClient := NewRDS(sdkClient)

			err = rdsClient.DisableDBClusterDeletionProtection(tt.args.ctx, tt.args.dbClusterId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
			}
		})
	}
}
