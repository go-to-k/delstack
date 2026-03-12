package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/goleak"
)

func TestEC2Client_DescribeNetworkInterfaces(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		filters            []types.Filter
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    []types.NetworkInterface
		wantErr bool
	}{
		{
			name: "describe network interfaces successfully",
			args: args{
				ctx: context.Background(),
				filters: []types.Filter{
					{
						Name:   aws.String("description"),
						Values: []string{"test-description"},
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeNetworkInterfacesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeNetworkInterfacesOutput{
										NetworkInterfaces: []types.NetworkInterface{
											{NetworkInterfaceId: aws.String("eni-111")},
											{NetworkInterfaceId: aws.String("eni-222")},
										},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: []types.NetworkInterface{
				{NetworkInterfaceId: aws.String("eni-111")},
				{NetworkInterfaceId: aws.String("eni-222")},
			},
			wantErr: false,
		},
		{
			name: "describe network interfaces with no results",
			args: args{
				ctx:     context.Background(),
				filters: []types.Filter{},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeNetworkInterfacesEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeNetworkInterfacesOutput{
										NetworkInterfaces: []types.NetworkInterface{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want:    []types.NetworkInterface{},
			wantErr: false,
		},
		{
			name: "describe network interfaces failure",
			args: args{
				ctx:     context.Background(),
				filters: []types.Filter{},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeNetworkInterfacesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeNetworkInterfacesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeNetworkInterfacesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    nil,
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

			sdkClient := ec2.NewFromConfig(cfg)
			ec2Client := NewEC2Client(sdkClient)

			got, err := ec2Client.DescribeNetworkInterfaces(tt.args.ctx, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("got %d interfaces, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if *got[i].NetworkInterfaceId != *tt.want[i].NetworkInterfaceId {
						t.Errorf("got[%d].NetworkInterfaceId = %v, want %v", i, *got[i].NetworkInterfaceId, *tt.want[i].NetworkInterfaceId)
					}
				}
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

func TestEC2Client_DeleteNetworkInterface(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		networkInterfaceId *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete network interface successfully",
			args: args{
				ctx:                context.Background(),
				networkInterfaceId: aws.String("eni-111"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteNetworkInterfaceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteNetworkInterfaceOutput{},
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
			name: "delete network interface already deleted (NotFound)",
			args: args{
				ctx:                context.Background(),
				networkInterfaceId: aws.String("eni-222"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteNetworkInterfaceNotFoundMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteNetworkInterfaceOutput{},
								}, middleware.Metadata{}, fmt.Errorf("InvalidNetworkInterfaceID.NotFound: The networkInterface ID 'eni-222' does not exist")
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: false,
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

			sdkClient := ec2.NewFromConfig(cfg)
			ec2Client := NewEC2Client(sdkClient)

			err = ec2Client.DeleteNetworkInterface(tt.args.ctx, tt.args.networkInterfaceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.networkInterfaceId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.networkInterfaceId)
				}
			}
		})
	}
}
