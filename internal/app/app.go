package app

import (
	"context"
	"fmt"
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

func (a *App) Run(ctx context.Context) error {
	return a.Cli.RunContext(ctx, os.Args)
}

func (a *App) getAction() func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if !a.InteractiveMode && a.StackName == "" {
			errMsg := fmt.Sprintln("The stack name must be specified in command options (-s) or a flow of the interactive mode.")
			return fmt.Errorf("StackNameNotSpecifiedError: %v", errMsg)
		}

		config, err := client.LoadAWSConfig(c.Context, a.Region, a.Profile)
		if err != nil {
			return err
		}

		var targetResourceTypes []string
		var keyword string
		continuation := true
		if a.InteractiveMode {
			targetResourceTypes, keyword, continuation = a.doInteractiveMode()
		} else {
			targetResourceTypes = resourcetype.GetResourceTypes()
		}

		if !continuation {
			return nil
		}

		operatorFactory := operation.NewOperatorFactory(config)
		cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(targetResourceTypes)

		if a.InteractiveMode && a.StackName == "" {
			stackNames, err := cloudformationStackOperator.ListStacksFilteredByKeyword(c.Context, aws.String(keyword))
			if err != nil {
				return err
			}
			if len(stackNames) == 0 {
				errMsg := fmt.Sprintf("No stacks matching the keyword %s.", keyword)
				return fmt.Errorf("NotExistsError: %v", errMsg)
			}

			stackName := a.selectStackName(stackNames)
			if stackName == "" {
				return nil
			}

			a.StackName = stackName
		}

		isRootStack := true
		operatorCollection := operation.NewOperatorCollection(config, operatorFactory, targetResourceTypes)
		operatorManager := operation.NewOperatorManager(operatorCollection)

		io.Logger.Info().Msgf("Start deletion, %v", a.StackName)

		if err := cloudformationStackOperator.DeleteCloudFormationStack(c.Context, aws.String(a.StackName), isRootStack, operatorManager); err != nil {
			return err
		}

		io.Logger.Info().Msgf("Successfully deleted, %v", a.StackName)
		return nil
	}
}

func (a *App) doInteractiveMode() ([]string, string, bool) {
	var checkboxes []string
	var keyword string

	label := "Select ResourceTypes you wish to delete even if DELETE_FAILED." +
		"\n" +
		"However, if resources of the selected ResourceTypes will not be DELETE_FAILED when the stack is deleted, the resources will be deleted even if you selected. " +
		"\n"
	opts := resourcetype.GetResourceTypes()

	if a.StackName == "" {
		stackNameLabel := "Filter a keyword of stack names(case-insensitive): "
		keyword = io.InputKeywordForFilter(stackNameLabel)
	}

	for {
		checkboxes = io.GetCheckboxes(label, opts)

		if len(checkboxes) == 0 {
			ok := io.GetYesNo("No selection?")
			if ok {
				return checkboxes, keyword, true
			}

			// The case for interruption(Ctrl + C)
			ok = io.GetYesNo("Do you want to finish?")
			if ok {
				io.Logger.Info().Msg("Finished...")
				return checkboxes, keyword, false
			}
			continue
		}

		ok := io.GetYesNo("OK?")
		if ok {
			return checkboxes, keyword, true
		}
	}
}

func (a *App) selectStackName(stackNames []string) string {
	var stackName string

	label := "Select StackName." + "\n"

	for {
		stackName = io.GetSelection(label, stackNames)

		if stackName == "" {
			io.Logger.Warn().Msg("Select StackName!")
			ok := io.GetYesNo("Do you want to finish?")
			if ok {
				io.Logger.Info().Msg("Finished...")
				return stackName
			}
			continue
		}

		ok := io.GetYesNo("OK?")
		if ok {
			return stackName
		}
	}
}
