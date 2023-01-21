package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/smithy-go/middleware"
)

type tokenKeyForEcr struct{}

func getNextTokenForEcrInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	switch v := in.Parameters.(type) {
	case *ecr.DescribeRepositoriesInput:
		ctx = middleware.WithStackValue(ctx, tokenKeyForIam{}, v.NextToken)
	}
	return next.HandleInitialize(ctx, in)
}

/*
	Test Cases
*/

func TestEcr_DeleteRepository(t *testing.T) {
	type args struct {
		ctx                context.Context
		repositoryName     *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete repository successfully",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRepositoryMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ecr.DeleteRepositoryOutput{},
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
			name: "delete repository failure",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRepositoryErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ecr.DeleteRepositoryOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteRepositoryError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error ECR: DeleteRepository, DeleteRepositoryError"),
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

			client := ecr.NewFromConfig(cfg)
			ecrClient := NewEcr(client)

			err = ecrClient.DeleteRepository(tt.args.ctx, tt.args.repositoryName)
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

func TestEcr_CheckRepository(t *testing.T) {
	type args struct {
		ctx                context.Context
		repositoryName     *string
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
			name: "check repository exists successfully",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeRepositoriesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ecr.DescribeRepositoriesOutput{
										Repositories: []types.Repository{
											{
												RepositoryName: aws.String("test"),
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
			name: "check repository not exists successfully",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeRepositoriesNotExistMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ecr.DescribeRepositoriesOutput{
										Repositories: []types.Repository{},
									},
								}, middleware.Metadata{}, fmt.Errorf("does not exist")
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
			name: "check repository exists failure",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DescribeRepositoriesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &ecr.DescribeRepositoriesOutput{
										Repositories: []types.Repository{},
									},
								}, middleware.Metadata{}, fmt.Errorf("DescribeRepositoriesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("operation error ECR: DescribeRepositories, DescribeRepositoriesError"),
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

			client := ecr.NewFromConfig(cfg)
			ecrClient := NewEcr(client)

			output, err := ecrClient.CheckEcrExists(tt.args.ctx, tt.args.repositoryName)
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
