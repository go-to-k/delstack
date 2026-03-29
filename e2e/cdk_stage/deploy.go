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

// This script deploys test stacks for the `delstack cdk` Stage E2E test.
// It deploys a CDK app with 2 stages (each containing 1 stack) in different regions.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-cdk-stg-%04d", randomNum)
	}

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

	// Upload objects to S3 buckets in each region
	type bucketRegion struct {
		name   string
		region string
	}
	buckets := []bucketRegion{
		{name: options.Stage + "-stage1-bucket", region: "us-east-1"},
		{name: options.Stage + "-stage2-bucket", region: "ap-northeast-1"},
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
	color.Green("\nTo test CDK Stage integration, run:")
	color.Green("  cd e2e/cdk_stage/cdk && delstack cdk -c PJ_PREFIX=%s -f -y", options.Stage)
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
