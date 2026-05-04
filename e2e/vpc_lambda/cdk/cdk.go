package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/resource"
)

type VpcLambdaTestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewVpcLambdaTestStack(scope constructs.Construct, id string, props *VpcLambdaTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	r := resource.NewVpcLambdaStack(stack)

	// Export the Subnet/SG IDs so deploy.go can attach synthetic ENIs to them.
	awscdk.NewCfnOutput(stack, jsii.String("PrivateSubnetId"), &awscdk.CfnOutputProps{
		Value:      r.PrivateSubnet.SubnetId(),
		ExportName: jsii.String(props.PjPrefix + "-PrivateSubnetId"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("LambdaSgId"), &awscdk.CfnOutputProps{
		Value:      r.LambdaSg.SecurityGroupId(),
		ExportName: jsii.String(props.PjPrefix + "-LambdaSgId"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "DelstackVpcLambdaTest"
	}

	stackName := pjPrefix

	NewVpcLambdaTestStack(app, stackName, &VpcLambdaTestStackProps{
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
