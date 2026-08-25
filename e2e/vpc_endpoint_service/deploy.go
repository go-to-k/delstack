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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/fatih/color"
)

const (
	region = "us-east-1"

	// consumerVpcCidr / consumerSubnetCidr are only used by the out-of-band consumer
	// VPC. They do not overlap with the provider VPC created by the CDK stack, though
	// the two VPCs are never peered.
	consumerVpcCidr    = "10.195.0.0/16"
	consumerSubnetCidr = "10.195.0.0/24"

	// stageTagKey marks every out-of-band consumer resource so that the cleanup mode
	// deletes exactly what this script created, and nothing else.
	stageTagKey = "delstack-e2e-vpc-endpoint-service"

	connectionWaitTimeout  = 5 * time.Minute
	connectionWaitInterval = 10 * time.Second
)

type Options struct {
	Profile string
	Stage   string
	Cleanup bool
}

type DeployStackService struct {
	Options       Options
	CfnPjPrefix   string
	CfnStackName  string
	ProfileOption string
	Ctx           context.Context
	CfnClient     *cloudformation.Client
	EC2Client     *ec2.Client
}

// This script reproduces the user-facing scenario of issue #656 in a deterministic
// way for E2E:
//
//  1. cdk deploy: creates the provider VPC, an internal NLB and the PrivateLink
//     endpoint service hosted on it. The stack outputs the endpoint service ID and
//     name.
//  2. Out of band (SDK only, so CloudFormation knows nothing about them): a small
//     consumer VPC, a subnet in an Availability Zone the endpoint service supports,
//     and an interface VPC endpoint connected to the endpoint service. Since the
//     service does not require acceptance, the connection reaches `Available`.
//
// `delstack` itself drives DeleteStack and waits for DELETE_FAILED via its internal
// CloudFormation delete waiter, so this script does not call DeleteStack. CFN's first
// delete pass fails on the endpoint service because a connection is still attached,
// the stack lands in DELETE_FAILED, and the new EC2VPCEndpointServiceOperator rejects
// the connection and deletes the endpoint service.
//
// The consumer resources survive the delstack run on purpose (rejecting a connection
// does not delete the consumer's endpoint), so run this script again with `-d` to
// remove them.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		// Cleanup targets resources tagged with the stage, so an auto-generated one
		// would silently match nothing and report success while the real consumer
		// resources keep running.
		if options.Cleanup {
			color.Red("-s <stage> is required with -d")
			os.Exit(1)
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-vpce-svc-%04d", randomNum)
	}

	service := NewDeployStackService(ctx, options)

	if err := service.initAWSClients(); err != nil {
		color.Red("Failed to initialize AWS clients: %v", err)
		os.Exit(1)
	}

	if options.Cleanup {
		if err := service.cleanupConsumerResources(); err != nil {
			color.Red("Failed to clean up consumer resources: %v", err)
			os.Exit(1)
		}
		color.Green("Consumer resources for stage %s have been cleaned up", service.CfnPjPrefix)
		return
	}

	if err := service.cdkDeploy(); err != nil {
		color.Red("Failed to deploy: %v", err)
		os.Exit(1)
	}

	serviceId, serviceName, err := service.fetchStackOutputs()
	if err != nil {
		color.Red("Failed to read stack outputs: %v", err)
		os.Exit(1)
	}

	vpcEndpointId, err := service.createConsumerEndpoint(serviceName)
	if err != nil {
		color.Red("Failed to create the consumer VPC endpoint: %v", err)
		os.Exit(1)
	}

	if err := service.waitForConnection(serviceId, vpcEndpointId); err != nil {
		color.Red("Failed to wait for the endpoint connection: %v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("STACK READY (a consumer endpoint is connected to the endpoint service)")
	color.Green("Stack Name: %s", service.CfnStackName)
	color.Green("===================================")
	color.Yellow("To force delete via delstack (will go through DELETE_FAILED internally):")
	color.Yellow("  delstack -s %s", service.CfnStackName)
	color.Yellow("Afterwards, remove the leftover consumer resources:")
	color.Yellow("  go run deploy.go -s %s -d", service.CfnPjPrefix)
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

	s.CfnClient = cloudformation.NewFromConfig(cfg)
	s.EC2Client = ec2.NewFromConfig(cfg)

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

func (s *DeployStackService) fetchStackOutputs() (serviceId, serviceName string, err error) {
	color.Green("=== fetch CFN stack outputs (endpoint service ID/name) ===")

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
		case "EndpointServiceId":
			serviceId = aws.ToString(o.OutputValue)
		case "EndpointServiceName":
			serviceName = aws.ToString(o.OutputValue)
		}
	}
	if serviceId == "" || serviceName == "" {
		return "", "", fmt.Errorf("missing CFN outputs EndpointServiceId / EndpointServiceName on stack %s", s.CfnStackName)
	}

	color.Green("  EndpointServiceId=%s EndpointServiceName=%s", serviceId, serviceName)
	return serviceId, serviceName, nil
}

