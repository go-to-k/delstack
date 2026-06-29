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
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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
	AccountID     string
	ProfileOption string
	Ctx           context.Context
	StsClient     *sts.Client
	CfnClient     *cloudformation.Client
}

// This script deploys a stack whose CREATE intentionally fails, leaving a phantom
// DELETE_FAILED resource (AWS::Cognito::UserPoolUICustomizationAttachment created without a
// UserPoolDomain). The stack is expected to end in ROLLBACK_FAILED so that delstack can then
// be run against it to verify it force-deletes such create-failed phantoms (issue #647).
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		// Generate a random number using current time as seed
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-create-failed-%04d", randomNum)
	}

	service := NewDeployStackService(ctx, options)

	// Initialize AWS clients
	if err := service.initAWSClients(); err != nil {
		color.Red("Failed to initialize AWS clients: %v", err)
		os.Exit(1)
	}

	// The CREATE is designed to fail, so a non-zero exit from cdk deploy is expected, not fatal.
	if err := service.cdkDeploy(); err != nil {
		color.Yellow("cdk deploy failed as expected (the stack CREATE is designed to fail): %v", err)
	}

	// Verify the stack was left as a phantom DELETE_FAILED in a *_FAILED state.
	if err := service.verifyPhantom(); err != nil {
		color.Red("%v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("CREATE-FAILED PHANTOM STACK IS READY!")
	color.Green("Stack Name: %s", service.CfnStackName)
	color.Green("===================================")
	color.Yellow("To delete this stack, run:")
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

	// Get account ID
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

	return nil
}

// verifyPhantom confirms the stack was left in a *_FAILED state with at least one
// DELETE_FAILED resource (the phantom), which is the precondition this E2E exercises.
func (s *DeployStackService) verifyPhantom() error {
	color.Green("=== verify_phantom ===")

	out, err := s.CfnClient.DescribeStacks(s.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.CfnStackName),
	})
	if err != nil {
		return fmt.Errorf("failed to describe stack %s: %w", s.CfnStackName, err)
	}
	if len(out.Stacks) == 0 {
		return fmt.Errorf("stack %s not found; expected it to remain in a *_FAILED state with a phantom", s.CfnStackName)
	}

	status := out.Stacks[0].StackStatus
	color.Green("Stack status: %s", status)

	deleteFailed, err := s.listDeleteFailedResources()
	if err != nil {
		return err
	}
	for _, r := range deleteFailed {
		color.Yellow("  DELETE_FAILED phantom: %s (%s)", aws.ToString(r.LogicalResourceId), aws.ToString(r.ResourceType))
	}

	if status != types.StackStatusRollbackFailed {
		return fmt.Errorf("expected stack %s to be ROLLBACK_FAILED with a phantom left behind, but it is %s; the CREATE may not have produced a phantom in this account/region", s.CfnStackName, status)
	}
	if len(deleteFailed) == 0 {
		return fmt.Errorf("stack %s is %s but has no DELETE_FAILED resources; no phantom was produced", s.CfnStackName, status)
	}

	return nil
}

func (s *DeployStackService) listDeleteFailedResources() ([]types.StackResourceSummary, error) {
	deleteFailed := []types.StackResourceSummary{}

	var nextToken *string
	for {
		out, err := s.CfnClient.ListStackResources(s.Ctx, &cloudformation.ListStackResourcesInput{
			StackName: aws.String(s.CfnStackName),
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list stack resources for %s: %w", s.CfnStackName, err)
		}
		for _, r := range out.StackResourceSummaries {
			if r.ResourceStatus == types.ResourceStatusDeleteFailed {
				deleteFailed = append(deleteFailed, r)
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			break
		}
	}

	return deleteFailed, nil
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
