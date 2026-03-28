package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// E2E test for `delstack cdk` cross-region deletion.
//
// Creates 2 stacks in different regions with dependency:
//   MainStack (ap-northeast-1, Import: ExportFromEdge) -> EdgeStack (us-east-1, Export: ExportFromEdge)
//
// delstack cdk must:
// 1. Detect regions from manifest.json
// 2. Delete MainStack first (dependent), then EdgeStack
// 3. Handle cross-region AWS sessions

func NewEdgeStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	bucket := awss3.NewBucket(stack, jsii.String("EdgeBucket"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-edge-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	awscdk.NewCfnOutput(stack, jsii.String("ExportFromEdge"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(pjPrefix + "-ExportFromEdge"),
	})

	return stack
}

func NewMainStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	awss3.NewBucket(stack, jsii.String("MainBucket"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-main-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-xr"
	}

	account := os.Getenv("CDK_DEFAULT_ACCOUNT")

	edgeStack := NewEdgeStack(app, pjPrefix+"-EdgeStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("us-east-1"),
		},
		StackName: jsii.String(pjPrefix + "-EdgeStack"),
	}, pjPrefix)

	mainStack := NewMainStack(app, pjPrefix+"-MainStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("ap-northeast-1"),
		},
		StackName: jsii.String(pjPrefix + "-MainStack"),
	}, pjPrefix)
	mainStack.AddDependency(edgeStack, jsii.String("MainStack depends on EdgeStack"))

	app.Synth(nil)
}
