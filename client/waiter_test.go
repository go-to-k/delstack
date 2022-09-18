package client

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/option"
)

func TestWaitForRetry(t *testing.T) {
	logger.NewLogger()

	type args struct {
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
				retryCount:         option.MaxRetryCount - 1,
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
				retryCount:         option.MaxRetryCount,
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
				retryCount:         option.MaxRetryCount + 1,
				sleepTimeSec:       1,
				targetResourceType: aws.String("resource"),
				err:                fmt.Errorf("API Error"),
			},
			want:    fmt.Errorf("RetryCountOverError: %v", "resource"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := WaitForRetry(
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
