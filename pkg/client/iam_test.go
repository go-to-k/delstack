package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsMiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/middleware"
)

func TestIam_DeleteGroup(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		groupName          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete group successfully",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteGroupMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteGroupOutput{},
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
			name: "delete group failure",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteGroupErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteGroupOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteGroupError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteGroup, DeleteGroupError"),
			},
			wantErr: true,
		},
		{
			name: "delete group failure for api error",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteGroupApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: &iam.DeleteGroupOutput{},
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxAttempts,
										Err:     fmt.Errorf("api error Throttling: Rate exceeded"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteGroup, exceeded maximum number of attempts, 10, api error Throttling: Rate exceeded"),
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

			err = iamClient.DeleteGroup(tt.args.ctx, tt.args.groupName)
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

func TestIam_CheckGroupExists(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		groupName          *string
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
			name: "check group exists successfully",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetGroupOutput{
										Group: &types.Group{
											GroupName: aws.String("GroupName"),
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
			name: "check group not exists successfully",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupNotExistsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetGroupOutput{},
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
			name: "check group exists failure",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetGroupOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetGroupError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: GetGroup, GetGroupError"),
				},
			},
			wantErr: true,
		},
		{
			name: "check group exists failure for api error",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: &iam.GetGroupOutput{},
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxAttempts,
										Err:     fmt.Errorf("api error Throttling: Rate exceeded"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: GetGroup, exceeded maximum number of attempts, 10, api error Throttling: Rate exceeded"),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			output, err := iamClient.CheckGroupExists(tt.args.ctx, tt.args.groupName)
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

func TestIam_GetGroupUsers(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		groupName          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		output []types.User
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "get group users successfully",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetGroupOutput{
										Users: []types.User{
											{
												UserName: aws.String("UserName1"),
											},
											{
												UserName: aws.String("UserName2"),
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
				output: []types.User{
					{
						UserName: aws.String("UserName1"),
					},
					{
						UserName: aws.String("UserName2"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "get group users are empty",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetGroupOutput{
										Users: []types.User{},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: []types.User{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "get group users failure",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetGroupOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetGroupError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: GetGroup, GetGroupError"),
				},
			},
			wantErr: true,
		},
		{
			name: "get group users failure for api error",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetGroupApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: &iam.GetGroupOutput{},
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxAttempts,
										Err:     fmt.Errorf("api error Throttling: Rate exceeded"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				output: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: GetGroup, exceeded maximum number of attempts, 10, api error Throttling: Rate exceeded"),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			output, err := iamClient.GetGroupUsers(tt.args.ctx, tt.args.groupName)
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

func TestIam_RemoveUsersFromGroup(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		groupName          *string
		users              []types.User
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "remove users from group successfully",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				users: []types.User{
					{
						UserName: aws.String("UserName1"),
					},
					{
						UserName: aws.String("UserName2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"RemoveUserFromGroupMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.RemoveUserFromGroupOutput{},
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
			name: "remove group with empty users successfully",
			args: args{
				ctx:                context.Background(),
				groupName:          aws.String("test"),
				users:              []types.User{},
				withAPIOptionsFunc: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove users from group failure",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				users: []types.User{
					{
						UserName: aws.String("UserName1"),
					},
					{
						UserName: aws.String("UserName2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"RemoveUserFromGroupErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.RemoveUserFromGroupOutput{},
								}, middleware.Metadata{}, fmt.Errorf("RemoveUserFromGroupError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: RemoveUserFromGroup, RemoveUserFromGroupError"),
			},
			wantErr: true,
		},
		{
			name: "remove users from group failure for api error",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("test"),
				users: []types.User{
					{
						UserName: aws.String("UserName1"),
					},
					{
						UserName: aws.String("UserName2"),
					},
				},
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"RemoveUserFromGroupApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: &iam.RemoveUserFromGroupOutput{},
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxAttempts,
										Err:     fmt.Errorf("api error Throttling: Rate exceeded"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: RemoveUserFromGroup, exceeded maximum number of attempts, 10, api error Throttling: Rate exceeded"),
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

			err = iamClient.RemoveUsersFromGroup(tt.args.ctx, tt.args.groupName, tt.args.users)
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

func TestIam_CheckUserExists(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
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
			name: "check user exists successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetUserMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetUserOutput{
										User: &types.User{
											UserName: aws.String("test"),
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
			name: "check user not exists successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetUserNotExistsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetUserOutput{},
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
			name: "check user exists failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetUserErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.GetUserOutput{},
								}, middleware.Metadata{}, fmt.Errorf("GetUserError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: GetUser, GetUserError"),
				},
			},
			wantErr: true,
		},
		{
			name: "check user exists failure for api error",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"GetUserApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: &iam.GetUserOutput{},
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxAttempts,
										Err:     fmt.Errorf("api error Throttling: Rate exceeded"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				exists: false,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: GetUser, exceeded maximum number of attempts, 10, api error Throttling: Rate exceeded"),
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

			client := iam.NewFromConfig(cfg)
			iamClient := NewIam(client)

			output, err := iamClient.CheckUserExists(tt.args.ctx, tt.args.userName)
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

func TestIam_DetachUserPolicies(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "detach user policies successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachUserPoliciesMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListAttachedUserPolicies" {
									return middleware.FinalizeOutput{
										Result: &iam.ListAttachedUserPoliciesOutput{
											AttachedPolicies: []types.AttachedPolicy{
												{
													PolicyArn:  aws.String("arn:aws:iam::policy/Policy1"),
													PolicyName: aws.String("Policy1"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DetachUserPolicy" {
									return middleware.FinalizeOutput{
										Result: &iam.DetachUserPolicyOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "detach user policies with empty policies successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachUserPoliciesEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedUserPoliciesOutput{
										AttachedPolicies: []types.AttachedPolicy{},
									},
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
			name: "detach user policies failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachUserPoliciesErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedUserPoliciesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListAttachedUserPoliciesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListAttachedUserPolicies, ListAttachedUserPoliciesError"),
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

			err = iamClient.DetachUserPolicies(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteUserInlinePolicies(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete user inline policies successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserInlinePoliciesMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListUserPolicies" {
									return middleware.FinalizeOutput{
										Result: &iam.ListUserPoliciesOutput{
											PolicyNames: []string{"InlinePolicy1"},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeleteUserPolicy" {
									return middleware.FinalizeOutput{
										Result: &iam.DeleteUserPolicyOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete user inline policies with empty policies successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserInlinePoliciesEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListUserPoliciesOutput{
										PolicyNames: []string{},
									},
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
			name: "delete user inline policies failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserInlinePoliciesErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListUserPoliciesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListUserPoliciesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListUserPolicies, ListUserPoliciesError"),
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

			err = iamClient.DeleteUserInlinePolicies(tt.args.ctx, tt.args.userName)
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

func TestIam_DeactivateAndDeleteMFADevices(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "deactivate and delete mfa devices successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeactivateAndDeleteMFADevicesMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListMFADevices" {
									return middleware.FinalizeOutput{
										Result: &iam.ListMFADevicesOutput{
											MFADevices: []types.MFADevice{
												{
													SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"),
													UserName:     aws.String("test"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeactivateMFADevice" {
									return middleware.FinalizeOutput{
										Result: &iam.DeactivateMFADeviceOutput{},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeleteVirtualMFADevice" {
									return middleware.FinalizeOutput{
										Result: &iam.DeleteVirtualMFADeviceOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "deactivate and delete mfa devices with empty devices successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeactivateAndDeleteMFADevicesEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListMFADevicesOutput{
										MFADevices: []types.MFADevice{},
									},
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
			name: "deactivate and delete mfa devices failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeactivateAndDeleteMFADevicesErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListMFADevicesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListMFADevicesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListMFADevices, ListMFADevicesError"),
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

			err = iamClient.DeactivateAndDeleteMFADevices(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteAccessKeys(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete access keys successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteAccessKeysMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListAccessKeys" {
									return middleware.FinalizeOutput{
										Result: &iam.ListAccessKeysOutput{
											AccessKeyMetadata: []types.AccessKeyMetadata{
												{
													AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
													UserName:    aws.String("test"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeleteAccessKey" {
									return middleware.FinalizeOutput{
										Result: &iam.DeleteAccessKeyOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete access keys with empty keys successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteAccessKeysEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAccessKeysOutput{
										AccessKeyMetadata: []types.AccessKeyMetadata{},
									},
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
			name: "delete access keys failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteAccessKeysErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAccessKeysOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListAccessKeysError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListAccessKeys, ListAccessKeysError"),
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

			err = iamClient.DeleteAccessKeys(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteLoginProfile(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete login profile successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteLoginProfileMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteLoginProfileOutput{},
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
			name: "delete login profile not exists successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteLoginProfileNotExistsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteLoginProfileOutput{},
								}, middleware.Metadata{}, fmt.Errorf("NoSuchEntity")
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
			name: "delete login profile failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteLoginProfileErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteLoginProfileOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteLoginProfileError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteLoginProfile, DeleteLoginProfileError"),
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

			err = iamClient.DeleteLoginProfile(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteSigningCertificates(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete signing certificates successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSigningCertificatesMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListSigningCertificates" {
									return middleware.FinalizeOutput{
										Result: &iam.ListSigningCertificatesOutput{
											Certificates: []types.SigningCertificate{
												{
													CertificateId: aws.String("cert-id-1"),
													UserName:      aws.String("test"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeleteSigningCertificate" {
									return middleware.FinalizeOutput{
										Result: &iam.DeleteSigningCertificateOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete signing certificates with empty certificates successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSigningCertificatesEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListSigningCertificatesOutput{
										Certificates: []types.SigningCertificate{},
									},
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
			name: "delete signing certificates failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSigningCertificatesErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListSigningCertificatesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListSigningCertificatesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListSigningCertificates, ListSigningCertificatesError"),
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

			err = iamClient.DeleteSigningCertificates(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteSSHPublicKeys(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete ssh public keys successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSSHPublicKeysMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListSSHPublicKeys" {
									return middleware.FinalizeOutput{
										Result: &iam.ListSSHPublicKeysOutput{
											SSHPublicKeys: []types.SSHPublicKeyMetadata{
												{
													SSHPublicKeyId: aws.String("ssh-key-1"),
													UserName:       aws.String("test"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeleteSSHPublicKey" {
									return middleware.FinalizeOutput{
										Result: &iam.DeleteSSHPublicKeyOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete ssh public keys with empty keys successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSSHPublicKeysEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListSSHPublicKeysOutput{
										SSHPublicKeys: []types.SSHPublicKeyMetadata{},
									},
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
			name: "delete ssh public keys failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSSHPublicKeysErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListSSHPublicKeysOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListSSHPublicKeysError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListSSHPublicKeys, ListSSHPublicKeysError"),
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

			err = iamClient.DeleteSSHPublicKeys(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteServiceSpecificCredentials(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete service specific credentials successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteServiceSpecificCredentialsMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListServiceSpecificCredentials" {
									return middleware.FinalizeOutput{
										Result: &iam.ListServiceSpecificCredentialsOutput{
											ServiceSpecificCredentials: []types.ServiceSpecificCredentialMetadata{
												{
													ServiceSpecificCredentialId: aws.String("cred-id-1"),
													UserName:                    aws.String("test"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "DeleteServiceSpecificCredential" {
									return middleware.FinalizeOutput{
										Result: &iam.DeleteServiceSpecificCredentialOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "delete service specific credentials with empty credentials successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteServiceSpecificCredentialsEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListServiceSpecificCredentialsOutput{
										ServiceSpecificCredentials: []types.ServiceSpecificCredentialMetadata{},
									},
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
			name: "delete service specific credentials failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteServiceSpecificCredentialsErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListServiceSpecificCredentialsOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListServiceSpecificCredentialsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListServiceSpecificCredentials, ListServiceSpecificCredentialsError"),
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

			err = iamClient.DeleteServiceSpecificCredentials(tt.args.ctx, tt.args.userName)
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

func TestIam_RemoveUserFromGroups(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "remove user from groups successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"RemoveUserFromGroupsMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								operationName := awsMiddleware.GetOperationName(ctx)
								if operationName == "ListGroupsForUser" {
									return middleware.FinalizeOutput{
										Result: &iam.ListGroupsForUserOutput{
											Groups: []types.Group{
												{
													GroupName: aws.String("Group1"),
												},
											},
										},
									}, middleware.Metadata{}, nil
								}
								if operationName == "RemoveUserFromGroup" {
									return middleware.FinalizeOutput{
										Result: &iam.RemoveUserFromGroupOutput{},
									}, middleware.Metadata{}, nil
								}
								return middleware.FinalizeOutput{}, middleware.Metadata{}, nil
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
			name: "remove user from groups with empty groups successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"RemoveUserFromGroupsEmptyMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListGroupsForUserOutput{
										Groups: []types.Group{},
									},
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
			name: "remove user from groups failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"RemoveUserFromGroupsErrorMock",
							func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListGroupsForUserOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListGroupsForUserError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: ListGroupsForUser, ListGroupsForUserError"),
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

			err = iamClient.RemoveUserFromGroups(tt.args.ctx, tt.args.userName)
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

func TestIam_DeleteUser(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete user successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteUserOutput{},
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
			name: "delete user failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteUserOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteUserError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteUser, DeleteUserError"),
			},
			wantErr: true,
		},
		{
			name: "delete user failure for api error",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserApiErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
										Result: &iam.DeleteUserOutput{},
									}, middleware.Metadata{}, &retry.MaxAttemptsError{
										Attempt: MaxAttempts,
										Err:     fmt.Errorf("api error Throttling: Rate exceeded"),
									}
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteUser, exceeded maximum number of attempts, 10, api error Throttling: Rate exceeded"),
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

			err = iamClient.DeleteUser(tt.args.ctx, tt.args.userName)
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
