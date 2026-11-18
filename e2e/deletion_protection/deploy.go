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
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/fatih/color"
)

const (
	region = "us-east-1"
)

type Options struct {
	Profile                 string
	Stage                   string
	NoTerminationProtection bool
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
	Route53Client *route53.Client
}

// This script allows you to deploy the deletion protection test stack for delstack testing.
func main() {
	ctx := context.Background()
	options := parseArgs()

	if options.Stage == "" {
		// Generate a random number using current time as seed
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomNum := r.Intn(10000)
		options.Stage = fmt.Sprintf("delstack-deletion-protection-%04d", randomNum)
	}

	service := NewDeployStackService(ctx, options)

	// Initialize AWS clients
	if err := service.initAWSClients(); err != nil {
		color.Red("Failed to initialize AWS clients: %v", err)
		os.Exit(1)
	}

	// Deploy using CDK
	if err := service.cdkDeploy(); err != nil {
		color.Red("Failed to deploy: %v", err)
		os.Exit(1)
	}

	// Create resource record set protections outside of CFn (CFn does not support them)
	color.Green("=== protect_hosted_zone_record_sets ===")
	if err := service.protectHostedZoneRecordSets(); err != nil {
		color.Red("Failed to protect hosted zone record sets: %v", err)
		os.Exit(1)
	}

	color.Green("===================================")
	color.Green("STACK DEPLOYMENT COMPLETED!")
	color.Green("Stack Name: %s", service.CfnStackName)
	color.Green("===================================")
	color.Yellow("To test deletion protection error, run:")
	color.Yellow("  delstack -s %s", service.CfnStackName)
	color.Yellow("")
	color.Yellow("To force delete (disable protection and delete), run:")
	color.Yellow("  delstack -s %s -f", service.CfnStackName)
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
	s.Route53Client = route53.NewFromConfig(cfg)

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

	terminationProtection := "true"
	if s.Options.NoTerminationProtection {
		terminationProtection = "false"
	}

	cmd := fmt.Sprintf(
		"cd cdk && cdk deploy --all --require-approval never %s -c PJ_PREFIX=%s -c TERMINATION_PROTECTION=%s",
		s.ProfileOption,
		s.CfnPjPrefix,
		terminationProtection,
	)

	if err := runCommand(cmd); err != nil {
		return fmt.Errorf("cdk deploy failed: %w", err)
	}

	color.Green("CDK deployment completed successfully")
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
		case "-t", "--no-tp":
			options.NoTerminationProtection = true
		case "-h", "--help":
			fmt.Println("Usage: go run deploy.go [options]")
			fmt.Println("Options:")
			fmt.Println("  -p, --profile <profile>  AWS profile name")
			fmt.Println("  -s, --stage <stage>      Stage name (default: auto-generated)")
			fmt.Println("  -t, --no-tp              Disable stack TerminationProtection (default: enabled)")
			os.Exit(0)
		}
	}

	return options
}

// Creates a record set outside of CFn and protects it, together with the default NS record at the zone apex.
func (s *DeployStackService) protectHostedZoneRecordSets() error {
	output, err := s.CfnClient.DescribeStackResources(s.Ctx, &cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(s.CfnStackName),
	})
	if err != nil {
		return fmt.Errorf("failed to describe stack resources: %v", err)
	}

	for _, resource := range output.StackResources {
		if aws.ToString(resource.ResourceType) != "AWS::Route53::HostedZone" {
			continue
		}

		hostedZoneId := resource.PhysicalResourceId
		hostedZone, err := s.Route53Client.GetHostedZone(s.Ctx, &route53.GetHostedZoneInput{
			Id: hostedZoneId,
		})
		if err != nil {
			return fmt.Errorf("failed to get hosted zone %s: %v", aws.ToString(hostedZoneId), err)
		}
		zoneName := aws.ToString(hostedZone.HostedZone.Name)
		protectedName := "protected." + zoneName

		if _, err := s.Route53Client.ChangeResourceRecordSets(s.Ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: hostedZoneId,
			ChangeBatch: &route53types.ChangeBatch{
				Changes: []route53types.Change{
					{
						Action: route53types.ChangeActionCreate,
						ResourceRecordSet: &route53types.ResourceRecordSet{
							Name:            aws.String(protectedName),
							Type:            route53types.RRTypeA,
							TTL:             aws.Int64(300),
							ResourceRecords: []route53types.ResourceRecord{{Value: aws.String("192.0.2.1")}},
						},
					},
				},
			},
		}); err != nil {
			return fmt.Errorf("failed to create record set in hosted zone %s: %v", aws.ToString(hostedZoneId), err)
		}

		if _, err := s.Route53Client.CreateResourceRecordSetProtections(s.Ctx, &route53.CreateResourceRecordSetProtectionsInput{
			HostedZoneId: hostedZoneId,
			Protections: []route53types.ResourceRecordSetProtectionConfig{
				{RecordName: aws.String(protectedName), RecordType: route53types.RRTypeA},
				{RecordName: aws.String(zoneName), RecordType: route53types.RRTypeNs},
			},
		}); err != nil {
			return fmt.Errorf("failed to create record set protections in hosted zone %s: %v", aws.ToString(hostedZoneId), err)
		}
		fmt.Printf("Protected record sets in hosted zone %s (%s)\n", aws.ToString(hostedZoneId), zoneName)
	}

	return nil
}
