package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/resource"
)

type VpcEndpointServiceTestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewVpcEndpointServiceTestStack(scope constructs.Construct, id string, props *VpcEndpointServiceTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	r := resource.NewVpcEndpointServiceStack(stack)

	// Output (NOT Export) the endpoint service identifiers so deploy.go can read them
	// via DescribeStacks. We deliberately do NOT set ExportName: a CFN Export on a
	// stack that goes through DELETE_FAILED makes ListImports return ValidationError
	// once the stack starts tearing down, which trips delstack's dependency-graph
	// analysis.
	awscdk.NewCfnOutput(stack, jsii.String("EndpointServiceId"), &awscdk.CfnOutputProps{
		Value: r.EndpointService.VpcEndpointServiceId(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("EndpointServiceName"), &awscdk.CfnOutputProps{
		Value: r.EndpointService.VpcEndpointServiceName(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "DelstackVpcEndpointServiceTest"
	}

	stackName := pjPrefix

	NewVpcEndpointServiceTestStack(app, stackName, &VpcEndpointServiceTestStackProps{
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
