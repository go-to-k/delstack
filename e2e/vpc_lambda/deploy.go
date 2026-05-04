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
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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
	CfnClient     *cloudformation.Client
}

// This script reproduces the real user-facing scenario of issue #637:
// a stack is left in DELETE_FAILED by an ordinary CloudFormation delete-stack
// because AWS Lambda has not yet released its VPC ENIs. The user then runs
// `delstack` to recover from that stuck state. Steps:
//
//  1. cdk deploy: creates VPC, Subnet, SG, VPC Lambda.
//  2. Invoke the Lambda once so AWS Lambda provisions the VPC (Hyperplane) ENIs.
//  3. Trigger an ordinary CloudFormation DeleteStack (NOT delstack). CFN deletes
//     the Lambda first, then tries to delete the Subnet / SecurityGroup; those
//     fail with "has dependencies" / "has a dependent object" because Lambda has
//     not yet released the ENIs (release is asynchronous, ~30 min).
//  4. Wait until the stack reaches DELETE_FAILED.
//
// After this script the stack is in the exact state issue #637 describes.
// Running `delstack -s <stage>` should then detect the orphan Lambda VPC ENIs,
// delete them, and delete the Subnet / SecurityGroup themselves.
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

	if err := service.triggerCfnDeleteAndWaitForFailure(); err != nil {
		color.Red("Failed to bring the stack into DELETE_FAILED: %v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("STACK IS NOW IN DELETE_FAILED (orphan Lambda VPC ENIs blocking Subnet/SG)")
	color.Green("Stack Name: %s", service.CfnStackName)
	color.Green("===================================")
	color.Yellow("To force delete this stuck stack, run:")
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
	s.CfnClient = cloudformation.NewFromConfig(cfg)

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

func (s *DeployStackService) triggerCfnDeleteAndWaitForFailure() error {
	color.Green("=== trigger ordinary CloudFormation DeleteStack and wait for DELETE_FAILED ===")

	_, err := s.CfnClient.DeleteStack(s.Ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(s.CfnStackName),
	})
	if err != nil {
		return fmt.Errorf("DeleteStack %s failed: %w", s.CfnStackName, err)
	}

	// Poll until the stack reaches DELETE_FAILED, DELETE_COMPLETE (rare: AWS released
	// the ENIs in time and CFN deleted the stack cleanly), or the timeout fires.
	deadline := time.Now().Add(20 * time.Minute)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for stack %s to reach DELETE_FAILED", s.CfnStackName)
		}

		out, err := s.CfnClient.DescribeStacks(s.Ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.CfnStackName),
		})
		if err != nil {
			// Stack may have been removed entirely (DELETE_COMPLETE drops it from
			// DescribeStacks). That means the orphan ENI condition did not occur —
			// still report it so the human can decide whether to retry.
			color.Yellow("DescribeStacks returned error (stack may be gone): %v", err)
			return fmt.Errorf("stack %s no longer exists; orphan ENI scenario was not reproduced", s.CfnStackName)
		}
		if len(out.Stacks) == 0 {
			return fmt.Errorf("stack %s no longer exists; orphan ENI scenario was not reproduced", s.CfnStackName)
		}

		status := out.Stacks[0].StackStatus
		color.Cyan("  current StackStatus: %s", status)
		switch status {
		case cfntypes.StackStatusDeleteFailed:
			color.Green("Stack reached DELETE_FAILED as expected.")
			return nil
		case cfntypes.StackStatusDeleteComplete:
			return fmt.Errorf("stack %s reached DELETE_COMPLETE; orphan ENI scenario was not reproduced (Lambda may have released ENIs in time)", s.CfnStackName)
		case cfntypes.StackStatusDeleteInProgress:
			// keep polling
		default:
			return fmt.Errorf("unexpected StackStatus while waiting for DELETE_FAILED: %s", status)
		}

		time.Sleep(15 * time.Second)
	}
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
