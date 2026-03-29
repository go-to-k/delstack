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
//   EdgeStack (us-east-1, Export: ExportFromEdge) -- S3 Bucket
//   MainStack (ap-northeast-1, depends on EdgeStack via crossRegionReferences) -- S3 Bucket
//
// Uses CDK's crossRegionReferences feature to create cross-region SSM parameter-backed references.
//
// delstack cdk must:
// 1. Detect regions from manifest.json
// 2. Delete MainStack first (dependent), then EdgeStack
// 3. Handle cross-region AWS sessions

func NewEdgeStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string, isRetain bool) (awscdk.Stack, awss3.Bucket) {
	stack := awscdk.NewStack(scope, &id, props)

	removalPolicy := awscdk.RemovalPolicy_DESTROY
	if isRetain {
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	}

	bucket := awss3.NewBucket(stack, jsii.String("EdgeBucket"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-edge-bucket"),
		RemovalPolicy: removalPolicy,
	})

	return stack, bucket
}

func NewMainStack(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string, isRetain bool, edgeBucketArn *string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	removalPolicy := awscdk.RemovalPolicy_DESTROY
	if isRetain {
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	}

	bucket := awss3.NewBucket(stack, jsii.String("MainBucket"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-main-bucket"),
		RemovalPolicy: removalPolicy,
	})

	// Use cross-region reference: tag with the EdgeStack bucket ARN
	// CDK's crossRegionReferences uses SSM parameters to pass values across regions
	awscdk.Tags_Of(bucket).Add(jsii.String("EdgeBucketArn"), edgeBucketArn, nil)

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-cdk-xr"
	}

	retainMode := app.Node().TryGetContext(jsii.String("RETAIN_MODE")).(string)
	var isRetain bool
	if retainMode == "true" {
		isRetain = true
	}

	account := os.Getenv("CDK_DEFAULT_ACCOUNT")

	edgeStack, edgeBucket := NewEdgeStack(app, pjPrefix+"-EdgeStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("us-east-1"),
		},
		StackName:             jsii.String(pjPrefix + "-EdgeStack"),
		CrossRegionReferences: jsii.Bool(true),
	}, pjPrefix, isRetain)

	mainStack := NewMainStack(app, pjPrefix+"-MainStack", &awscdk.StackProps{
		Env: &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("ap-northeast-1"),
		},
		StackName:             jsii.String(pjPrefix + "-MainStack"),
		CrossRegionReferences: jsii.Bool(true),
	}, pjPrefix, isRetain, edgeBucket.BucketArn())
	mainStack.AddDependency(edgeStack, jsii.String("MainStack depends on EdgeStack"))

	app.Synth(nil)
}
