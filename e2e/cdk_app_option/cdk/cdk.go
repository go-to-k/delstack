package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/jsii-runtime-go"
)

// Minimal CDK app for testing `delstack cdk -a` option.
// Single stack with a single SNS Topic.

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-appopt"
	}

	stack := awscdk.NewStack(app, jsii.String(pjPrefix+"-AppOptStack"), &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
			Region:  jsii.String("us-east-1"),
		},
		StackName: jsii.String(pjPrefix + "-AppOptStack"),
	})

	awssns.NewTopic(stack, jsii.String("Topic"), nil)

	app.Synth(nil)
}
