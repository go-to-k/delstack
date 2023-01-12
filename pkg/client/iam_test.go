package client

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

const sleepTimeSecForIam = 1

/*
	Test Cases
*/

func TestIam_DeleteRole(t *testing.T) {
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		client   IIamSDKClient
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
				client:   mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				client:   errorMock,
			},
			want:    fmt.Errorf("DeleteRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for api error",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				client:   apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			err := iamClient.DeleteRole(tt.args.ctx, tt.args.roleName, sleepTimeSecForIam)
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

func TestIam_deleteRoleWithRetry(t *testing.T) {
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		input    *iam.DeleteRoleInput
		roleName *string
		client   IIamSDKClient
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
				ctx: context.Background(),
				input: &iam.DeleteRoleInput{
					RoleName: aws.String("test"),
				},
				roleName: aws.String("test"),
				client:   mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete role failure",
			args: args{
				ctx: context.Background(),
				input: &iam.DeleteRoleInput{
					RoleName: aws.String("test"),
				},
				roleName: aws.String("test"),
				client:   errorMock,
			},
			want:    fmt.Errorf("DeleteRoleError"),
			wantErr: true,
		},
		{
			name: "delete role failure for api error",
			args: args{
				ctx: context.Background(),
				input: &iam.DeleteRoleInput{
					RoleName: aws.String("test"),
				},
				roleName: aws.String("test"),
				client:   apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			_, err := iamClient.deleteRoleWithRetry(tt.args.ctx, tt.args.input, tt.args.roleName, sleepTimeSecForIam)
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
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		client   IIamSDKClient
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
				client:   mock,
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
			name: "list attached role policies failure",
			args: args{
				ctx:      context.Background(),
				roleName: aws.String("test"),
				client:   errorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("ListAttachedRolePoliciesError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			output, err := iamClient.ListAttachedRolePolicies(tt.args.ctx, tt.args.roleName)
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
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		policies []types.AttachedPolicy
		client   IIamSDKClient
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
				client: mock,
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
				client:   mock,
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
				client: errorMock,
			},
			want:    fmt.Errorf("DetachRolePolicyError"),
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
				client: apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			err := iamClient.DetachRolePolicies(tt.args.ctx, tt.args.roleName, tt.args.policies, sleepTimeSecForIam)
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
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx       context.Context
		roleName  *string
		PolicyArn *string
		client    IIamSDKClient
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
				client:    mock,
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
				client:    errorMock,
			},
			want:    fmt.Errorf("DetachRolePolicyError"),
			wantErr: true,
		},
		{
			name: "detach role policy failure for api error",
			args: args{
				ctx:       context.Background(),
				roleName:  aws.String("test"),
				PolicyArn: aws.String("test"),
				client:    apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			err := iamClient.DetachRolePolicy(tt.args.ctx, tt.args.roleName, tt.args.PolicyArn, sleepTimeSecForIam)
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

func TestIam_detachRolePolicyWithRetry(t *testing.T) {
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	apiErrorMock := NewApiErrorMockIamSDKClient()

	type args struct {
		ctx      context.Context
		input    *iam.DetachRolePolicyInput
		roleName *string
		client   IIamSDKClient
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
				ctx: context.Background(),
				input: &iam.DetachRolePolicyInput{
					PolicyArn: aws.String("test"),
					RoleName:  aws.String("test"),
				},
				roleName: aws.String("test"),
				client:   mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "detach role policy failure",
			args: args{
				ctx: context.Background(),
				input: &iam.DetachRolePolicyInput{
					PolicyArn: aws.String("test"),
					RoleName:  aws.String("test"),
				},
				roleName: aws.String("test"),
				client:   errorMock,
			},
			want:    fmt.Errorf("DetachRolePolicyError"),
			wantErr: true,
		},
		{
			name: "detach role policy failure for api error",
			args: args{
				ctx: context.Background(),
				input: &iam.DetachRolePolicyInput{
					PolicyArn: aws.String("test"),
					RoleName:  aws.String("test"),
				},
				roleName: aws.String("test"),
				client:   apiErrorMock,
			},
			want:    fmt.Errorf("RetryCountOverError: test, api error Throttling: Rate exceeded\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			iamClient := NewIam(tt.args.client)

			_, err := iamClient.detachRolePolicyWithRetry(tt.args.ctx, tt.args.input, tt.args.roleName, sleepTimeSecForIam)
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
	mock := NewMockIamSDKClient()
	errorMock := NewErrorMockIamSDKClient()
	notExitsMock := NewNotExistsMockForGetRoleIamSDKClient()

	type args struct {
		ctx      context.Context
		roleName *string
		client   IIamSDKClient
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
				client:   mock,
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
				client:   notExitsMock,
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
				client:   errorMock,
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("GetRoleError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ecrClient := NewIam(tt.args.client)

			output, err := ecrClient.CheckRoleExists(tt.args.ctx, tt.args.roleName)
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
