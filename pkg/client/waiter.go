package client

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-to-k/delstack/internal/option"
	"github.com/go-to-k/delstack/pkg/logger"
)

func WaitForRetry(retryCount int, sleepTimeSec int, targetResource *string, err error) error {
	if retryCount > option.MaxRetryCount {
		errorDetail := err.Error() + "\nRetryCount(" + strconv.Itoa(option.MaxRetryCount) + ") over, but failed to delete. "
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
