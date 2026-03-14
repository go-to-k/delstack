package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/resource"
)

type DeletionProtectionTestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewDeletionProtectionTestStack(scope constructs.Construct, id string, props *DeletionProtectionTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create VPC resources (shared by EC2, RDS, and ELBv2)
	vpc := resource.NewVpc(stack, props.PjPrefix)

	// Create EC2 Instance with API termination protection
	resource.NewEc2Instance(stack, props.PjPrefix, vpc)

	// Create RDS DBInstance with deletion protection
	resource.NewRdsInstance(stack, props.PjPrefix, vpc)

	// Create Cognito UserPool with deletion protection
	resource.NewCognitoUserPool(stack, props.PjPrefix)

	// Create ELBv2 ALB with deletion protection
	resource.NewAlb(stack, props.PjPrefix, vpc)

	// Create CloudWatch LogGroup with deletion protection
	resource.NewLogGroup(stack, props.PjPrefix)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "DelstackDeletionProtectionTest"
	}

	stackName := pjPrefix

	NewDeletionProtectionTestStack(app, stackName, &DeletionProtectionTestStackProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(stackName),
		},
		PjPrefix: pjPrefix,
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	if region == "" {
		region = "us-east-1"
	}

	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
