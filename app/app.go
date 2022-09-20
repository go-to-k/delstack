package app

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/operation"
	"github.com/go-to-k/delstack/option"
	"github.com/urfave/cli/v2"
)

type App struct {
	Cli       *cli.App
	StackName string
	Profile   string
	Region    string
}

func NewApp(version string) *App {
	app := App{}

	app.Cli = &cli.App{
		Name:  option.AppName,
		Usage: "A CLI tool to force delete the entire CloudFormation stack.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "stackName",
				Aliases:     []string{"s"},
				Usage:       "CloudFormation stack name",
				Required:    true,
				Destination: &app.StackName,
			},
			&cli.StringFlag{
				Name:        "profile",
				Aliases:     []string{"p"},
				Usage:       "AWS profile name",
				Destination: &app.Profile,
			},
			&cli.StringFlag{
				Name:        "region",
				Aliases:     []string{"r"},
				Value:       "ap-northeast-1",
				Usage:       "CloudFormation stack name",
				Destination: &app.Region,
			},
		},
	}

	app.Cli.Version = version
	app.Cli.Action = app.getAction()
	app.Cli.HideHelpCommand = true

	return &app
}

func (app *App) Run(ctx context.Context) error {
	return app.Cli.RunContext(ctx, os.Args)
}

func (app *App) getAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		config, err := app.loadAwsConfig()
		if err != nil {
			return err
		}

		logger.Logger.Info().Msgf("Start deletion, %v", app.StackName)

		stackOperatorFactory := operation.NewStackOperatorFactory(config)
		stackOperator := stackOperatorFactory.CreateStackOperator()

		isRootStack := true
		operatorFactory := operation.NewOperatorFactory(config)
		operatorCollection := operation.NewOperatorCollection(config, operatorFactory)
		operatorManager := operation.NewOperatorManager(operatorCollection)

		if err := stackOperator.DeleteStackResources(aws.String(app.StackName), isRootStack, operatorManager); err != nil {
			return err
		}

		logger.Logger.Info().Msgf("Successfully deleted, %v", app.StackName)
		return nil
	}
}

func (app *App) loadAwsConfig() (aws.Config, error) {
	var (
		cfg aws.Config
		err error
	)

	if app.Profile != "" {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(app.Region), config.WithSharedConfigProfile(app.Profile))
	} else {
		cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(app.Region))
	}

	return cfg, err
}
