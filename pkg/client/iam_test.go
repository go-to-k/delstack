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
										Attempt: MaxRetryCount,
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
										Attempt: MaxRetryCount,
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
										Attempt: MaxRetryCount,
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
										Attempt: MaxRetryCount,
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
