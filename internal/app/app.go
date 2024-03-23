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

		config, err := client.LoadAWSConfig(c.Context, a.Region, a.Profile)
		if err != nil {
			return err
		}

		operatorFactory := operation.NewOperatorFactory(config)

		if a.InteractiveMode && len(a.StackNames.Value()) == 0 {
			keyword := a.inputKeywordForFilter()
			cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(resourcetype.GetResourceTypes())

			stacks, err := cloudformationStackOperator.ListStacksFilteredByKeyword(c.Context, aws.String(keyword))
			if err != nil {
				return err
			}
			if len(stacks) == 0 {
				errMsg := fmt.Sprintf("No stacks matching the keyword %s.", keyword)
				return fmt.Errorf("NotExistsError: %v", errMsg)
			}

			selectedStacks := a.selectStackNames(stacks)
			if len(selectedStacks) == 0 {
				return nil
			}

			// The `describeStacks` returns the stacks in descending order of CreationTime.
			// Therefore, by deleting stacks in the same order, we can delete from a new stack that is not depended on by any stack.
			for _, stack := range selectedStacks {
				a.StackNames.Set(stack)
			}
		}

		type stackItem struct {
			stackName           string
			targetResourceTypes []string
		}
		var stackItemList []stackItem
		for _, stackName := range a.StackNames.Value() {
			var targetResourceTypes []string
			continuation := true
			if a.InteractiveMode {
				targetResourceTypes, continuation = a.selectResourceTypes(stackName)
			} else {
				targetResourceTypes = resourcetype.GetResourceTypes()
			}
			if !continuation {
				return nil
			}
			stackItemList = append(stackItemList, stackItem{
				stackName:           stackName,
				targetResourceTypes: targetResourceTypes,
			})
		}

		isRootStack := true
		for _, stackItem := range stackItemList {
			operatorCollection := operation.NewOperatorCollection(config, operatorFactory, stackItem.targetResourceTypes)
			operatorManager := operation.NewOperatorManager(operatorCollection)
			cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator(stackItem.targetResourceTypes)

			io.Logger.Info().Msgf("%v: Start deletion. Please wait a few minutes...", stackItem.stackName)

			if err := cloudformationStackOperator.DeleteCloudFormationStack(c.Context, aws.String(stackItem.stackName), isRootStack, operatorManager); err != nil {
				return err
			}

			io.Logger.Info().Msgf("%v: Successfully deleted!!", stackItem.stackName)
		}
		return nil
	}
}

func (a *App) inputKeywordForFilter() string {
	label := "Filter a keyword of stack names(case-insensitive): "
	return io.InputKeywordForFilter(label)
}

func (a *App) selectResourceTypes(stackName string) ([]string, bool) {
	var checkboxes []string

	label := stackName +
		"\n" +
		"Select ResourceTypes you wish to delete even if DELETE_FAILED.\n" +
		"However, if a resource can be deleted without becoming DELETE_FAILED by the normal CloudFormation stack deletion feature, the resource will be deleted even if you do not select that resource type. " +
		"\n"

	opts := resourcetype.GetResourceTypes()

	for {
		checkboxes = io.GetCheckboxes(label, opts)

		if len(checkboxes) == 0 {
			ok := io.GetYesNo("No selection?")
			if ok {
				return checkboxes, true
			}

			// The case for interruption(Ctrl + C)
			ok = io.GetYesNo("Do you want to finish?")
			if ok {
				io.Logger.Info().Msg("Finished...")
				return checkboxes, false
			}
			continue
		}

		ok := io.GetYesNo("OK?")
		if ok {
			return checkboxes, true
		}
	}
}

func (a *App) selectStackNames(stackNameList []string) []string {
	var stackNames []string

	label := "Select StackNames." + "\n" +
		"Nested child stacks, XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks and EnableTerminationProtection stacks are not displayed." +
		"\n"

	for {
		stackNames = io.GetCheckboxes(label, stackNameList)

		if len(stackNames) == 0 {
			// The case for interruption(Ctrl + C)
			ok := io.GetYesNo("Do you want to finish?")
			if ok {
				io.Logger.Info().Msg("Finished...")
				return stackNames
			}
			continue
		}

		ok := io.GetYesNo("OK?")
		if ok {
			return stackNames
		}
	}
}
