package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
)

// This script deploys test stacks for the `delstack cdk` cross-region E2E test.
// It deploys a CDK app with 2 stacks in different regions:
//   - EdgeStack (us-east-1)
//   - MainStack (ap-northeast-1)
//
// After deployment, run `delstack cdk` from the cdk/ directory to test
// cross-region CDK-integrated deletion.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-cdk-xr-%04d", randomNum)
	}

	// Initialize AWS clients
	var cfg aws.Config
	var err error

	if options.Profile != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithSharedConfigProfile(options.Profile),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx)
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
	os.Setenv("CDK_DEFAULT_ACCOUNT", accountID)

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s", options.Profile)
	}

	retainMode := "false"
	if options.RetainMode {
		retainMode = "true"
	}
	deployCmd := fmt.Sprintf("cd cdk && npx cdk deploy --all -c PJ_PREFIX=%s -c RETAIN_MODE=%s --require-approval never %s", options.Stage, retainMode, profileOption)
	cmd := exec.Command("bash", "-c", deployCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		color.Red("Failed to deploy with CDK: %v", err)
		os.Exit(1)
	}

	// Upload objects to S3 buckets in each region
	type bucketRegion struct {
		name   string
		region string
	}
	buckets := []bucketRegion{
		{name: options.Stage + "-edge-a-bucket", region: "us-east-1"},
		{name: options.Stage + "-main-a-bucket", region: "ap-northeast-1"},
		{name: options.Stage + "-edge-b-bucket", region: "us-east-1"},
		{name: options.Stage + "-main-b-bucket", region: "ap-northeast-1"},
	}

	for _, b := range buckets {
		regionCfg := cfg.Copy()
		regionCfg.Region = b.region
		s3Client := s3.NewFromConfig(regionCfg)

		color.Green("Uploading objects to %s (%s)...", b.name, b.region)
		for i := 1; i <= 5; i++ {
			_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
				Bucket: aws.String(b.name),
				Key:    aws.String(fmt.Sprintf("test-%d.txt", i)),
				Body:   strings.NewReader("test content"),
			})
			if err != nil {
				color.Red("Failed to upload object to %s: %v", b.name, err)
				os.Exit(1)
			}
		}
		color.Green("Uploaded 5 objects to %s", b.name)
	}

	color.Green("\n=== Deployment complete ===")
	color.Green("Stage: %s", options.Stage)
	color.Green("\nTo test cross-region CDK integration, run:")
	color.Green("  cd e2e/cdk_cross_region/cdk && delstack cdk -c PJ_PREFIX=%s -f -y", options.Stage)
}

type Options struct {
	Profile    string
	Stage      string
	RetainMode bool
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
		} else if os.Args[i] == "-r" {
			options.RetainMode = true
		}
	}
	return options
}
