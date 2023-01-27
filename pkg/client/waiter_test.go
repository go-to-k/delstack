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
	tests := []struct {
		name    string
		args    RetryInput[struct{}, struct{}, struct{}]
		want    error
		wantErr bool
	}{
		{
			name: "retry and then occur an error: ApiCaller returns error and RetryableChecker returns true",
			args: RetryInput[struct{}, struct{}, struct{}]{
				Ctx:            context.Background(),
				SleepTimeSec:   1,
				TargetResource: aws.String("resource"),
				Input:          &struct{}{},
				ApiCaller: func(ctx context.Context, input *struct{}, optFns ...func(*struct{})) (*struct{}, error) {
					return input, fmt.Errorf("ApiFuncError")
				},
				RetryableChecker: func(err error) bool {
					return true
				},
			},
			want:    fmt.Errorf("RetryCountOverError: resource, ApiFuncError\nRetryCount(" + strconv.Itoa(MaxRetryCount) + ") over, but failed to delete. "),
			wantErr: true,
		},
		{
			name: "do not retry and then occur an error: ApiCaller returns error and RetryableChecker returns false",
			args: RetryInput[struct{}, struct{}, struct{}]{
				Ctx:            context.Background(),
				SleepTimeSec:   1,
				TargetResource: aws.String("resource"),
				Input:          &struct{}{},
				ApiCaller: func(ctx context.Context, input *struct{}, optFns ...func(*struct{})) (*struct{}, error) {
					return input, fmt.Errorf("ApiFuncError")
				},
				RetryableChecker: func(err error) bool {
					return false
				},
			},
			want:    fmt.Errorf("ApiFuncError"),
			wantErr: true,
		},
		{
			name: "success: ApiCaller do not return error and RetryableChecker returns true but is not concerned about",
			args: RetryInput[struct{}, struct{}, struct{}]{
				Ctx:            context.Background(),
				SleepTimeSec:   1,
				TargetResource: aws.String("resource"),
				Input:          &struct{}{},
				ApiCaller: func(ctx context.Context, input *struct{}, optFns ...func(*struct{})) (*struct{}, error) {
					return input, nil
				},
				RetryableChecker: func(err error) bool {
					return true
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: ApiCaller do not return error and RetryableChecker returns false but is not concerned about",
			args: RetryInput[struct{}, struct{}, struct{}]{
				Ctx:            context.Background(),
				SleepTimeSec:   1,
				TargetResource: aws.String("resource"),
				Input:          &struct{}{},
				ApiCaller: func(ctx context.Context, input *struct{}, optFns ...func(*struct{})) (*struct{}, error) {
					return input, nil
				},
				RetryableChecker: func(err error) bool {
					return false
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: ApiCaller do not return error with empty ApiOptions",
			args: RetryInput[struct{}, struct{}, struct{}]{
				Ctx:            context.Background(),
				SleepTimeSec:   1,
				TargetResource: aws.String("resource"),
				Input:          &struct{}{},
				ApiOptions:     []func(*struct{}){},
				ApiCaller: func(ctx context.Context, input *struct{}, optFns ...func(*struct{})) (*struct{}, error) {
					return input, nil
				},
				RetryableChecker: func(err error) bool {
					return false
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "success: ApiCaller do not return error with some ApiOptions",
			args: RetryInput[struct{}, struct{}, struct{}]{
				Ctx:            context.Background(),
				SleepTimeSec:   1,
				TargetResource: aws.String("resource"),
				Input:          &struct{}{},
				ApiOptions: []func(*struct{}){
					func(*struct{}) {},
					func(*struct{}) {},
				},
				ApiCaller: func(ctx context.Context, input *struct{}, optFns ...func(*struct{})) (*struct{}, error) {
					return input, nil
				},
				RetryableChecker: func(err error) bool {
					return false
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Retry(&tt.args)
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
				retryCount:         MaxRetryCount - 1,
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
				retryCount:         MaxRetryCount,
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
				retryCount:         MaxRetryCount + 1,
				sleepTimeSec:       1,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    fmt.Errorf("RetryCountOverError: resource, API Error\nRetryCount(" + strconv.Itoa(MaxRetryCount) + ") over, but failed to delete. "),
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
