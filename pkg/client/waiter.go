package client

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-to-k/delstack/pkg/logger"
)

const maxRetryCount = 10

func WaitForRetry(retryCount int, sleepTimeSec int, targetResource *string, err error) error {
	if retryCount > maxRetryCount {
		errorDetail := err.Error() + "\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "
		return fmt.Errorf("RetryCountOverError: %v, %v", *targetResource, errorDetail)
	}

	logger.Logger.Debug().Msg(err.Error() + "\nDon't worry. Retrying...")

	rand.Seed(time.Now().UnixNano())
	waitTime := 1
	if sleepTimeSec > 1 {
		waitTime += rand.Intn(sleepTimeSec)
	}
	time.Sleep(time.Duration(waitTime) * time.Second)

	return nil
}
