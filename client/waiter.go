package client

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/option"
)

func WaitForRetry(retryCount int, sleepTimeSec int, targetResourceType *string, err error) error {
	if retryCount > option.MaxRetryCount {
		logger.Logger.Warn().Msg(err.Error() + "\nRetryCount(" + strconv.Itoa(option.MaxRetryCount) + ") over, but failed to delete. ")
		return fmt.Errorf("RetryCountOverError: %v", *targetResourceType)
	}

	logger.Logger.Warn().Msg(err.Error() + "\nRetrying...")
	time.Sleep(time.Duration(sleepTimeSec) * time.Second)

	return nil
}
