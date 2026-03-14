package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go/middleware"
	"go.uber.org/goleak"
)

func TestELBV2_CheckLoadBalancerDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		loadBalancerArn    *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "check load balancer deletion protection enabled",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLoadBalancerAttributesProtectionEnabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{
										Attributes: []types.LoadBalancerAttribute{
											{
												Key:   aws.String("deletion_protection.enabled"),
												Value: aws.String("true"),
											},
											{
												Key:   aws.String("idle_timeout.timeout_seconds"),
												Value: aws.String("60"),
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
			name: "check load balancer deletion protection disabled",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLoadBalancerAttributesProtectionDisabledMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{
										Attributes: []types.LoadBalancerAttribute{
											{
												Key:   aws.String("deletion_protection.enabled"),
												Value: aws.String("false"),
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
			name: "check load balancer deletion protection with no matching attribute",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLoadBalancerAttributesNoMatchMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{
										Attributes: []types.LoadBalancerAttribute{
											{
												Key:   aws.String("idle_timeout.timeout_seconds"),
												Value: aws.String("60"),
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
			name: "check load balancer deletion protection with empty attributes",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLoadBalancerAttributesEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{
										Attributes: []types.LoadBalancerAttribute{},
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
			name: "check load balancer deletion protection failure",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeLoadBalancerAttributesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DescribeLoadBalancerAttributesError")
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

			sdkClient := elasticloadbalancingv2.NewFromConfig(cfg)
			elbv2Client := NewELBV2(sdkClient)

			got, err := elbv2Client.CheckLoadBalancerDeletionProtection(tt.args.ctx, tt.args.loadBalancerArn)
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

func TestELBV2_DisableLoadBalancerDeletionProtection(t *testing.T) {
	defer goleak.VerifyNone(t)

	type args struct {
		ctx                context.Context
		loadBalancerArn    *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "disable load balancer deletion protection successfully",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyLoadBalancerAttributesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.ModifyLoadBalancerAttributesOutput{},
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
			name: "disable load balancer deletion protection failure",
			args: args{
				ctx:             context.Background(),
				loadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-lb/1234567890123456"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ModifyLoadBalancerAttributesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &elasticloadbalancingv2.ModifyLoadBalancerAttributesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ModifyLoadBalancerAttributesError")
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

			sdkClient := elasticloadbalancingv2.NewFromConfig(cfg)
			elbv2Client := NewELBV2(sdkClient)

			err = elbv2Client.DisableLoadBalancerDeletionProtection(tt.args.ctx, tt.args.loadBalancerArn)
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
