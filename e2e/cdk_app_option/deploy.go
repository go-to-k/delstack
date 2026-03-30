package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
)

// Minimal deploy script for --app option E2E test.
// Deploys a single stack with an SNS Topic.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-cdk-appopt-%04d", randomNum)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		color.Red("Failed to load AWS config: %v", err)
		os.Exit(1)
	}

	stsClient := sts.NewFromConfig(cfg)
	stsOutput, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		color.Red("Failed to get AWS account ID: %v", err)
		os.Exit(1)
	}

	os.Setenv("CDK_DEFAULT_ACCOUNT", *stsOutput.Account)

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s", options.Profile)
	}

	deployCmd := fmt.Sprintf("cd cdk && npx cdk deploy --all -c PJ_PREFIX=%s --require-approval never %s", options.Stage, profileOption)
	cmd := exec.Command("bash", "-c", deployCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		color.Red("Failed to deploy with CDK: %v", err)
		os.Exit(1)
	}

	color.Green("\n=== Deployment complete ===")
	color.Green("Stage: %s", options.Stage)
}

type Options struct {
	Profile string
	Stage   string
}

func parseArgs() Options {
	options := Options{}
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-p" && i+1 < len(os.Args) {
			options.Profile = os.Args[i+1]
			i++
		} else if os.Args[i] == "-s" && i+1 < len(os.Args) {
			options.Stage = os.Args[i+1]
			i++
		}
	}
	return options
}
