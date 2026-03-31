package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
)

const (
	region = "us-east-1"
)

// This script deploys test stacks for the `delstack cdk` TerminationProtection E2E test.
// It deploys a CDK app with 2 stacks:
//   - TPStack: TerminationProtection enabled
//   - NormalStack: No TerminationProtection
//
// After deployment, run `delstack cdk -f -y` from the cdk/ directory to test
// CDK-integrated deletion with TerminationProtection handling.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-cdk-tp-e2e-%04d", randomNum)
	}

	// Initialize AWS clients
	var cfg aws.Config
	var err error

	if options.Profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(options.Profile),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
		)
	}
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
	accountID := *stsOutput.Account

	// Deploy with CDK
	os.Setenv("CDK_DEFAULT_REGION", region)
	os.Setenv("CDK_DEFAULT_ACCOUNT", accountID)

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
	color.Green("\nTo test CDK TerminationProtection deletion, run:")
	color.Green("  cd e2e/cdk_termination_protection/cdk && delstack cdk -c PJ_PREFIX=%s -f -y", options.Stage)
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