// createConsumerEndpoint builds the consumer side entirely with the SDK: a VPC, a
// subnet in an Availability Zone the endpoint service supports, and an interface VPC
// endpoint pointing at the service.
func (s *DeployStackService) createConsumerEndpoint(serviceName string) (string, error) {
	color.Green("=== create the out-of-band consumer VPC endpoint ===")

	availabilityZone, err := s.supportedAvailabilityZone(serviceName)
	if err != nil {
		return "", err
	}

	vpcOutput, err := s.EC2Client.CreateVpc(s.Ctx, &ec2.CreateVpcInput{
		CidrBlock:         aws.String(consumerVpcCidr),
		TagSpecifications: s.tagSpecifications(types.ResourceTypeVpc, "consumer-vpc"),
	})
	if err != nil {
		return "", fmt.Errorf("CreateVpc failed: %w", err)
	}
	vpcId := aws.ToString(vpcOutput.Vpc.VpcId)
	color.Cyan("  created consumer VPC %s", vpcId)

	subnetOutput, err := s.EC2Client.CreateSubnet(s.Ctx, &ec2.CreateSubnetInput{
		VpcId:             aws.String(vpcId),
		CidrBlock:         aws.String(consumerSubnetCidr),
		AvailabilityZone:  aws.String(availabilityZone),
		TagSpecifications: s.tagSpecifications(types.ResourceTypeSubnet, "consumer-subnet"),
	})
	if err != nil {
		return "", fmt.Errorf("CreateSubnet failed: %w", err)
	}
	subnetId := aws.ToString(subnetOutput.Subnet.SubnetId)
	color.Cyan("  created consumer subnet %s in %s", subnetId, availabilityZone)

	securityGroupId, err := s.defaultSecurityGroupId(vpcId)
	if err != nil {
		return "", err
	}

	endpointOutput, err := s.EC2Client.CreateVpcEndpoint(s.Ctx, &ec2.CreateVpcEndpointInput{
		VpcId:             aws.String(vpcId),
		ServiceName:       aws.String(serviceName),
		VpcEndpointType:   types.VpcEndpointTypeInterface,
		SubnetIds:         []string{subnetId},
		SecurityGroupIds:  []string{securityGroupId},
		PrivateDnsEnabled: aws.Bool(false),
		TagSpecifications: s.tagSpecifications(types.ResourceTypeVpcEndpoint, "consumer-endpoint"),
	})
	if err != nil {
		return "", fmt.Errorf("CreateVpcEndpoint failed: %w", err)
	}
	vpcEndpointId := aws.ToString(endpointOutput.VpcEndpoint.VpcEndpointId)
	color.Cyan("  created consumer VPC endpoint %s", vpcEndpointId)

	return vpcEndpointId, nil
}

// supportedAvailabilityZone returns an AZ the endpoint service is available in. The
// consumer subnet must live in one of them, otherwise CreateVpcEndpoint fails.
func (s *DeployStackService) supportedAvailabilityZone(serviceName string) (string, error) {
	output, err := s.EC2Client.DescribeVpcEndpointServices(s.Ctx, &ec2.DescribeVpcEndpointServicesInput{
		ServiceNames: []string{serviceName},
	})
	if err != nil {
		return "", fmt.Errorf("DescribeVpcEndpointServices failed: %w", err)
	}
	for _, detail := range output.ServiceDetails {
		if len(detail.AvailabilityZones) > 0 {
			return detail.AvailabilityZones[0], nil
		}
	}

	return "", fmt.Errorf("no Availability Zone found for the endpoint service %s", serviceName)
}

func (s *DeployStackService) defaultSecurityGroupId(vpcId string) (string, error) {
	output, err := s.EC2Client.DescribeSecurityGroups(s.Ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcId}},
			{Name: aws.String("group-name"), Values: []string{"default"}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("DescribeSecurityGroups failed: %w", err)
	}
	if len(output.SecurityGroups) == 0 {
		return "", fmt.Errorf("default security group not found in VPC %s", vpcId)
	}

	return aws.ToString(output.SecurityGroups[0].GroupId), nil
}

// waitForConnection blocks until the connection shows up on the endpoint service side.
// Without it the stack could be deleted before the connection exists, and nothing
// would block the endpoint service deletion.
func (s *DeployStackService) waitForConnection(serviceId, vpcEndpointId string) error {
	color.Green("=== wait for the endpoint connection to attach to %s ===", serviceId)

	deadline := time.Now().Add(connectionWaitTimeout)
	for {
		output, err := s.EC2Client.DescribeVpcEndpointConnections(s.Ctx, &ec2.DescribeVpcEndpointConnectionsInput{
			Filters: []types.Filter{
				{Name: aws.String("service-id"), Values: []string{serviceId}},
			},
		})
		if err != nil {
			return fmt.Errorf("DescribeVpcEndpointConnections failed: %w", err)
		}

		for _, connection := range output.VpcEndpointConnections {
			if aws.ToString(connection.VpcEndpointId) != vpcEndpointId {
				continue
			}
			if connection.VpcEndpointState == types.StateAvailable || connection.VpcEndpointState == types.StatePendingAcceptance {
				color.Green("  connection %s is %s", vpcEndpointId, connection.VpcEndpointState)
				return nil
			}
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for the connection from %s to reach a blocking state", vpcEndpointId)
		}
		time.Sleep(connectionWaitInterval)
	}
}

