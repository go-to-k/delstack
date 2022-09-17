package app

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/urfave/cli/v2"
)

var ConcurrencyNum = runtime.NumCPU()

type App struct {
	Cli       *cli.App
	StackName string
	Profile   string
	Region    string
}

func NewApp() *App {
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
				Value:       "ap-northeast-1",
				Usage:       "CloudFormation stack name",
				Destination: &app.Region,
			},
		},
	}

	return &app
}

func (app *App) LoadAwsConfig() (aws.Config, error) {
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
