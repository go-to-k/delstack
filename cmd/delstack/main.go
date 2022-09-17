package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/app"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/operation"
	"github.com/urfave/cli/v2"
)

var version = "version"

func main() {
	logger.NewLogger()
	ctx := context.TODO()
	app := app.NewApp()

	app.Cli.Version = version
	app.Cli.Action = action(app)

	if err := app.Cli.RunContext(ctx, os.Args); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}
}

func action(app *app.App) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		config, err := app.LoadAwsConfig()
		if err != nil {
			return err
		}

		logger.Logger.Info().Msgf("Start deletion, %v", app.StackName)

		cfnOperator := operation.NewStackOperator(config)
		isRootStack := true
		if err := cfnOperator.DeleteStackResources(aws.String(app.StackName), isRootStack); err != nil {
			return err
		}

		logger.Logger.Info().Msgf("Successfully deleted, %v", app.StackName)
		return nil
	}
}
