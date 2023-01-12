package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

/*
	Test Cases
*/

func TestEcr_DeleteRepository(t *testing.T) {
	ctx := context.Background()
	mock := NewMockEcrSDKClient()
	errorMock := NewErrorMockEcrSDKClient()

	type args struct {
		ctx            context.Context
		repositoryName *string
		client         IEcrSDKClient
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
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete repository failure",
			args: args{
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         errorMock,
			},
			want:    fmt.Errorf("DeleteRepositoryError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ecrClient := NewEcr(tt.args.client)

			err := ecrClient.DeleteRepository(tt.args.ctx, tt.args.repositoryName)
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
	ctx := context.Background()
	mock := NewMockEcrSDKClient()
	errorMock := NewErrorMockEcrSDKClient()
	notExitsMock := NewNotExistsMockForDescribeRepositoriesEcrSDKClient()

	type args struct {
		ctx            context.Context
		repositoryName *string
		client         IEcrSDKClient
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
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         mock,
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
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         notExitsMock,
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
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         errorMock,
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("DescribeRepositoriesError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ecrClient := NewEcr(tt.args.client)

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
