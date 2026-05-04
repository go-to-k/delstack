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
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
)

const (
	region = "us-east-1"
)

type Options struct {
	Profile string
	Stage   string
}

type DeployStackService struct {
	Options       Options
	CfnPjPrefix   string
	CfnStackName  string
	FunctionName  string
	AccountID     string
	ProfileOption string
	Ctx           context.Context
	StsClient     *sts.Client
	LambdaClient  *lambda.Client
}

// This script deploys a VPC + VPC-Lambda stack and intentionally leaves the Lambda
// in a state that produces orphan VPC ENIs. Steps:
//  1. cdk deploy (creates VPC, Subnet, SG, VPC Lambda).
//  2. Invoke the Lambda once so AWS Lambda provisions the VPC ENIs.
//  3. Delete the Lambda function via the Lambda SDK directly (NOT via CloudFormation).
//     AWS Lambda releases the ENIs asynchronously, so they remain in `available`
//     state for ~30 minutes, blocking subsequent Subnet/SecurityGroup deletion.
//
// After this script, running `delstack -s <stage>` should clean up the stack by
// deleting those orphan ENIs and then the Subnet/SecurityGroup themselves.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-vpc-lambda-%04d", randomNum)
	}

	service := NewDeployStackService(ctx, options)

	if err := service.initAWSClients(); err != nil {
		color.Red("Failed to initialize AWS clients: %v", err)
		os.Exit(1)
	}

	if err := service.cdkDeploy(); err != nil {
		color.Red("Failed to deploy: %v", err)
		os.Exit(1)
	}

	if err := service.invokeLambda(); err != nil {
		color.Red("Failed to invoke Lambda (ENI creation): %v", err)
		os.Exit(1)
	}

	if err := service.deleteLambdaOutOfBand(); err != nil {
		color.Red("Failed to delete Lambda out-of-band: %v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("STACK DEPLOYMENT + ORPHAN ENI SETUP COMPLETED!")
	color.Green("Stack Name: %s", service.CfnStackName)
	color.Green("===================================")
	color.Yellow("To force delete this stack, run:")
	color.Yellow("  delstack -s %s", service.CfnStackName)
}

func NewDeployStackService(ctx context.Context, options Options) *DeployStackService {
	cfnPjPrefix := options.Stage

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s --region %s", options.Profile, region)
	}

	return &DeployStackService{
		Options:       options,
		CfnPjPrefix:   cfnPjPrefix,
		CfnStackName:  cfnPjPrefix,
		FunctionName:  cfnPjPrefix + "-VpcLambda",
		ProfileOption: profileOption,
		Ctx:           ctx,
	}
}

func (s *DeployStackService) initAWSClients() error {
	var cfg aws.Config
	var err error

	if s.Options.Profile != "" {
		cfg, err = config.LoadDefaultConfig(s.Ctx,
			config.WithRegion(region),
			config.WithSharedConfigProfile(s.Options.Profile),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(s.Ctx,
			config.WithRegion(region),
		)
	}
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	s.StsClient = sts.NewFromConfig(cfg)
	s.LambdaClient = lambda.NewFromConfig(cfg)

	stsOutput, err := s.StsClient.GetCallerIdentity(s.Ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to get AWS account ID: %v", err)
	}
	s.AccountID = *stsOutput.Account

	return nil
}

func (s *DeployStackService) cdkDeploy() error {
	color.Green("=== cdk_deploy ===")

	cmd := fmt.Sprintf(
		"cd cdk && cdk deploy --all --require-approval never %s -c PJ_PREFIX=%s",
		s.ProfileOption,
		s.CfnPjPrefix,
	)

	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("cdk deploy failed: %w", err)
	}

	color.Green("CDK deployment completed successfully")
	return nil
}

func (s *DeployStackService) invokeLambda() error {
	color.Green("=== invoke Lambda to provision VPC ENIs ===")

	_, err := s.LambdaClient.Invoke(s.Ctx, &lambda.InvokeInput{
		FunctionName: aws.String(s.FunctionName),
	})
	if err != nil {
		return fmt.Errorf("invoke %s failed: %w", s.FunctionName, err)
	}

	color.Green("Lambda invoked. Waiting briefly for ENIs to be visible...")
	time.Sleep(15 * time.Second)
	return nil
}

func (s *DeployStackService) deleteLambdaOutOfBand() error {
	color.Green("=== delete Lambda function out-of-band (NOT via CFN) ===")

	_, err := s.LambdaClient.DeleteFunction(s.Ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(s.FunctionName),
	})
	if err != nil {
		return fmt.Errorf("delete function %s failed: %w", s.FunctionName, err)
	}

	color.Green("Lambda %s deleted out-of-band; orphan VPC ENIs will linger in available state", s.FunctionName)
	return nil
}

func runCommand(command string) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func parseArgs() Options {
	options := Options{}

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-p", "--profile":
			if i+1 < len(os.Args) {
				options.Profile = os.Args[i+1]
				i++
			}
		case "-s", "--stage":
			if i+1 < len(os.Args) {
				options.Stage = os.Args[i+1]
				i++
			}
		case "-h", "--help":
			fmt.Println("Usage: go run deploy.go [options]")
			fmt.Println("Options:")
			fmt.Println("  -p, --profile <profile>  AWS profile name")
			fmt.Println("  -s, --stage <stage>      Stage name (default: auto-generated)")
			os.Exit(0)
		}
	}

	return options
}
