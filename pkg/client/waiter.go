package client

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

const maxRetryCount = 10

type apiFunc[T, U any] func(ctx context.Context, input T) (U, error)

func Retry(
	ctx context.Context,
	sleepTimeSec int,
	targetResource *string,
	input interface{},
	f apiFunc[any, any],
	retryable func(error) bool,
) (interface{}, error) {
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		output, err := f(ctx, input)
		if retryable(err) {
			retryCount++
			if err := waitForRetry(retryCount, sleepTimeSec, targetResource, err); err != nil {
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
