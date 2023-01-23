package client

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
)

const sleepTimeSecForIam = 1

type markerKeyForIam struct{}

func getNextMarkerForIamInitialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	switch v := in.Parameters.(type) {
	case *iam.ListAttachedRolePoliciesInput:
		ctx = middleware.WithStackValue(ctx, markerKeyForIam{}, v.Marker)
	}
	return next.HandleInitialize(ctx, in)
}

/*
	Test Cases
*/

func TestIam_DeleteRole(t *testing.T) {
	type args struct {
		ctx                context.Context
		roleName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete role successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRoleMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteRoleOutput{},
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
			name: "delete role failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRoleErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteRoleOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteRoleError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error IAM: DeleteRole, DeleteRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for api error",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteRoleApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteRoleOutput{},
								}, middleware.Metadata{}, fmt.Errorf("api error Throttling: Rate exceeded")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("RetryCountOverError: test, operation error IAM: DeleteRole, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			err = iamClient.DeleteRole(tt.args.ctx, tt.args.roleName, sleepTimeSecForIam)
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

func TestIam_ListAttachedRolePolicies(t *testing.T) {
	type args struct {
		ctx                context.Context
		roleName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output []types.AttachedPolicy
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list attached role policies successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedRolePoliciesOutput{
										AttachedPolicies: []types.AttachedPolicy{
											{
												PolicyArn:  aws.String("PolicyArn1"),
												PolicyName: aws.String("PolicyName1"),
											},
											{
												PolicyArn:  aws.String("PolicyArn2"),
												PolicyName: aws.String("PolicyName2"),
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
				output: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list attached role policies are empty",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedRolePoliciesOutput{
										AttachedPolicies: []types.AttachedPolicy{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.AttachedPolicy{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "list attached role policies failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedRolePoliciesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListAttachedRolePoliciesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("operation error IAM: ListAttachedRolePolicies, ListAttachedRolePoliciesError"),
			},
			wantErr: true,
		},
		{
			name: "list attached role policies failure for api error",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedRolePoliciesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("api error Throttling: Rate exceeded")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("RetryCountOverError: test, operation error IAM: ListAttachedRolePolicies, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			},
			wantErr: true,
		},
		{
			name: "list attached role policies with next marker successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextMarker",
							getNextMarkerForIamInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesWithNextMarkerMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								marker := middleware.GetStackValue(ctx, markerKeyForIam{}).(*string)

								var nextMarker *string
								var attachedPolicies []types.AttachedPolicy
								if marker == nil {
									nextMarker = aws.String("NextMarker")
									attachedPolicies = []types.AttachedPolicy{
										{
											PolicyArn:  aws.String("PolicyArn1"),
											PolicyName: aws.String("PolicyName1"),
										},
										{
											PolicyArn:  aws.String("PolicyArn2"),
											PolicyName: aws.String("PolicyName2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedRolePoliciesOutput{
											Marker:           nextMarker,
											AttachedPolicies: attachedPolicies,
										},
									}, middleware.Metadata{}, nil
								} else {
									attachedPolicies = []types.AttachedPolicy{
										{
											PolicyArn:  aws.String("PolicyArn3"),
											PolicyName: aws.String("PolicyName3"),
										},
										{
											PolicyArn:  aws.String("PolicyArn4"),
											PolicyName: aws.String("PolicyName4"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedRolePoliciesOutput{
											Marker:           nextMarker,
											AttachedPolicies: attachedPolicies,
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
				output: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
					{
						PolicyArn:  aws.String("PolicyArn3"),
						PolicyName: aws.String("PolicyName3"),
					},
					{
						PolicyArn:  aws.String("PolicyArn4"),
						PolicyName: aws.String("PolicyName4"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list attached role policies with next marker failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextMarker",
							getNextMarkerForIamInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesWithNextMarkerErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								marker := middleware.GetStackValue(ctx, markerKeyForIam{}).(*string)

								var nextMarker *string
								var attachedPolicies []types.AttachedPolicy
								if marker == nil {
									nextMarker = aws.String("NextMarker")
									attachedPolicies = []types.AttachedPolicy{
										{
											PolicyArn:  aws.String("PolicyArn1"),
											PolicyName: aws.String("PolicyName1"),
										},
										{
											PolicyArn:  aws.String("PolicyArn2"),
											PolicyName: aws.String("PolicyName2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedRolePoliciesOutput{
											Marker:           nextMarker,
											AttachedPolicies: attachedPolicies,
										},
									}, middleware.Metadata{}, nil
								} else {
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedRolePoliciesOutput{},
									}, middleware.Metadata{}, fmt.Errorf("ListAttachedRolePoliciesError")
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
				err:    fmt.Errorf("operation error IAM: ListAttachedRolePolicies, ListAttachedRolePoliciesError"),
			},
			wantErr: true,
		},
		{
			name: "list attached role policies with next marker failure for api error",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					err := stack.Initialize.Add(
						middleware.InitializeMiddlewareFunc(
							"GetNextMarker",
							getNextMarkerForIamInitialize,
						), middleware.Before,
					)
					if err != nil {
						return err
					}

					err = stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedRolePoliciesWithNextMarkerApiErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								marker := middleware.GetStackValue(ctx, markerKeyForIam{}).(*string)

								var nextMarker *string
								var attachedPolicies []types.AttachedPolicy
								if marker == nil {
									nextMarker = aws.String("NextMarker")
									attachedPolicies = []types.AttachedPolicy{
										{
											PolicyArn:  aws.String("PolicyArn1"),
											PolicyName: aws.String("PolicyName1"),
										},
										{
											PolicyArn:  aws.String("PolicyArn2"),
											PolicyName: aws.String("PolicyName2"),
										},
									}
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedRolePoliciesOutput{
											Marker:           nextMarker,
											AttachedPolicies: attachedPolicies,
										},
									}, middleware.Metadata{}, nil
								} else {
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedRolePoliciesOutput{},
									}, middleware.Metadata{}, fmt.Errorf("api error Throttling: Rate exceeded")
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
				err:    fmt.Errorf("RetryCountOverError: test, operation error IAM: ListAttachedRolePolicies, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			output, err := iamClient.ListAttachedRolePolicies(tt.args.ctx, tt.args.roleName, sleepTimeSecForIam)
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

func TestIam_DetachRolePolicies(t *testing.T) {
	type args struct {
		ctx                context.Context
		roleName           *string
		policies           []types.AttachedPolicy
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "detach role policies successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
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
			name: "detach empty role policies successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyEmptyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
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
			name: "detach role policies failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DetachRolePolicyError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error IAM: DetachRolePolicy, DetachRolePolicyError"),
			wantErr: true,
		},
		{
			name: "detach role policies failure for api error",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				policies: []types.AttachedPolicy{
					{
						PolicyArn:  aws.String("PolicyArn1"),
						PolicyName: aws.String("PolicyName1"),
					},
					{
						PolicyArn:  aws.String("PolicyArn2"),
						PolicyName: aws.String("PolicyName2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("api error Throttling: Rate exceeded")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("RetryCountOverError: test, operation error IAM: DetachRolePolicy, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			err = iamClient.DetachRolePolicies(tt.args.ctx, tt.args.roleName, tt.args.policies, sleepTimeSecForIam)
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

func TestIam_DetachRolePolicy(t *testing.T) {
	type args struct {
		ctx                context.Context
		roleName           *string
		PolicyArn          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "detach role policy successfully",
			args: args{
				ctx:       context.Background(),
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
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
			name: "detach role policy failure",
			args: args{
				ctx:       context.Background(),
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DetachRolePolicyError")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("operation error IAM: DetachRolePolicy, DetachRolePolicyError"),
			wantErr: true,
		},
		{
			name: "detach role policy failure for api error",
			args: args{
				ctx:       context.Background(),
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachRolePolicyApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachRolePolicyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("api error Throttling: Rate exceeded")
							},
						),
						middleware.Before,
					)
				},
			},
			want:    fmt.Errorf("RetryCountOverError: test, operation error IAM: DetachRolePolicy, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			err = iamClient.DetachRolePolicy(tt.args.ctx, tt.args.roleName, tt.args.PolicyArn, sleepTimeSecForIam)
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

func TestIam_CheckRoleExists(t *testing.T) {
	type args struct {
		ctx                context.Context
		roleName           *string
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
			name: "check role exists successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetRoleMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetRoleOutput{
										Role: &types.Role{
											RoleName: aws.String("RoleName"),
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
			name: "check role not exists successfully",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetRoleNotExistsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetRoleOutput{},
								}, middleware.Metadata{}, fmt.Errorf("NoSuchEntity")
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
			name: "check role exists failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetRoleErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetRoleOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetRoleError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("operation error IAM: GetRole, GetRoleError"),
			},
			wantErr: true,
		},
		{
			name: "check role exists failure for api error",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetRoleApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetRoleOutput{},
								}, middleware.Metadata{}, fmt.Errorf("api error Throttling: Rate exceeded")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("RetryCountOverError: test, operation error IAM: GetRole, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			output, err := iamClient.CheckRoleExists(tt.args.ctx, tt.args.roleName, sleepTimeSecForIam)
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
