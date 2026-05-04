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

func TestEC2Client_DeleteSubnet(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		subnetId           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete subnet successfully",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-111"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSubnetMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteSubnetOutput{},
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
			name: "delete subnet already deleted (NotFound)",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-222"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSubnetNotFoundMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteSubnetOutput{},
								}, middleware.Metadata{}, fmt.Errorf("InvalidSubnetID.NotFound: The subnet ID 'subnet-222' does not exist")
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: false,
		},
		{
			name: "delete subnet with dependency error",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-333"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSubnetDependencyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteSubnetOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DependencyViolation: The subnet 'subnet-333' has dependencies and cannot be deleted")
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

			sdkClient := ec2.NewFromConfig(cfg)
			ec2Client := NewEC2Client(sdkClient)

			err = ec2Client.DeleteSubnet(tt.args.ctx, tt.args.subnetId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.subnetId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.subnetId)
				}
			}
		})
	}
}

func TestEC2Client_DeleteSecurityGroup(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		securityGroupId    *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "delete security group successfully",
			args: args{
				ctx:             context.Background(),
				securityGroupId: aws.String("sg-111"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSecurityGroupMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteSecurityGroupOutput{},
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
			name: "delete security group already deleted (NotFound)",
			args: args{
				ctx:             context.Background(),
				securityGroupId: aws.String("sg-222"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSecurityGroupNotFoundMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteSecurityGroupOutput{},
								}, middleware.Metadata{}, fmt.Errorf("InvalidGroup.NotFound: The security group 'sg-222' does not exist")
							},
						),
						middleware.Before,
					)
				},
			},
			wantErr: false,
		},
		{
			name: "delete security group with dependency error",
			args: args{
				ctx:             context.Background(),
				securityGroupId: aws.String("sg-333"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSecurityGroupDependencyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DeleteSecurityGroupOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DependencyViolation: resource sg-333 has a dependent object")
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

			sdkClient := ec2.NewFromConfig(cfg)
			ec2Client := NewEC2Client(sdkClient)

			err = ec2Client.DeleteSecurityGroup(tt.args.ctx, tt.args.securityGroupId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.securityGroupId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.securityGroupId)
				}
			}
		})
	}
}

func TestEC2Client_CheckTerminationProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		instanceId         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check termination protection enabled successfully",
			args: args{
				ctx:        context.Background(),
				instanceId: aws.String("i-111"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeInstanceAttributeEnabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeInstanceAttributeOutput{
										DisableApiTermination: &types.AttributeBooleanValue{
											Value: aws.Bool(true),
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
			name: "check termination protection disabled successfully",
			args: args{
				ctx:        context.Background(),
				instanceId: aws.String("i-222"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeInstanceAttributeDisabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeInstanceAttributeOutput{
										DisableApiTermination: &types.AttributeBooleanValue{
											Value: aws.Bool(false),
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
			name: "check termination protection failure",
			args: args{
				ctx:        context.Background(),
				instanceId: aws.String("i-333"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeInstanceAttributeErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeInstanceAttributeOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeInstanceAttributeError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "check termination protection with nil value",
			args: args{
				ctx:        context.Background(),
				instanceId: aws.String("i-444"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeInstanceAttributeNilValueMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.DescribeInstanceAttributeOutput{
										DisableApiTermination: &types.AttributeBooleanValue{
											Value: nil,
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

			got, err := ec2Client.CheckTerminationProtection(tt.args.ctx, tt.args.instanceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.instanceId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.instanceId)
				}
			}
		})
	}
}

func TestEC2Client_DisableTerminationProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		instanceId         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "disable termination protection successfully",
			args: args{
				ctx:        context.Background(),
				instanceId: aws.String("i-111"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyInstanceAttributeMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.ModifyInstanceAttributeOutput{},
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
			name: "disable termination protection failure",
			args: args{
				ctx:        context.Background(),
				instanceId: aws.String("i-222"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyInstanceAttributeErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ec2.ModifyInstanceAttributeOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ModifyInstanceAttributeError")
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

			sdkClient := ec2.NewFromConfig(cfg)
			ec2Client := NewEC2Client(sdkClient)

			err = ec2Client.DisableTerminationProtection(tt.args.ctx, tt.args.instanceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				var clientErr *ClientError
				if !errors.As(err, &clientErr) {
					t.Errorf("expected ClientError, got = %#v", err)
				}
				if clientErr == nil || *clientErr.ResourceName != *tt.args.instanceId {
					t.Errorf("ClientError ResourceName = %#v, want %#v", clientErr, tt.args.instanceId)
				}
			}
		})
	}
}
