package app

import (
	"context"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	UnspecifiedConcurrencyNumber = 0
)

type App struct {
	Cli               *cli.App
	StackNames        *cli.StringSlice
	Profile           string
	Region            string
	InteractiveMode   bool
	ForceMode         bool
	YesMode           bool
	ConcurrencyNumber int

	// CDK subcommand fields
	CdkAppPath  string
	CdkContexts *cli.StringSlice
}

func NewApp(version string) *App {
	app := App{}
	app.StackNames = cli.NewStringSlice()
	app.CdkContexts = cli.NewStringSlice()

	app.Cli = &cli.App{
		Name:  "delstack",
		Usage: "A CLI tool to force delete the entire CloudFormation stack.",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "stackName",
				Aliases:     []string{"s"},
				Usage:       "CloudFormation stack names(one or more)",
				Destination: app.StackNames,
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
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Value:       false,
				Usage:       "Force Mode to delete stacks including resources with deletion policy Retain/RetainExceptOnCreate, resources with deletion protection, and stacks with TerminationProtection",
				Destination: &app.ForceMode,
			},
			&cli.BoolFlag{
				Name:        "yes",
				Aliases:     []string{"y"},
				Value:       false,
				Usage:       "Skip confirmation prompts",
				Destination: &app.YesMode,
			},
			&cli.IntFlag{
				Name:        "concurrencyNumber",
				Aliases:     []string{"n"},
				Value:       UnspecifiedConcurrencyNumber,
				Usage:       "Specify the number of parallel stack deletions. Default is unlimited (delete all stacks in parallel).",
				Destination: &app.ConcurrencyNumber,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "cdk",
				Usage: "Delete stacks from a CDK app by synthesizing or reading an existing cdk.out",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "app",
						Aliases:     []string{"a"},
						Usage:       "CDK app command or path to cdk.out directory (e.g. 'npx ts-node bin/app.ts' or './cdk.out')",
						Destination: &app.CdkAppPath,
					},
					&cli.StringSliceFlag{
						Name:        "context",
						Aliases:     []string{"c"},
						Usage:       "CDK context values in key=value format (repeatable)",
						Destination: app.CdkContexts,
					},
					// Shared flags (same as global flags, redefined so they work after the subcommand name)
					&cli.StringSliceFlag{
						Name:        "stackName",
						Aliases:     []string{"s"},
						Usage:       "CloudFormation stack names(one or more)",
						Destination: app.StackNames,
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
					&cli.BoolFlag{
						Name:        "force",
						Aliases:     []string{"f"},
						Value:       false,
						Usage:       "Force Mode to delete stacks including resources with deletion policy Retain/RetainExceptOnCreate, resources with deletion protection, and stacks with TerminationProtection",
						Destination: &app.ForceMode,
					},
					&cli.BoolFlag{
						Name:        "yes",
						Aliases:     []string{"y"},
						Value:       false,
						Usage:       "Skip confirmation prompts",
						Destination: &app.YesMode,
					},
					&cli.IntFlag{
						Name:        "concurrencyNumber",
						Aliases:     []string{"n"},
						Value:       UnspecifiedConcurrencyNumber,
						Usage:       "Specify the number of parallel stack deletions. Default is unlimited (delete all stacks in parallel).",
						Destination: &app.ConcurrencyNumber,
					},
				},
				Action: func(c *cli.Context) error {
					return NewCdkAction(
						app.StackNames.Value(),
						app.Profile,
						app.Region,
						app.InteractiveMode,
						app.ForceMode,
						app.YesMode,
						app.ConcurrencyNumber,
						app.CdkAppPath,
						app.CdkContexts.Value(),
					).Run(c.Context)
				},
			},
		},
	}

	app.Cli.Version = version
	app.Cli.Action = func(c *cli.Context) error {
		return NewRootAction(
			app.StackNames.Value(),
			app.Profile,
			app.Region,
			app.InteractiveMode,
			app.ForceMode,
			app.YesMode,
			app.ConcurrencyNumber,
		).Run(c.Context)
	}
	app.Cli.HideHelpCommand = true

	return &app
}

func (a *App) Run(ctx context.Context) error {
	return a.Cli.RunContext(ctx, os.Args)
}
