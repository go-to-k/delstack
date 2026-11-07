package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/aws/smithy-go/middleware"
)

/*
	Test Cases
*/

func TestRoute53Resolver_GetResolverConfig(t *testing.T) {
	type args struct {
		ctx                context.Context
		resourceId         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		config *types.ResolverConfig
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
			name: "get resolver config successfully",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetResolverConfigMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &route53resolver.GetResolverConfigOutput{
										ResolverConfig: &types.ResolverConfig{
											ResourceId:         aws.String("vpc-test"),
											AutodefinedReverse: types.ResolverAutodefinedReverseStatusDisabled,
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
				config: &types.ResolverConfig{
					ResourceId:         aws.String("vpc-test"),
					AutodefinedReverse: types.ResolverAutodefinedReverseStatusDisabled,
				},
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "get resolver config returns nil when resource not found",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetResolverConfigNotFoundMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &route53resolver.GetResolverConfigOutput{},
								}, middleware.Metadata{}, &types.ResourceNotFoundException{Message: aws.String("[RSLVR-00913] Can't find the resource")}
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				config: nil,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "get resolver config failure",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetResolverConfigErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &route53resolver.GetResolverConfigOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetResolverConfigError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				config: nil,
				err: &ClientError{
					ResourceName: aws.String("vpc-test"),
					Err:          fmt.Errorf("operation error Route53Resolver: GetResolverConfig, GetResolverConfigError"),
				},
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

			client := route53resolver.NewFromConfig(cfg)
			route53ResolverClient := NewRoute53Resolver(client)

			output, exists, err := route53ResolverClient.GetResolverConfig(tt.args.ctx, tt.args.resourceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.config) {
				t.Errorf("output = %#v, want %#v", output, tt.want.config)
			}
			if exists != tt.want.exists {
				t.Errorf("exists = %#v, want %#v", exists, tt.want.exists)
			}
		})
	}
}

func TestRoute53Resolver_UpdateResolverConfig(t *testing.T) {
	type args struct {
		ctx                context.Context
		resourceId         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "update resolver config successfully",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateResolverConfigMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &route53resolver.UpdateResolverConfigOutput{},
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
			name: "update resolver config failure",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"UpdateResolverConfigErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &route53resolver.UpdateResolverConfigOutput{},
								}, middleware.Metadata{}, fmt.Errorf("UpdateResolverConfigError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("vpc-test"),
				Err:          fmt.Errorf("operation error Route53Resolver: UpdateResolverConfig, UpdateResolverConfigError"),
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

			client := route53resolver.NewFromConfig(cfg)
			route53ResolverClient := NewRoute53Resolver(client)

			err = route53ResolverClient.UpdateResolverConfig(tt.args.ctx, tt.args.resourceId, types.AutodefinedReverseFlagEnable)
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
