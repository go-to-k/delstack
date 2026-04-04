package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestIam_ListAttachedUserPolicies(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		policies []types.AttachedPolicy
		err      error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list attached user policies successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedUserPoliciesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedUserPoliciesOutput{
										AttachedPolicies: []types.AttachedPolicy{
											{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
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
				policies: []types.AttachedPolicy{
					{PolicyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess")},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list attached user policies failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAttachedUserPoliciesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAttachedUserPoliciesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListAttachedUserPoliciesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				policies: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListAttachedUserPolicies, ListAttachedUserPoliciesError"),
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

			output, _, err := iamClient.ListAttachedUserPolicies(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.policies) {
				t.Errorf("output = %#v, want %#v", output, tt.want.policies)
			}
		})
	}
}

func TestIam_DetachUserPolicy(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		policyArn          *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "detach user policy successfully",
			args: args{
				ctx:       context.Background(),
				userName:  aws.String("test"),
				policyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachUserPolicyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachUserPolicyOutput{},
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
			name: "detach user policy failure",
			args: args{
				ctx:       context.Background(),
				userName:  aws.String("test"),
				policyArn: aws.String("arn:aws:iam::aws:policy/ReadOnlyAccess"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DetachUserPolicyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DetachUserPolicyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DetachUserPolicyError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DetachUserPolicy, DetachUserPolicyError"),
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

			err = iamClient.DetachUserPolicy(tt.args.ctx, tt.args.userName, tt.args.policyArn)
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

func TestIam_ListUserPolicies(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		policies []string
		err      error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list user policies successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListUserPoliciesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListUserPoliciesOutput{
										PolicyNames: []string{"InlinePolicy1"},
									},
								}, middleware.Metadata{}, nil
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				policies: []string{"InlinePolicy1"},
				err:      nil,
			},
			wantErr: false,
		},
		{
			name: "list user policies failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListUserPoliciesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListUserPoliciesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListUserPoliciesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				policies: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListUserPolicies, ListUserPoliciesError"),
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

			output, _, err := iamClient.ListUserPolicies(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.policies) {
				t.Errorf("output = %#v, want %#v", output, tt.want.policies)
			}
		})
	}
}

