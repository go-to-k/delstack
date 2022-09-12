package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/operation"
	"github.com/go-to-k/delstack/option"
)

/*
	TODO: message handler or high usability messages("Started, Successfully, etc...")
	TODO: use tables or logging by one log for message outputs in OperatorCollection.RaiseNotSupportedServicesError()
*/

func main() {
	logger.NewLogger()

	opts := option.NewOption()
	_, err := opts.Parse()
	if err != nil {
		return
	}

	config, err := opts.LoadAwsConfig()
	if err != nil {
		logger.Logger.Fatal().Msgf("Error: %v\n", err.Error())
	}

	cfnOperator := operation.NewStackOperator(config)
	isRootStack := true
	if err := cfnOperator.DeleteStackResources(aws.String(opts.StackName), isRootStack); err != nil {
		logger.Logger.Fatal().Msgf("Error: %v\n", err.Error())
	}
}
