package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib/nest"
	"cdk/lib/resource"
)

type TestStackProps struct {
	awscdk.StackProps
	PjPrefix string
	IsRetain bool
}

func NewTestStack(scope constructs.Construct, id string, props *TestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	if props.IsRetain {
		awscdk.RemovalPolicies_Of(scope).Retain(nil)
	}

	resource.NewEcr(stack)
	resource.NewS3Bucket(stack)
	resource.NewS3DirectoryBucket(stack, props.PjPrefix+"-Root")
	resource.NewS3TableBucket(stack, props.PjPrefix+"-Root") // can only contain [2 AWS::S3Tables::TableBucket] : Table bucket can only have up to 10 buckets created per AWS account (per region), and we want to be able to make up to 5 stacks
	resource.NewIamGroup(stack)                              // can only contain [2 AWS::IAM::Group] in this CDK app: 1 IAM user (DelstackTestUser) can only belong to 10 IAM groups, and we want to be able to make up to 5 stacks
	resource.NewCustomResources(stack)
	resource.NewDynamoDB(stack, props.PjPrefix+"-Root")
	resource.NewBackup(stack, props.PjPrefix+"-Root")

	nest.NewChildStack(stack, "Child", &nest.ChildStackProps{
		PjPrefix: props.PjPrefix,
	})
	nest.NewChildStack2(stack, "ChildTwo", &nest.ChildStack2Props{
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

	retainMode := app.Node().TryGetContext(jsii.String("RETAIN_MODE")).(string)
	var isRetain bool
	if retainMode == "true" {
		isRetain = true
	} else {
		isRetain = false
	}

	stackName := pjPrefix + "-Test-Stack"

	NewTestStack(app, stackName, &TestStackProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(stackName),
		},
		PjPrefix: pjPrefix,
		IsRetain: isRetain,
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
