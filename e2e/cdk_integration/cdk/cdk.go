package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// E2E test for `delstack cdk` subcommand.
//
// Creates 2 stacks with dependency:
//   AppStack (Import: ExportFromBase) -> BaseStack (Export: ExportFromBase)
//
// Each stack has a non-empty S3 bucket to test force deletion.

func NewBaseStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	bucket := awss3.NewBucket(stack, jsii.String("BaseBucket"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-base-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.NewCfnOutput(stack, jsii.String("ExportFromBase"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(pjPrefix + "-ExportFromBase"),
	})

	return stack
}

func NewAppStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	importedArn := awscdk.Fn_ImportValue(jsii.String(pjPrefix + "-ExportFromBase"))

	bucket := awss3.NewBucket(stack, jsii.String("AppBucket"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-app-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(bucket).Add(jsii.String("DependsOn"), importedArn, nil)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-e2e"
	}

	baseStack := NewBaseStack(app, pjPrefix+"-BaseStack", &awscdk.StackProps{
		Env:       env(),
		StackName: jsii.String(pjPrefix + "-BaseStack"),
	}, pjPrefix)

	appStack := NewAppStack(app, pjPrefix+"-AppStack", &awscdk.StackProps{
		Env:       env(),
		StackName: jsii.String(pjPrefix + "-AppStack"),
	}, pjPrefix)
	appStack.AddDependency(baseStack, jsii.String("AppStack depends on BaseStack"))

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
