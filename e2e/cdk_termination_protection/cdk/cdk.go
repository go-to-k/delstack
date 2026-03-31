package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// E2E test for `delstack cdk` with TerminationProtection.
//
// Creates 2 stacks:
//   - TPStack: TerminationProtection enabled, contains an SNS topic
//   - NormalStack: No TerminationProtection, contains an SNS topic
//
// This tests that `delstack cdk -f -y` can disable TerminationProtection and delete all stacks.

func NewTPStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	awssns.NewTopic(stack, jsii.String("TPTopic"), &awssns.TopicProps{
		TopicName: jsii.String(pjPrefix + "-tp-topic"),
	})

	return stack
}

func NewNormalStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	awssns.NewTopic(stack, jsii.String("NormalTopic"), &awssns.TopicProps{
		TopicName: jsii.String(pjPrefix + "-normal-topic"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-tp-e2e"
	}

	NewTPStack(app, pjPrefix+"-TPStack", &awscdk.StackProps{
		Env:                   env(),
		StackName:             jsii.String(pjPrefix + "-TPStack"),
		TerminationProtection: jsii.Bool(true),
	}, pjPrefix)

	NewNormalStack(app, pjPrefix+"-NormalStack", &awscdk.StackProps{
		Env:       env(),
		StackName: jsii.String(pjPrefix + "-NormalStack"),
	}, pjPrefix)

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
