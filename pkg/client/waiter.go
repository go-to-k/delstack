package client

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const maxRetryCount = 10

type ApiFunc[T, U any] func(ctx context.Context, input *T) (*U, error)

type RetryInput[T, U any] struct {
	Ctx              context.Context
	SleepTimeSec     int
	TargetResource   *string
	Input            *T
	ApiCaller        ApiFunc[T, U]
	RetryableChecker func(error) bool
}

func Retry[T, U any](
	in *RetryInput[T, U],
) (*U, error) {
	retryCount := 0

	for {
		output, err := in.ApiCaller(in.Ctx, in.Input)
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
	if retryCount > maxRetryCount {
		errorDetail := err.Error() + "\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "
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
