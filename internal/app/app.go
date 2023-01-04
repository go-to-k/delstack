package app

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/urfave/cli/v2"
)

type App struct {
	Cli             *cli.App
	StackName       string
	Profile         string
	Region          string
	InteractiveMode bool
}

func NewApp(version string) *App {
	app := App{}

	app.Cli = &cli.App{
		Name:  "delstack",
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
				Usage:       "AWS region",
				Destination: &app.Region,
			},
			&cli.BoolFlag{
				Name:        "interactive",
				Aliases:     []string{"i"},
				Value:       false,
				Usage:       "Interactive Mode",
				Destination: &app.InteractiveMode,
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
		config, err := client.LoadAWSConfig(c.Context, app.Region, app.Profile)
		if err != nil {
			return err
		}

		var targetResourceTypes []string
		continuation := true
		if app.InteractiveMode {
			targetResourceTypes, continuation = io.DoInteractiveMode()
		} else {
			targetResourceTypes = resourcetype.GetResourceTypes()
		}

		if !continuation {
			return nil
		}

		io.Logger.Info().Msgf("Start deletion, %v", app.StackName)

		stackOperatorFactory := operation.NewStackOperatorFactory(config)
		stackOperator := stackOperatorFactory.CreateStackOperator(targetResourceTypes)

		isRootStack := true
		operatorFactory := operation.NewOperatorFactory(config)
		operatorCollection := operation.NewOperatorCollection(config, operatorFactory, targetResourceTypes)
		operatorManager := operation.NewOperatorManager(operatorCollection)

		if err := stackOperator.DeleteStackResources(c.Context, aws.String(app.StackName), isRootStack, operatorManager); err != nil {
			return err
		}

		io.Logger.Info().Msgf("Successfully deleted, %v", app.StackName)
		return nil
	}
}
