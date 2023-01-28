package client

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const MaxRetryCount = 10

// T: Input type for API Request.
// U: Output type for API Response.
// V: Option Functions type for API Request.
type RetryInput[T, U, V any] struct {
	Ctx              context.Context
	SleepTimeSec     int
	TargetResource   *string
	Input            *T
	ApiOptions       []func(*V)
	ApiCaller        func(ctx context.Context, input *T, optFns ...func(*V)) (*U, error)
	RetryableChecker func(error) bool
}

// T: Input type for API Request.
// U: Output type for API Response.
// V: Option Functions type for API Request.
func Retry[T, U, V any](
	in *RetryInput[T, U, V],
) (*U, error) {
	retryCount := 0

	for {
		output, err := in.ApiCaller(in.Ctx, in.Input, in.ApiOptions...)
		if err == nil {
			return output, nil
		}

		if in.RetryableChecker(err) {
			retryCount++
			if err := waitForRetry(in.Ctx, retryCount, in.SleepTimeSec, in.TargetResource, err); err != nil {
				return nil, err
			}
			continue
		}
		return nil, err
	}
}

func waitForRetry(ctx context.Context, retryCount int, sleepTimeSec int, targetResource *string, err error) error {
	if retryCount > MaxRetryCount {
		errorDetail := err.Error() + "\nRetryCount(" + strconv.Itoa(MaxRetryCount) + ") over, but failed to delete. "
		return fmt.Errorf("RetryCountOverError: %v, %v", *targetResource, errorDetail)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(getRandomSleepTime(sleepTimeSec)):
	}
	return nil
}

func getRandomSleepTime(sleepTimeSec int) time.Duration {
	rand.Seed(time.Now().UnixNano())
	waitTime := 1
	if sleepTimeSec > 1 {
		waitTime += rand.Intn(sleepTimeSec)
	}
	return time.Duration(waitTime) * time.Second
}
