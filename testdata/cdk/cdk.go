package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib"
)

type TestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewTestStack(scope constructs.Construct, id string, props *TestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	lib.NewEcr(stack)
	lib.NewS3Bucket(stack)
	lib.NewS3DirectoryBucket(stack, props.PjPrefix+"-Root")
	lib.NewIamGroup(stack)
	lib.NewCustomResource(stack)
	lib.NewDynamoDB(stack)
	lib.NewBackup(stack, props.PjPrefix+"-Root")

	lib.NewChildStack(stack, "Child", &lib.ChildStackProps{
		PjPrefix: props.PjPrefix,
	})
	lib.NewChildStack2(stack, "ChildTwo", &lib.ChildStack2Props{
		PjPrefix: props.PjPrefix,
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack"
	}

	stackName := pjPrefix + "-Test-Stack"

	NewTestStack(app, stackName, &TestStackProps{
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
