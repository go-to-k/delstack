package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/nest"
	"cdk/lib/resource"
)

type PreprocessorTestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewPreprocessorTestStack(scope constructs.Construct, id string, props *PreprocessorTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create VPC resources
	vpc := resource.NewVpc(stack, props.PjPrefix)

	// Create Lambda functions with different VPC configurations
	resource.NewLambdaVpcAttached(stack, props.PjPrefix, vpc, false) // VPC attached, IPv6 disabled
	resource.NewLambdaVpcAttached(stack, props.PjPrefix, vpc, true)  // VPC attached, IPv6 enabled
	resource.NewLambdaNoVpc(stack, props.PjPrefix)                   // No VPC

	// Create AgentCore Runtime with VPC configuration
	resource.NewAgentCoreRuntimeVpcAttached(stack, props.PjPrefix, vpc)

	// Create nested stack with VPC-attached Lambda functions
	nest.NewChildStack(stack, "Child", &nest.ChildStackProps{
		PjPrefix: props.PjPrefix,
		Vpc:      vpc,
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "DelstackPreprocessorTest"
	}

	stackName := pjPrefix

	NewPreprocessorTestStack(app, stackName, &PreprocessorTestStackProps{
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
