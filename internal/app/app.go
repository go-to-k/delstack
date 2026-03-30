package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
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
						Usage:       "Path to an existing cdk.out directory (skips synthesis)",
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
				Action: app.getCdkAction(),
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
		if a.ConcurrencyNumber < UnspecifiedConcurrencyNumber {
			errMsg := fmt.Sprintln("You must specify a positive number for the -n option.")
			return fmt.Errorf("InvalidOptionError: %v", errMsg)
		}

		io.AutoYes = a.YesMode

		config, err := client.LoadAWSConfig(c.Context, a.Region, a.Profile)
		if err != nil {
			return err
		}

		operatorFactory := operation.NewOperatorFactory(config, a.ForceMode)
		cloudformationStackOperator := operatorFactory.CreateCloudFormationStackOperator()

		deduplicatedStackNames := a.deduplicateStackNames()

		sortedStackNames, tpStackNames, continuation, err := a.getSortedStackNames(c.Context, cloudformationStackOperator, deduplicatedStackNames)
		if err != nil {
			return err
		}
		if !continuation {
			return nil
		}

		if len(tpStackNames) > 0 {
			fmt.Fprintf(os.Stderr, "The following stacks have TerminationProtection enabled:\n")
			for _, name := range tpStackNames {
				fmt.Fprintf(os.Stderr, "  - %s\n", name)
			}
			fmt.Fprintf(os.Stderr, "\nTerminationProtection will be disabled before deletion.\n")
			if !io.GetYesNo("Do you want to proceed?") {
				io.Logger.Info().Msg("Canceled.")
				return nil
			}
		}

		deleter := NewStackDeleter(a.ForceMode, a.ConcurrencyNumber)

		stackLength := len(sortedStackNames)
		if stackLength > 1 && (a.ConcurrencyNumber == UnspecifiedConcurrencyNumber || a.ConcurrencyNumber > 1) {
			var concurrency int
			if a.ConcurrencyNumber == UnspecifiedConcurrencyNumber {
				concurrency = stackLength
			} else {
				concurrency = min(a.ConcurrencyNumber, stackLength)
			}
			io.Logger.Info().Msgf("The stacks will be removed concurrently, taking into account dependencies. (concurrency: %d)", concurrency)
		}
		if err := deleter.DeleteStacksConcurrently(c.Context, sortedStackNames, config, operatorFactory); err != nil {
			return err
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

func (a *App) getSortedStackNames(ctx context.Context, cloudformationStackOperator *operation.CloudFormationStackOperator, specifiedStackNames []string) ([]string, []string, bool, error) {
	if len(specifiedStackNames) != 0 {
		stackNames, tpStackNames, err := cloudformationStackOperator.GetSortedStackNames(ctx, specifiedStackNames, a.ForceMode)
		if err != nil {
			return nil, nil, false, err
		}
		return stackNames, tpStackNames, true, nil
	}

	if a.InteractiveMode {
		keyword := a.inputKeywordForFilter()
		stacks, err := cloudformationStackOperator.ListStacksFilteredByKeyword(ctx, aws.String(keyword), a.ForceMode)
		if err != nil {
			return nil, nil, false, err
		}

		// The `ListStacksFilteredByKeyword` with SDK's `DescribeStacks` returns the stacks in descending order of CreationTime.
		stackNames, continuation, err := a.selectStackNames(stacks)
		if err != nil {
			return nil, nil, false, err
		}
		if !continuation {
			return nil, nil, false, nil
		}

		// Strip TP marker and build TP stack list
		var cleanStackNames []string
		var tpStackNames []string
		for _, name := range stackNames {
			if strings.HasPrefix(name, operation.TerminationProtectionMarker) {
				cleanName := strings.TrimPrefix(name, operation.TerminationProtectionMarker)
				cleanStackNames = append(cleanStackNames, cleanName)
				tpStackNames = append(tpStackNames, cleanName)
			} else {
				cleanStackNames = append(cleanStackNames, name)
			}
		}
		return cleanStackNames, tpStackNames, true, nil
	}

	// never reach here
	return nil, nil, false, nil
}

func (a *App) inputKeywordForFilter() string {
	label := "Filter a keyword of stack names(case-insensitive): "
	return io.InputKeywordForFilter(label)
}

func (a *App) selectStackNames(stackNames []string) ([]string, bool, error) {
	label := []string{
		"Select StackNames.",
	}
	if a.ForceMode {
		label = append(label, "Nested child stacks and XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks are not displayed.")
		label = append(label, "(* = TerminationProtection)")
	} else {
		label = append(label, "Nested child stacks, XXX_IN_PROGRESS(e.g. ROLLBACK_IN_PROGRESS) status stacks and EnableTerminationProtection stacks are not displayed.")
	}

	selectedStackNames, continuation, err := io.GetCheckboxes(label, stackNames, false)
	if err != nil {
		return nil, false, err
	}
	return selectedStackNames, continuation, nil
}
