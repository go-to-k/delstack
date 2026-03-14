package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/resource"
)

func NewDeletionProtectionTestStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	// Create VPC resources (shared by EC2, RDS, and ELBv2)
	vpc := resource.NewVpc(stack)

	// Create EC2 Instance with API termination protection
	resource.NewEc2Instance(stack, vpc)

	// Create RDS DBInstance with deletion protection
	resource.NewRdsInstance(stack, vpc)

	// Create RDS Aurora DBCluster with deletion protection
	resource.NewRdsCluster(stack, vpc)

	// Create Cognito UserPool with deletion protection
	resource.NewCognitoUserPool(stack)

	// Create ELBv2 ALB with deletion protection
	resource.NewAlb(stack, vpc)

	// Create CloudWatch LogGroup with deletion protection
	resource.NewLogGroup(stack)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "DelstackDeletionProtectionTest"
	}

	terminationProtection := true
	if tp := app.Node().TryGetContext(jsii.String("TERMINATION_PROTECTION")); tp != nil {
		if tp.(string) == "false" {
			terminationProtection = false
		}
	}

	stackName := pjPrefix

	NewDeletionProtectionTestStack(app, stackName, &awscdk.StackProps{
		Env:                   env(),
		StackName:             jsii.String(stackName),
		TerminationProtection: jsii.Bool(terminationProtection),
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
