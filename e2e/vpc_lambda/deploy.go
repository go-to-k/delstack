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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
)

const (
	region = "us-east-1"

	// syntheticENIDescriptionPrefix matches the prefix internal/operation uses to
	// detect orphan AWS Lambda VPC ENIs. Using the same prefix lets the new
	// EC2SubnetOperator / EC2SecurityGroupOperator pick these synthetic ENIs up
	// exactly as they would real Lambda Hyperplane ENIs.
	syntheticENIDescriptionPrefix = "AWS Lambda VPC ENI"
)

type Options struct {
	Profile string
	Stage   string
}

type DeployStackService struct {
	Options       Options
	CfnPjPrefix   string
	CfnStackName  string
	AccountID     string
	ProfileOption string
	Ctx           context.Context
	StsClient     *sts.Client
	CfnClient     *cloudformation.Client
	EC2Client     *ec2.Client
}

// This script reproduces the user-facing scenario of issue #637 in a
// deterministic way for E2E:
//
//  1. cdk deploy: creates VPC, private Subnet, SecurityGroup. Stack outputs
//     the Subnet ID and SG ID so this script can attach ENIs to them.
//  2. CreateNetworkInterface (x2) on that Subnet+SG with description prefix
//     "AWS Lambda VPC ENI-..." so the new operators recognise them. The ENIs
//     are unattached and stay in `available` state, mirroring orphan Lambda
//     Hyperplane ENIs.
//  3. Trigger an ordinary CloudFormation DeleteStack. CFN tries to delete the
//     SecurityGroup and Subnet and fails with `DependencyViolation` because
//     of the synthetic ENIs.
//  4. Wait until the stack reaches DELETE_FAILED.
//
// Real Hyperplane ENI release timing is non-deterministic, which made the
// earlier real-Lambda approach flaky. This synthetic approach exercises the
// same operator code path the real scenario does.
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

	subnetId, sgId, err := service.fetchStackOutputs()
	if err != nil {
		color.Red("Failed to read stack outputs: %v", err)
		os.Exit(1)
	}

	if err := service.createSyntheticENIs(subnetId, sgId, 2); err != nil {
		color.Red("Failed to create synthetic ENIs: %v", err)
		os.Exit(1)
	}

	if err := service.triggerCfnDeleteAndWaitForFailure(); err != nil {
		color.Red("Failed to bring the stack into DELETE_FAILED: %v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("STACK IS NOW IN DELETE_FAILED (synthetic Lambda VPC ENIs blocking Subnet/SG)")
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
	s.CfnClient = cloudformation.NewFromConfig(cfg)
	s.EC2Client = ec2.NewFromConfig(cfg)

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

func (s *DeployStackService) fetchStackOutputs() (subnetId, sgId string, err error) {
	color.Green("=== fetch CFN stack outputs (Subnet/SG IDs) ===")

	out, err := s.CfnClient.DescribeStacks(s.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.CfnStackName),
	})
	if err != nil {
		return "", "", fmt.Errorf("DescribeStacks %s failed: %w", s.CfnStackName, err)
	}
	if len(out.Stacks) == 0 {
		return "", "", fmt.Errorf("stack %s not found", s.CfnStackName)
	}

	for _, o := range out.Stacks[0].Outputs {
		switch aws.ToString(o.OutputKey) {
		case "PrivateSubnetId":
			subnetId = aws.ToString(o.OutputValue)
		case "LambdaSgId":
			sgId = aws.ToString(o.OutputValue)
		}
	}
	if subnetId == "" || sgId == "" {
		return "", "", fmt.Errorf("missing CFN outputs PrivateSubnetId / LambdaSgId on stack %s", s.CfnStackName)
	}

	color.Green("  PrivateSubnetId=%s LambdaSgId=%s", subnetId, sgId)
	return subnetId, sgId, nil
}

func (s *DeployStackService) createSyntheticENIs(subnetId, sgId string, count int) error {
	color.Green("=== create %d synthetic Lambda VPC ENIs in %s / %s ===", count, subnetId, sgId)

	for i := 0; i < count; i++ {
		desc := fmt.Sprintf("%s-%s-%d", syntheticENIDescriptionPrefix, s.CfnPjPrefix, i)
		out, err := s.EC2Client.CreateNetworkInterface(s.Ctx, &ec2.CreateNetworkInterfaceInput{
			SubnetId:    aws.String(subnetId),
			Groups:      []string{sgId},
			Description: aws.String(desc),
		})
		if err != nil {
			return fmt.Errorf("CreateNetworkInterface failed: %w", err)
		}
		color.Cyan("  created ENI %s (description=%q)", aws.ToString(out.NetworkInterface.NetworkInterfaceId), desc)
	}

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

	deadline := time.Now().Add(15 * time.Minute)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for stack %s to reach DELETE_FAILED", s.CfnStackName)
		}

		out, err := s.CfnClient.DescribeStacks(s.Ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(s.CfnStackName),
		})
		if err != nil || len(out.Stacks) == 0 {
			return fmt.Errorf("stack %s no longer exists; the synthetic ENIs were not blocking deletion (unexpected)", s.CfnStackName)
		}

		status := out.Stacks[0].StackStatus
		color.Cyan("  current StackStatus: %s", status)
		switch status {
		case cfntypes.StackStatusDeleteFailed:
			color.Green("Stack reached DELETE_FAILED as expected.")
			return nil
		case cfntypes.StackStatusDeleteComplete:
			return fmt.Errorf("stack %s reached DELETE_COMPLETE; the synthetic ENIs were not blocking deletion (unexpected)", s.CfnStackName)
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
