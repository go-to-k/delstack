package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// E2E test for `delstack cdk -s` with glob patterns.
//
// Creates 3 top-level stacks and 1 Stage with 2 stacks:
//   - {prefix}-ApiStack
//   - {prefix}-ApiWorkerStack
//   - {prefix}-WebStack
//   - {prefix}-MyStage/{prefix}-StagedApiStack
//   - {prefix}-MyStage/{prefix}-StagedWebStack
//
// This allows testing glob patterns like "{prefix}-Api*" to select
// only ApiStack and ApiWorkerStack, leaving others untouched.
// The Stage stacks test that glob matches on flat stack names,
// not on display names with "/" separators.

type MyStageProps struct {
	awscdk.StageProps
	PjPrefix string
}

type MyStage struct {
	awscdk.Stage
}

func NewMyStage(scope constructs.Construct, id string, props *MyStageProps) *MyStage {
	stage := awscdk.NewStage(scope, &id, &props.StageProps)

	stackNames := []string{"StagedApiStack", "StagedWebStack"}
	for _, name := range stackNames {
		fullName := props.PjPrefix + "-" + name
		stack := awscdk.NewStack(stage, jsii.String(fullName), &awscdk.StackProps{
			Env:       props.Env,
			StackName: jsii.String(fullName),
		})
		awssns.NewTopic(stack, jsii.String("Topic"), nil)
	}

	return &MyStage{Stage: stage}
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-glob"
	}

	env := &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String("us-east-1"),
	}

	// Top-level stacks
	topLevelNames := []string{"ApiStack", "ApiWorkerStack", "WebStack"}
	for _, name := range topLevelNames {
		fullName := pjPrefix + "-" + name
		stack := awscdk.NewStack(app, jsii.String(fullName), &awscdk.StackProps{
			Env:       env,
			StackName: jsii.String(fullName),
		})
		awssns.NewTopic(stack, jsii.String("Topic"), nil)
	}

	// Stage with nested stacks
	NewMyStage(app, pjPrefix+"-MyStage", &MyStageProps{
		StageProps: awscdk.StageProps{
			Env: env,
		},
		PjPrefix: pjPrefix,
	})

	app.Synth(nil)
}
