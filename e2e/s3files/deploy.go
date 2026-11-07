package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/fatih/color"
)

const (
	region = "us-east-1"

	// pendingExportMessage is part of the error CloudFormation reports when the
	// file system still has data that has not been exported to the S3 bucket.
	pendingExportMessage = "has data pending export to S3"
)

type Options struct {
	Profile string
	Stage   string
	Cleanup bool
}

type DeployStackService struct {
	Options         Options
	CfnPjPrefix     string
	CfnStackName    string
	ClientStackName string
	ProfileOption   string
	Ctx             context.Context
	CfnClient       *cloudformation.Client
	SsmClient       *ssm.Client
}

// This script reproduces the user-facing scenario of issue #634:
//
//  1. cdk deploy the client stack: a VPC and an EC2 instance with the S3 Files
//     mount helper installed. It lives outside the target stack, so it keeps
//     writing while delstack deletes the target stack.
//  2. cdk deploy the target stack: a versioned bucket, an S3 Files file system
//     on it, and a mount target in the client stack's subnet.
//  3. Via SSM, mount the file system on the instance and start a background
//     loop that keeps writing files, so the file system always has data that
//     has not been exported to the bucket yet (export lags writes by about a
//     minute).
//
// delstack then deletes the target stack. CloudFormation deletes the mount
// target first and the file system right after it, which fails with "has data
// pending export to S3", and the S3FilesFileSystemOperator force-deletes it.
//
// Run the script again with -d after delstack finishes: it checks that the
// target stack really went through that DELETE_FAILED (otherwise the run tested
// nothing) and deletes the client stack.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Cleanup && options.Stage == "" {
		color.Red("-d requires -s <stage>")
		os.Exit(1)
	}
	if options.Stage == "" {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-s3files-%04d", randomNum)
	}

	service := NewDeployStackService(ctx, options)

	if err := service.initAWSClients(); err != nil {
		color.Red("Failed to initialize AWS clients: %v", err)
		os.Exit(1)
	}

	if options.Cleanup {
		verifyErr := service.verifyPendingExportFailure()
		if err := service.deleteClientStack(); err != nil {
			color.Red("Failed to delete the client stack: %v", err)
			os.Exit(1)
		}
		if verifyErr != nil {
			color.Red("Scenario was not reproduced: %v", verifyErr)
			os.Exit(1)
		}
		return
	}

	if err := service.cdkDeploy(service.ClientStackName, ""); err != nil {
		color.Red("Failed to deploy the client stack: %v", err)
		os.Exit(1)
	}
	clientOutputs, err := service.fetchStackOutputs(service.ClientStackName, "VpcId", "VpcCidr", "SubnetId", "InstanceId")
	if err != nil {
		color.Red("Failed to read client stack outputs: %v", err)
		os.Exit(1)
	}

	targetContext := fmt.Sprintf(
		"-c VPC_ID=%s -c SUBNET_ID=%s -c VPC_CIDR=%s",
		clientOutputs["VpcId"],
		clientOutputs["SubnetId"],
		clientOutputs["VpcCidr"],
	)
	if err := service.cdkDeploy(service.CfnStackName, targetContext); err != nil {
		color.Red("Failed to deploy the target stack: %v", err)
		os.Exit(1)
	}
	targetOutputs, err := service.fetchStackOutputs(service.CfnStackName, "FileSystemId")
	if err != nil {
		color.Red("Failed to read target stack outputs: %v", err)
		os.Exit(1)
	}

	if err := service.startWriting(clientOutputs["InstanceId"], targetOutputs["FileSystemId"]); err != nil {
		color.Red("Failed to start writing to the file system: %v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("STACK READY (the client instance keeps writing to the file system)")
	color.Green("Stack Name: %s", service.CfnStackName)
	color.Green("===================================")
	color.Yellow("Delete it right away, while the writes are still running:")
	color.Yellow("  delstack -s %s", service.CfnStackName)
	color.Yellow("Then verify the scenario and delete the client stack:")
	color.Yellow("  go run deploy.go -s %s -d", options.Stage)
}

func NewDeployStackService(ctx context.Context, options Options) *DeployStackService {
	cfnPjPrefix := options.Stage

	profileOption := ""
	if options.Profile != "" {
		profileOption = fmt.Sprintf("--profile %s --region %s", options.Profile, region)
	}

	return &DeployStackService{
		Options:         options,
		CfnPjPrefix:     cfnPjPrefix,
		CfnStackName:    cfnPjPrefix,
		ClientStackName: cfnPjPrefix + "-client",
		ProfileOption:   profileOption,
		Ctx:             ctx,
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

	s.CfnClient = cloudformation.NewFromConfig(cfg)
	s.SsmClient = ssm.NewFromConfig(cfg)

	return nil
}

func (s *DeployStackService) cdkDeploy(stackName, extraContext string) error {
	color.Green("=== cdk_deploy %s ===", stackName)

	cmd := fmt.Sprintf(
		"cd cdk && cdk deploy %s --require-approval never %s -c PJ_PREFIX=%s %s",
		stackName,
		s.ProfileOption,
		s.CfnPjPrefix,
		extraContext,
	)

	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("cdk deploy failed: %w", err)
	}
	return nil
}

func (s *DeployStackService) fetchStackOutputs(stackName string, keys ...string) (map[string]string, error) {
	out, err := s.CfnClient.DescribeStacks(s.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeStacks %s failed: %w", stackName, err)
	}
	if len(out.Stacks) == 0 {
		return nil, fmt.Errorf("stack %s not found", stackName)
	}

	outputs := map[string]string{}
	for _, o := range out.Stacks[0].Outputs {
		outputs[aws.ToString(o.OutputKey)] = aws.ToString(o.OutputValue)
	}
	for _, key := range keys {
		if outputs[key] == "" {
			return nil, fmt.Errorf("missing output %s on stack %s", key, stackName)
		}
	}
	return outputs, nil
}

func (s *DeployStackService) startWriting(instanceId, fileSystemId string) error {
	color.Green("=== mount %s on %s and start writing ===", fileSystemId, instanceId)

	if err := s.waitForSsmOnline(instanceId); err != nil {
		return err
	}

	commands := []string{
		// UserData installs the mount helper; wait until it has finished.
		"cloud-init status --wait > /dev/null",
		"mkdir -p /mnt/s3files",
		// The mount target can take a moment to become reachable after CREATE_COMPLETE.
		fmt.Sprintf("for i in $(seq 1 30); do mount -t s3files %s:/ /mnt/s3files && break; sleep 10; done", fileSystemId),
		"mountpoint -q /mnt/s3files",
		"mkdir -p /mnt/s3files/data",
		"nohup bash -c 'i=0; while :; do i=$((i+1)); dd if=/dev/urandom of=/mnt/s3files/data/w-$i bs=256k count=1 status=none || break; done' > /dev/null 2>&1 &",
		"sleep 30",
		"ls /mnt/s3files/data | wc -l",
	}

	out, err := s.SsmClient.SendCommand(s.Ctx, &ssm.SendCommandInput{
		InstanceIds:    []string{instanceId},
		DocumentName:   aws.String("AWS-RunShellScript"),
		TimeoutSeconds: aws.Int32(1800),
		Parameters:     map[string][]string{"commands": commands},
	})
	if err != nil {
		return fmt.Errorf("SendCommand failed: %w", err)
	}

	for {
		time.Sleep(10 * time.Second)
		inv, err := s.SsmClient.GetCommandInvocation(s.Ctx, &ssm.GetCommandInvocationInput{
			CommandId:  out.Command.CommandId,
			InstanceId: aws.String(instanceId),
		})
		if err != nil {
			// The invocation is not visible right after SendCommand.
			if strings.Contains(err.Error(), "InvocationDoesNotExist") {
				continue
			}
			return fmt.Errorf("GetCommandInvocation failed: %w", err)
		}
		switch inv.Status {
		case ssmTypes.CommandInvocationStatusPending, ssmTypes.CommandInvocationStatusInProgress, ssmTypes.CommandInvocationStatusDelayed:
			continue
		case ssmTypes.CommandInvocationStatusSuccess:
			count, err := strconv.Atoi(strings.TrimSpace(aws.ToString(inv.StandardOutputContent)))
			if err != nil || count == 0 {
				return fmt.Errorf("no files were written: %q", aws.ToString(inv.StandardOutputContent))
			}
			color.Cyan("  %d files written so far, the writer keeps running", count)
			return nil
		default:
			return fmt.Errorf("command %s: %s", inv.Status, aws.ToString(inv.StandardErrorContent))
		}
	}
}

func (s *DeployStackService) waitForSsmOnline(instanceId string) error {
	for i := 0; i < 60; i++ {
		out, err := s.SsmClient.DescribeInstanceInformation(s.Ctx, &ssm.DescribeInstanceInformationInput{
			Filters: []ssmTypes.InstanceInformationStringFilter{
				{Key: aws.String("InstanceIds"), Values: []string{instanceId}},
			},
		})
		if err != nil {
			return fmt.Errorf("DescribeInstanceInformation failed: %w", err)
		}
		if len(out.InstanceInformationList) > 0 && out.InstanceInformationList[0].PingStatus == ssmTypes.PingStatusOnline {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("instance %s did not register with SSM", instanceId)
}

// verifyPendingExportFailure checks the events of the deleted target stack for
// the FileSystem DELETE_FAILED caused by data pending export. Without it, the
// run passes even if the export finished before the delete and nothing blocked it.
func (s *DeployStackService) verifyPendingExportFailure() error {
	color.Green("=== verify that FileSystem failed with pending export data ===")

	var stackId string
	var latest time.Time
	paginator := cloudformation.NewListStacksPaginator(s.CfnClient, &cloudformation.ListStacksInput{
		StackStatusFilter: []cfnTypes.StackStatus{cfnTypes.StackStatusDeleteComplete},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(s.Ctx)
		if err != nil {
			return fmt.Errorf("ListStacks failed: %w", err)
		}
		for _, summary := range page.StackSummaries {
			if aws.ToString(summary.StackName) == s.CfnStackName && summary.CreationTime.After(latest) {
				stackId = aws.ToString(summary.StackId)
				latest = *summary.CreationTime
			}
		}
	}
	if stackId == "" {
		return fmt.Errorf("deleted stack %s not found, run delstack first", s.CfnStackName)
	}

	eventsPaginator := cloudformation.NewDescribeStackEventsPaginator(s.CfnClient, &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stackId),
	})
	for eventsPaginator.HasMorePages() {
		page, err := eventsPaginator.NextPage(s.Ctx)
		if err != nil {
			return fmt.Errorf("DescribeStackEvents failed: %w", err)
		}
		for _, event := range page.StackEvents {
			if aws.ToString(event.LogicalResourceId) == "FileSystem" &&
				event.ResourceStatus == cfnTypes.ResourceStatusDeleteFailed &&
				strings.Contains(aws.ToString(event.ResourceStatusReason), pendingExportMessage) {
				color.Cyan("  FileSystem DELETE_FAILED: %s", aws.ToString(event.ResourceStatusReason))
				return nil
			}
		}
	}
	return fmt.Errorf("stack %s has no FileSystem DELETE_FAILED event with %q", s.CfnStackName, pendingExportMessage)
}

func (s *DeployStackService) deleteClientStack() error {
	color.Green("=== delete the client stack %s ===", s.ClientStackName)

	if _, err := s.CfnClient.DeleteStack(s.Ctx, &cloudformation.DeleteStackInput{
		StackName: aws.String(s.ClientStackName),
	}); err != nil {
		return fmt.Errorf("DeleteStack failed: %w", err)
	}

	waiter := cloudformation.NewStackDeleteCompleteWaiter(s.CfnClient)
	return waiter.Wait(s.Ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.ClientStackName),
	}, 30*time.Minute)
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
		case "-d", "--cleanup":
			options.Cleanup = true
		case "-h", "--help":
			fmt.Println("Usage: go run deploy.go [options]")
			fmt.Println("Options:")
			fmt.Println("  -p, --profile <profile>  AWS profile name")
			fmt.Println("  -s, --stage <stage>      Stage name (default: auto-generated)")
			fmt.Println("  -d, --cleanup            Verify the deleted stack hit the pending-export failure, then delete the client stack")
			os.Exit(0)
		}
	}

	return options
}
