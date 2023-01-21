package client

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
)

/*
	Test Cases
*/

func TestRetry(t *testing.T) {
	type args struct {
		ctx            context.Context
		sleepTimeSec   int
		targetResource *string
		input          interface{}
		apiFunc        ApiFunc[any, any]
		retryable      func(error) bool
	}
	type want error
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "retry and then occur an error: apiFunc returns error and retryable returns true if error is not nil",
			args: args{
				ctx:            context.Background(),
				sleepTimeSec:   1,
				targetResource: aws.String("resource"),
				input:          struct{}{},
				apiFunc: func(ctx context.Context, input any) (any, error) {
					return input, fmt.Errorf("ApiFuncError")
				},
				retryable: func(err error) bool {
					return err != nil && true
				},
			},
			want:    fmt.Errorf("RetryCountOverError: resource, ApiFuncError\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
		{
			name: "do not retry and then occur an error: apiFunc returns error and retryable just returns false",
			args: args{
				ctx:            context.Background(),
				sleepTimeSec:   1,
				targetResource: aws.String("resource"),
				input:          struct{}{},
				apiFunc: func(ctx context.Context, input any) (any, error) {
					return input, fmt.Errorf("ApiFuncError")
				},
				retryable: func(err error) bool {
					return err != nil && false
				},
			},
			want:    fmt.Errorf("ApiFuncError"),
			wantErr: true,
		},
		{
			name: "success but do not retry: apiFunc do not return error and retryable returns true if error is not nil",
			args: args{
				ctx:            context.Background(),
				sleepTimeSec:   1,
				targetResource: aws.String("resource"),
				input:          struct{}{},
				apiFunc: func(ctx context.Context, input any) (any, error) {
					return input, nil
				},
				retryable: func(err error) bool {
					return err != nil
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Retry(
				&RetryInput{
					Ctx:            tt.args.ctx,
					SleepTimeSec:   tt.args.sleepTimeSec,
					TargetResource: tt.args.targetResource,
					Input:          tt.args.input,
					ApiFunc:        tt.args.apiFunc,
					Retryable:      tt.args.retryable,
				},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("Retry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("Retry() error = %v, want %v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func Test_waitForRetry(t *testing.T) {
	type args struct {
		ctx                context.Context
		retryCount         int
		sleepTimeSec       int
		targetResourceType *string
		err                error
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "sleepTimeSec = 0: not error",
			args: args{
				ctx:                context.Background(),
				retryCount:         0,
				sleepTimeSec:       0,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "retryCount = 0: not error",
			args: args{
				ctx:                context.Background(),
				retryCount:         0,
				sleepTimeSec:       1,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "retryCount = MaxRetryCount - 1: not error",
			args: args{
				ctx:                context.Background(),
				retryCount:         maxRetryCount - 1,
				sleepTimeSec:       1,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "retryCount = MaxRetryCount: not error",
			args: args{
				ctx:                context.Background(),
				retryCount:         maxRetryCount,
				sleepTimeSec:       1,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "retryCount = MaxRetryCount + 1: RetryCountOverError",
			args: args{
				ctx:                context.Background(),
				retryCount:         maxRetryCount + 1,
				sleepTimeSec:       1,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    fmt.Errorf("RetryCountOverError: resource, API Error\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := waitForRetry(
				tt.args.ctx,
				tt.args.retryCount,
				tt.args.sleepTimeSec,
				tt.args.targetResourceType,
				tt.args.err,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}

}
