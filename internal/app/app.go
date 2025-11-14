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
	StackNames      *cli.StringSlice
	Profile         string
	Region          string
	InteractiveMode bool
	ForceMode       bool
}

func NewApp(version string) *App {
	app := App{}
	app.StackNames = cli.NewStringSlice()

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
				Usage:       "Force Mode to delete stacks including resources with the deletion policy Retain or RetainExceptOnCreate",
				Destination: &app.ForceMode,
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
		if !a.InteractiveMode && len(a.StackNames.Value()) == 0 {
			errMsg := fmt.Sprintln("At least one stack name must be specified in command options (-s) or a flow of the interactive mode (-i).")
			return fmt.Errorf("InvalidOptionError: %v", errMsg)
		}
		if a.InteractiveMode && len(a.StackNames.Value()) != 0 {
			errMsg := fmt.Sprintln("Stack names (-s) cannot be specified when using Interactive Mode (-i).")
			return fmt.Errorf("InvalidOptionError: %v", errMsg)
		}

		config, err := client.LoadAWSConfig(c.Context, a.Region, a.Profile)
		if err != nil {
			return err
		}

		operatorFactory := operation.NewOperatorFactory(config)
		cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(resourcetype.GetResourceTypes())

		deduplicatedStackNames := a.deduplicateStackNames()

		sortedStackNames, continuation, err := a.getSortedStackNames(c.Context, cloudformationStackOperator, deduplicatedStackNames)
		if err != nil {
			return err
		}
		if !continuation {
			return nil
		}

		// Explanation of deletion order in the case of multiple stacks
		if len(sortedStackNames) > 1 {
			io.Logger.Info().Msg("The stacks are removed in order of the latest creation time, taking into account dependencies.")
		}

		targetResourceTypes := resourcetype.GetResourceTypes()
		isRootStack := true
		for _, stackName := range sortedStackNames {
			operatorCollection := operation.NewOperatorCollection(config, operatorFactory, targetResourceTypes)
			operatorManager := operation.NewOperatorManager(operatorCollection)
			cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(targetResourceTypes)

			io.Logger.Info().Msgf("%v: Start deletion. Please wait a few minutes...", stackName)

			if a.ForceMode {
				if err := cloudformationStackOperator.RemoveDeletionPolicy(c.Context, aws.String(stackName)); err != nil {
					return err
				}
			}

			if err := cloudformationStackOperator.DeleteCloudFormationStack(c.Context, aws.String(stackName), isRootStack, operatorManager); err != nil {
				return err
			}

			io.Logger.Info().Msgf("%v: Successfully deleted!!", stackName)
		}
		return nil
	}
}

func (a *App) deduplicateStackNames() []string {
	deduplicatedStackNames := []string{}

	for _, stackName := range a.StackNames.Value() {
		var isDuplicated bool
		for _, deduplicatedStackName := range deduplicatedStackNames {
			if stackName == deduplicatedStackName {
				isDuplicated = true
				break
			}
		}
		if !isDuplicated {
			deduplicatedStackNames = append(deduplicatedStackNames, stackName)
		}
	}

	return deduplicatedStackNames
}

func (a *App) getSortedStackNames(ctx context.Context, cloudformationStackOperator *operation.CloudFormationStackOperator, specifiedStackNames []string) ([]string, bool, error) {
	if len(specifiedStackNames) != 0 {
		stackNames, err := cloudformationStackOperator.GetSortedStackNames(ctx, specifiedStackNames)
		if err != nil {
			return nil, false, err
		}
		return stackNames, true, nil
	}

	if a.InteractiveMode {
		keyword := a.inputKeywordForFilter()
		stacks, err := cloudformationStackOperator.ListStacksFilteredByKeyword(ctx, aws.String(keyword))
		if err != nil {
			return nil, false, err
		}

		// The `ListStacksFilteredByKeyword` with SDK's `DescribeStacks` returns the stacks in descending order of CreationTime.
		// Therefore, by deleting stacks in the same order, we can delete from a new stack that is not depended on by any stack.
		stackNames, continuation, err := a.selectStackNames(stacks)
		if err != nil {
			return nil, false, err
		}

		return stackNames, continuation, nil
	}

	// never reach here
	return nil, false, nil
}

func (a *App) inputKeywordForFilter() string {
	label := "Filter a keyword of stack names(case-insensitive): "
	return io.InputKeywordForFilter(label)
}

func (a *App) selectStackNames(stackNames []string) ([]string, bool, error) {
	label := []string{
		"Select StackNames.",
		"Nested child stacks, XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks and EnableTerminationProtection stacks are not displayed.",
	}

	selectedStackNames, continuation, err := io.GetCheckboxes(label, stackNames, false)
	if err != nil {
		return nil, false, err
	}
	return selectedStackNames, continuation, nil
}
