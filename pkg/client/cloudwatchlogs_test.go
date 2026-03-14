package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/goleak"
)

func TestCloudWatchLogs_CheckLogGroupDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		logGroupName       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check log group deletion protection enabled",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLogGroupsProtectionEnabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.DescribeLogGroupsOutput{
										LogGroups: []cloudwatchlogstypes.LogGroup{
											{
												LogGroupName:              aws.String("/aws/test/log-group"),
												DeletionProtectionEnabled: aws.Bool(true),
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
			name: "check log group deletion protection disabled",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLogGroupsProtectionDisabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.DescribeLogGroupsOutput{
										LogGroups: []cloudwatchlogstypes.LogGroup{
											{
												LogGroupName:              aws.String("/aws/test/log-group"),
												DeletionProtectionEnabled: aws.Bool(false),
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
			name: "check log group deletion protection with no exact match",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLogGroupsNoMatchMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.DescribeLogGroupsOutput{
										LogGroups: []cloudwatchlogstypes.LogGroup{
											{
												LogGroupName:              aws.String("/aws/test/log-group-other"),
												DeletionProtectionEnabled: aws.Bool(true),
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
			name: "check log group deletion protection with empty results",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLogGroupsEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.DescribeLogGroupsOutput{
										LogGroups: []cloudwatchlogstypes.LogGroup{},
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
			name: "check log group deletion protection with multiple results and exact match",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLogGroupsMultipleMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.DescribeLogGroupsOutput{
										LogGroups: []cloudwatchlogstypes.LogGroup{
											{
												LogGroupName:              aws.String("/aws/test/log-group"),
												DeletionProtectionEnabled: aws.Bool(true),
											},
											{
												LogGroupName:              aws.String("/aws/test/log-group-2"),
												DeletionProtectionEnabled: aws.Bool(false),
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
			name: "check log group deletion protection failure",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLogGroupsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.DescribeLogGroupsOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeLogGroupsError")
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

			sdkClient := cloudwatchlogs.NewFromConfig(cfg)
			cwlClient := NewCloudWatchLogs(sdkClient)

			got, err := cwlClient.CheckLogGroupDeletionProtection(tt.args.ctx, tt.args.logGroupName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got = %#v, want %#v", got, tt.want)
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

func TestCloudWatchLogs_DisableLogGroupDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		logGroupName       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "disable log group deletion protection successfully",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"PutLogGroupDeletionProtectionMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.PutLogGroupDeletionProtectionOutput{},
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
			name: "disable log group deletion protection failure",
			args: args{
				ctx:          context.Background(),
				logGroupName: aws.String("/aws/test/log-group"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"PutLogGroupDeletionProtectionErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &cloudwatchlogs.PutLogGroupDeletionProtectionOutput{},
								}, middleware.Metadata{}, fmt.Errorf("PutLogGroupDeletionProtectionError")
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

			sdkClient := cloudwatchlogs.NewFromConfig(cfg)
			cwlClient := NewCloudWatchLogs(sdkClient)

			err = cwlClient.DisableLogGroupDeletionProtection(tt.args.ctx, tt.args.logGroupName)
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
