package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// E2E test for `delstack cdk` with CDK Stages.
//
// Creates 2 stages, each with 1 stack:
//   MyStage1/AppStack (us-east-1) -- S3 Bucket
//   MyStage2/AppStack (ap-northeast-1) -- S3 Bucket
//
// Stacks are inside nested Cloud Assemblies (assembly-MyStage1/, assembly-MyStage2/).
// delstack cdk must recursively parse manifest.json to discover them.

type MyStageProps struct {
	awscdk.StageProps
	PjPrefix  string
	StageName string
}

func NewMyStage(scope constructs.Construct, id string, props *MyStageProps) awscdk.Stage {
	stage := awscdk.NewStage(scope, &id, &props.StageProps)

	stack := awscdk.NewStack(stage, jsii.String("AppStack"), &awscdk.StackProps{
		StackName: jsii.String(props.PjPrefix + "-" + props.StageName + "-AppStack"),
	})

	awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-" + props.StageName + "-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	return stage
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-stage"
	}

	account := os.Getenv("CDK_DEFAULT_ACCOUNT")

	NewMyStage(app, "MyStage1", &MyStageProps{
		StageProps: awscdk.StageProps{
			Env: &awscdk.Environment{
				Account: jsii.String(account),
				Region:  jsii.String("us-east-1"),
			},
		},
		PjPrefix:  pjPrefix,
		StageName: "stage1",
	})

	NewMyStage(app, "MyStage2", &MyStageProps{
		StageProps: awscdk.StageProps{
			Env: &awscdk.Environment{
				Account: jsii.String(account),
				Region:  jsii.String("ap-northeast-1"),
			},
		},
		PjPrefix:  pjPrefix,
		StageName: "stage2",
	})

	app.Synth(nil)
}