func TestIam_DeleteUserPolicy(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		policyName         *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete user policy successfully",
			args: args{
				ctx:        context.Background(),
				userName:   aws.String("test"),
				policyName: aws.String("InlinePolicy1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserPolicyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteUserPolicyOutput{},
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
			name: "delete user policy failure",
			args: args{
				ctx:        context.Background(),
				userName:   aws.String("test"),
				policyName: aws.String("InlinePolicy1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteUserPolicyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteUserPolicyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteUserPolicyError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteUserPolicy, DeleteUserPolicyError"),
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

			err = iamClient.DeleteUserPolicy(tt.args.ctx, tt.args.userName, tt.args.policyName)
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

func TestIam_ListMFADevices(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		devices []types.MFADevice
		err     error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list mfa devices successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListMFADevicesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
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
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				devices: []types.MFADevice{
					{
						SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"),
						UserName:     aws.String("test"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list mfa devices failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListMFADevicesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListMFADevicesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListMFADevicesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				devices: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListMFADevices, ListMFADevicesError"),
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

			output, _, err := iamClient.ListMFADevices(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.devices) {
				t.Errorf("output = %#v, want %#v", output, tt.want.devices)
			}
		})
	}
}

func TestIam_DeactivateMFADevice(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		serialNumber       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "deactivate mfa device successfully",
			args: args{
				ctx:          context.Background(),
				userName:     aws.String("test"),
				serialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeactivateMFADeviceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeactivateMFADeviceOutput{},
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
			name: "deactivate mfa device failure",
			args: args{
				ctx:          context.Background(),
				userName:     aws.String("test"),
				serialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeactivateMFADeviceErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeactivateMFADeviceOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeactivateMFADeviceError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeactivateMFADevice, DeactivateMFADeviceError"),
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

			err = iamClient.DeactivateMFADevice(tt.args.ctx, tt.args.userName, tt.args.serialNumber)
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

func TestIam_DeleteVirtualMFADevice(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		serialNumber       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete virtual mfa device successfully",
			args: args{
				ctx:          context.Background(),
				serialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteVirtualMFADeviceMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteVirtualMFADeviceOutput{},
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
			name: "delete virtual mfa device failure",
			args: args{
				ctx:          context.Background(),
				serialNumber: aws.String("arn:aws:iam::123456789012:mfa/test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteVirtualMFADeviceErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteVirtualMFADeviceOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteVirtualMFADeviceError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("arn:aws:iam::123456789012:mfa/test"),
				Err:          fmt.Errorf("operation error IAM: DeleteVirtualMFADevice, DeleteVirtualMFADeviceError"),
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

			err = iamClient.DeleteVirtualMFADevice(tt.args.ctx, tt.args.serialNumber)
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

func TestIam_ListAccessKeys(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		keys []types.AccessKeyMetadata
		err  error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list access keys successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAccessKeysMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
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
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				keys: []types.AccessKeyMetadata{
					{
						AccessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
						UserName:    aws.String("test"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list access keys failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListAccessKeysErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListAccessKeysOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListAccessKeysError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				keys: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListAccessKeys, ListAccessKeysError"),
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

			output, _, err := iamClient.ListAccessKeys(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.keys) {
				t.Errorf("output = %#v, want %#v", output, tt.want.keys)
			}
		})
	}
}

func TestIam_DeleteAccessKey(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		accessKeyId        *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete access key successfully",
			args: args{
				ctx:         context.Background(),
				userName:    aws.String("test"),
				accessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteAccessKeyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteAccessKeyOutput{},
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
			name: "delete access key failure",
			args: args{
				ctx:         context.Background(),
				userName:    aws.String("test"),
				accessKeyId: aws.String("AKIAIOSFODNN7EXAMPLE"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteAccessKeyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteAccessKeyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteAccessKeyError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteAccessKey, DeleteAccessKeyError"),
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

			err = iamClient.DeleteAccessKey(tt.args.ctx, tt.args.userName, tt.args.accessKeyId)
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

func TestIam_ListSigningCertificates(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		certificates []types.SigningCertificate
		err          error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list signing certificates successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListSigningCertificatesMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
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
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				certificates: []types.SigningCertificate{
					{
						CertificateId: aws.String("cert-id-1"),
						UserName:      aws.String("test"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list signing certificates failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListSigningCertificatesErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListSigningCertificatesOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListSigningCertificatesError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				certificates: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListSigningCertificates, ListSigningCertificatesError"),
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

			output, _, err := iamClient.ListSigningCertificates(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.certificates) {
				t.Errorf("output = %#v, want %#v", output, tt.want.certificates)
			}
		})
	}
}

func TestIam_DeleteSigningCertificate(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		certificateId      *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete signing certificate successfully",
			args: args{
				ctx:           context.Background(),
				userName:      aws.String("test"),
				certificateId: aws.String("cert-id-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSigningCertificateMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteSigningCertificateOutput{},
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
			name: "delete signing certificate failure",
			args: args{
				ctx:           context.Background(),
				userName:      aws.String("test"),
				certificateId: aws.String("cert-id-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSigningCertificateErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteSigningCertificateOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteSigningCertificateError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteSigningCertificate, DeleteSigningCertificateError"),
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

			err = iamClient.DeleteSigningCertificate(tt.args.ctx, tt.args.userName, tt.args.certificateId)
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

func TestIam_ListSSHPublicKeys(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		keys []types.SSHPublicKeyMetadata
		err  error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list ssh public keys successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListSSHPublicKeysMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
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
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				keys: []types.SSHPublicKeyMetadata{
					{
						SSHPublicKeyId: aws.String("ssh-key-1"),
						UserName:       aws.String("test"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list ssh public keys failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListSSHPublicKeysErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListSSHPublicKeysOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListSSHPublicKeysError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				keys: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListSSHPublicKeys, ListSSHPublicKeysError"),
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

			output, _, err := iamClient.ListSSHPublicKeys(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.keys) {
				t.Errorf("output = %#v, want %#v", output, tt.want.keys)
			}
		})
	}
}

func TestIam_DeleteSSHPublicKey(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		sshPublicKeyId     *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete ssh public key successfully",
			args: args{
				ctx:            context.Background(),
				userName:       aws.String("test"),
				sshPublicKeyId: aws.String("ssh-key-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSSHPublicKeyMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteSSHPublicKeyOutput{},
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
			name: "delete ssh public key failure",
			args: args{
				ctx:            context.Background(),
				userName:       aws.String("test"),
				sshPublicKeyId: aws.String("ssh-key-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteSSHPublicKeyErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteSSHPublicKeyOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteSSHPublicKeyError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteSSHPublicKey, DeleteSSHPublicKeyError"),
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

			err = iamClient.DeleteSSHPublicKey(tt.args.ctx, tt.args.userName, tt.args.sshPublicKeyId)
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

func TestIam_ListServiceSpecificCredentials(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		credentials []types.ServiceSpecificCredentialMetadata
		err         error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list service specific credentials successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListServiceSpecificCredentialsMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
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
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				credentials: []types.ServiceSpecificCredentialMetadata{
					{
						ServiceSpecificCredentialId: aws.String("cred-id-1"),
						UserName:                    aws.String("test"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list service specific credentials failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListServiceSpecificCredentialsErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListServiceSpecificCredentialsOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListServiceSpecificCredentialsError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				credentials: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListServiceSpecificCredentials, ListServiceSpecificCredentialsError"),
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

			output, err := iamClient.ListServiceSpecificCredentials(tt.args.ctx, tt.args.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.credentials) {
				t.Errorf("output = %#v, want %#v", output, tt.want.credentials)
			}
		})
	}
}

func TestIam_DeleteServiceSpecificCredential(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		credentialId       *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete service specific credential successfully",
			args: args{
				ctx:          context.Background(),
				userName:     aws.String("test"),
				credentialId: aws.String("cred-id-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteServiceSpecificCredentialMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteServiceSpecificCredentialOutput{},
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
			name: "delete service specific credential failure",
			args: args{
				ctx:          context.Background(),
				userName:     aws.String("test"),
				credentialId: aws.String("cred-id-1"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"DeleteServiceSpecificCredentialErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.DeleteServiceSpecificCredentialOutput{},
								}, middleware.Metadata{}, fmt.Errorf("DeleteServiceSpecificCredentialError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: &ClientError{
				ResourceName: aws.String("test"),
				Err:          fmt.Errorf("operation error IAM: DeleteServiceSpecificCredential, DeleteServiceSpecificCredentialError"),
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

			err = iamClient.DeleteServiceSpecificCredential(tt.args.ctx, tt.args.userName, tt.args.credentialId)
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

func TestIam_ListGroupsForUser(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		userName           *string
		withAPIOptionsFunc func(*middleware.Stack) error
	}

	type want struct {
		groups []types.Group
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list groups for user successfully",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListGroupsForUserMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListGroupsForUserOutput{
										Groups: []types.Group{
											{
												GroupName: aws.String("Group1"),
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
				groups: []types.Group{
					{
						GroupName: aws.String("Group1"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list groups for user failure",
			args: args{
				ctx:      context.Background(),
				userName: aws.String("test"),
				withAPIOptionsFunc: func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"ListGroupsForUserErrorMock",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: &iam.ListGroupsForUserOutput{},
								}, middleware.Metadata{}, fmt.Errorf("ListGroupsForUserError")
							},
						),
						middleware.Before,
					)
				},
			},
			want: want{
				groups: nil,
				err: &ClientError{
					ResourceName: aws.String("test"),
					Err:          fmt.Errorf("operation error IAM: ListGroupsForUser, ListGroupsForUserError"),
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

			output, _, err := iamClient.ListGroupsForUser(tt.args.ctx, tt.args.userName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.groups) {
				t.Errorf("output = %#v, want %#v", output, tt.want.groups)
			}
		})
	}
}

func TestIam_RemoveUserFromGroup(t *testing.T) {
	SleepTimeSecForIam = 1
	type args struct {
		ctx                context.Context
		groupName          *string
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
			name: "remove user from group successfully",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("Group1"),
				userName:  aws.String("test"),
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
			name: "remove user from group failure",
			args: args{
				ctx:       context.Background(),
				groupName: aws.String("Group1"),
				userName:  aws.String("test"),
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
			want:    fmt.Errorf("operation error IAM: RemoveUserFromGroup, RemoveUserFromGroupError"),
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

			err = iamClient.RemoveUserFromGroup(tt.args.ctx, tt.args.groupName, tt.args.userName)
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
