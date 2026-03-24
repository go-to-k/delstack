package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/resource"
)

type LambdaEdgeTestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewLambdaEdgeTestStack(scope constructs.Construct, id string, props *LambdaEdgeTestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	resource.NewLambdaEdge(stack)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := "DelstackLambdaEdgeTest"
	if v := app.Node().TryGetContext(jsii.String("PJ_PREFIX")); v != nil {
		pjPrefix = v.(string)
	}

	stackName := pjPrefix

	NewLambdaEdgeTestStack(app, stackName, &LambdaEdgeTestStackProps{
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
