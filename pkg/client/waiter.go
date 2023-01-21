package client

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const maxRetryCount = 10

type ApiFunc[T, U any] func(ctx context.Context, input T) (U, error)

type RetryInput struct {
	Ctx            context.Context
	SleepTimeSec   int
	TargetResource *string
	Input          interface{}
	ApiFunc        ApiFunc[any, any]
	Retryable      func(error) bool
}

func Retry(
	in *RetryInput,
) (interface{}, error) {
	retryCount := 0

	for {
		select {
		case <-in.Ctx.Done():
			return nil, in.Ctx.Err()
		default:
		}

		output, err := in.ApiFunc(in.Ctx, in.Input)
		if in.Retryable(err) {
			retryCount++
			if err := waitForRetry(retryCount, in.SleepTimeSec, in.TargetResource, err); err != nil {
				return nil, err
			}
			continue
		}
		if err != nil {
			return nil, err
		}

		return output, nil
	}
}

func waitForRetry(retryCount int, sleepTimeSec int, targetResource *string, err error) error {
	if retryCount > maxRetryCount {
		errorDetail := err.Error() + "\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "
		return fmt.Errorf("RetryCountOverError: %v, %v", *targetResource, errorDetail)
	}

	rand.Seed(time.Now().UnixNano())
	waitTime := 1
	if sleepTimeSec > 1 {
		waitTime += rand.Intn(sleepTimeSec)
	}
	time.Sleep(time.Duration(waitTime) * time.Second)

	return nil
}