// cleanupConsumerResources removes the out-of-band consumer resources this script
// created for the stage. delstack only rejects the connection, so the consumer
// endpoint and its VPC are still there after the stack is gone.
func (s *DeployStackService) cleanupConsumerResources() error {
	color.Green("=== clean up the out-of-band consumer resources for %s ===", s.CfnPjPrefix)

	vpcsOutput, err := s.EC2Client.DescribeVpcs(s.Ctx, &ec2.DescribeVpcsInput{
		Filters: []types.Filter{
			{Name: aws.String("tag:" + stageTagKey), Values: []string{s.CfnPjPrefix}},
		},
	})
	if err != nil {
		return fmt.Errorf("DescribeVpcs failed: %w", err)
	}
	if len(vpcsOutput.Vpcs) == 0 {
		color.Yellow("  no consumer VPC found for stage %s, nothing to do", s.CfnPjPrefix)
		return nil
	}

	for _, vpc := range vpcsOutput.Vpcs {
		vpcId := aws.ToString(vpc.VpcId)

		if err := s.deleteVpcEndpoints(vpcId); err != nil {
			return err
		}
		if err := s.deleteSubnets(vpcId); err != nil {
			return err
		}
		if _, err := s.EC2Client.DeleteVpc(s.Ctx, &ec2.DeleteVpcInput{VpcId: aws.String(vpcId)}); err != nil {
			return fmt.Errorf("DeleteVpc %s failed: %w", vpcId, err)
		}
		color.Cyan("  deleted consumer VPC %s", vpcId)
	}

	return nil
}

func (s *DeployStackService) deleteVpcEndpoints(vpcId string) error {
	output, err := s.EC2Client.DescribeVpcEndpoints(s.Ctx, &ec2.DescribeVpcEndpointsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcId}},
		},
	})
	if err != nil {
		return fmt.Errorf("DescribeVpcEndpoints failed: %w", err)
	}

	vpcEndpointIds := []string{}
	for _, endpoint := range output.VpcEndpoints {
		vpcEndpointIds = append(vpcEndpointIds, aws.ToString(endpoint.VpcEndpointId))
	}
	if len(vpcEndpointIds) == 0 {
		return nil
	}

	if _, err := s.EC2Client.DeleteVpcEndpoints(s.Ctx, &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: vpcEndpointIds,
	}); err != nil {
		return fmt.Errorf("DeleteVpcEndpoints failed: %w", err)
	}
	color.Cyan("  deleted consumer VPC endpoints %v", vpcEndpointIds)

	// The endpoint ENIs are removed asynchronously and block the subnet deletion
	// until they are gone.
	deadline := time.Now().Add(connectionWaitTimeout)
	for {
		remaining, err := s.EC2Client.DescribeVpcEndpoints(s.Ctx, &ec2.DescribeVpcEndpointsInput{
			Filters: []types.Filter{
				{Name: aws.String("vpc-id"), Values: []string{vpcId}},
			},
		})
		if err != nil {
			return fmt.Errorf("DescribeVpcEndpoints failed: %w", err)
		}
		if len(remaining.VpcEndpoints) == 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for the consumer VPC endpoints in %s to be deleted", vpcId)
		}
		time.Sleep(connectionWaitInterval)
	}
}

func (s *DeployStackService) deleteSubnets(vpcId string) error {
	output, err := s.EC2Client.DescribeSubnets(s.Ctx, &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcId}},
		},
	})
	if err != nil {
		return fmt.Errorf("DescribeSubnets failed: %w", err)
	}

	for _, subnet := range output.Subnets {
		subnetId := aws.ToString(subnet.SubnetId)
		if _, err := s.EC2Client.DeleteSubnet(s.Ctx, &ec2.DeleteSubnetInput{SubnetId: aws.String(subnetId)}); err != nil {
			return fmt.Errorf("DeleteSubnet %s failed: %w", subnetId, err)
		}
		color.Cyan("  deleted consumer subnet %s", subnetId)
	}

	return nil
}

func (s *DeployStackService) tagSpecifications(resourceType types.ResourceType, name string) []types.TagSpecification {
	return []types.TagSpecification{
		{
			ResourceType: resourceType,
			Tags: []types.Tag{
				{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("%s-%s", s.CfnPjPrefix, name))},
				{Key: aws.String(stageTagKey), Value: aws.String(s.CfnPjPrefix)},
			},
		},
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
		case "-d", "--cleanup":
			options.Cleanup = true
		case "-h", "--help":
			fmt.Println("Usage: go run deploy.go [options]")
			fmt.Println("Options:")
			fmt.Println("  -p, --profile <profile>  AWS profile name")
			fmt.Println("  -s, --stage <stage>      Stage name (default: auto-generated)")
			fmt.Println("  -d, --cleanup            Delete the out-of-band consumer resources of the stage")
			os.Exit(0)
		}
	}

	return options
}
