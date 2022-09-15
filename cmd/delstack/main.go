package main

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/operation"
	"github.com/go-to-k/delstack/option"
)

func main() {
	logger.NewLogger()

	opts := option.NewOption()
	_, err := opts.Parse()
	if err != nil {
		return
	}

	config, err := opts.LoadAwsConfig()
	if err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}

	logger.Logger.Info().Msgf("Start deletion, %v", opts.StackName)

	cfnOperator := operation.NewStackOperator(config)
	isRootStack := true
	if err := cfnOperator.DeleteStackResources(aws.String(opts.StackName), isRootStack); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}

	logger.Logger.Info().Msgf("Successfully deleted, %v", opts.StackName)
}
