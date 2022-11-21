package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-to-k/delstack/internal/logger"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/internal/resourcetype"
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
				Value:       "ap-northeast-1",
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
		config, err := app.loadAwsConfig()
		if err != nil {
			return err
		}

		var targetResourceTypes []string
		continuation := true
		if app.InteractiveMode {
			targetResourceTypes, continuation = doInteractiveMode()
		} else {
			targetResourceTypes = resourcetype.GetResourceTypes()
		}

		if !continuation {
			return nil
		}

		logger.Logger.Info().Msgf("Start deletion, %v", app.StackName)

		stackOperatorFactory := operation.NewStackOperatorFactory(config)
		stackOperator := stackOperatorFactory.CreateStackOperator(targetResourceTypes)

		isRootStack := true
		operatorFactory := operation.NewOperatorFactory(config)
		operatorCollection := operation.NewOperatorCollection(config, operatorFactory, targetResourceTypes)
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

func doInteractiveMode() ([]string, bool) {
	var checkboxes []string

	for {
		checkboxes = getCheckboxes()

		if len(checkboxes) == 0 {
			logger.Logger.Warn().Msg("Select ResourceTypes!")
			ok := getYesNo("Do you want to finish?")
			if ok {
				logger.Logger.Info().Msg("Finished...")
				return checkboxes, false
			}
			continue
		}

		ok := getYesNo("OK?")
		if ok {
			return checkboxes, true
		}
	}
}

func getCheckboxes() []string {
	label := "Select ResourceTypes you wish to delete even if DELETE_FAILED." +
		"\n" +
		"However, if resources of the selected ResourceTypes will not be DELETE_FAILED when the stack is deleted, the resources will be deleted even if you selected. " +
		"\n"
	opts := resourcetype.GetResourceTypes()
	res := []string{}

	prompt := &survey.MultiSelect{
		Message: label,
		Options: opts,
	}
	survey.AskOne(prompt, &res)

	return res
}

func getYesNo(label string) bool {
	choices := "Y/n"
	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		fmt.Fprintln(os.Stderr)

		s = strings.TrimSpace(s)
		if s == "" {
			return true
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}
