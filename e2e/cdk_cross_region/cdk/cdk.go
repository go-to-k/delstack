package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// E2E test for `delstack cdk` cross-region deletion with 4 stacks.
//
// Pattern A: AddDependency (explicit manifest dependency)
//   EdgeStackA (ap-northeast-1) -- S3 Bucket
//   MainStackA (us-east-1) -- S3 Bucket
//   Dependency via AddDependency (no CFn Export/Import — not supported cross-region)
//
// Pattern B: crossRegionReferences (SSM parameter-backed)
//   EdgeStackB (us-east-1) -- S3 Bucket
//   MainStackB (ap-northeast-1, references EdgeStackB bucket ARN via crossRegionReferences) -- S3 Bucket
//   Dependency auto-resolved by CDK via SSM parameters
//
// delstack cdk must handle both patterns:
// 1. Detect regions and dependencies from manifest.json
// 2. Delete MainStacks first (dependents), then EdgeStacks
// 3. Handle cross-region AWS sessions

// Pattern A: AddDependency only (manifest-level dependency, no CFn reference)

func NewEdgeStackA(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string, isRetain bool) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	removalPolicy := awscdk.RemovalPolicy_DESTROY
	if isRetain {
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	}

	awss3.NewBucket(stack, jsii.String("EdgeBucketA"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-edge-a-bucket"),
		RemovalPolicy: removalPolicy,
	})

	return stack
}

func NewMainStackA(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string, isRetain bool) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	removalPolicy := awscdk.RemovalPolicy_DESTROY
	if isRetain {
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	}

	awss3.NewBucket(stack, jsii.String("MainBucketA"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-main-a-bucket"),
		RemovalPolicy: removalPolicy,
	})

	return stack
}

// Pattern B: crossRegionReferences (SSM-backed)

func NewEdgeStackB(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string, isRetain bool) (awscdk.Stack, awss3.Bucket) {
	stack := awscdk.NewStack(scope, &id, props)

	removalPolicy := awscdk.RemovalPolicy_DESTROY
	if isRetain {
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	}

	bucket := awss3.NewBucket(stack, jsii.String("EdgeBucketB"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-edge-b-bucket"),
		RemovalPolicy: removalPolicy,
	})

	return stack, bucket
}

func NewMainStackB(scope constructs.Construct, id string, props *awscdk.StackProps, pjPrefix string, isRetain bool, edgeBucketArn *string) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	removalPolicy := awscdk.RemovalPolicy_DESTROY
	if isRetain {
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	}

	bucket := awss3.NewBucket(stack, jsii.String("MainBucketB"), &awss3.BucketProps{
		BucketName:    jsii.String(pjPrefix + "-main-b-bucket"),
		RemovalPolicy: removalPolicy,
	})
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

	var isRetain bool
	if retainMode, ok := app.Node().TryGetContext(jsii.String("RETAIN_MODE")).(string); ok && retainMode == "true" {
		isRetain = true
	}

	account := os.Getenv("CDK_DEFAULT_ACCOUNT")

	usEast1Env := &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String("us-east-1"),
	}
	apNortheast1Env := &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String("ap-northeast-1"),
	}

	// Pattern A: AddDependency (EdgeA in ap-northeast-1, MainA in us-east-1)
	edgeStackA := NewEdgeStackA(app, pjPrefix+"-EdgeStackA", &awscdk.StackProps{
		Env:       apNortheast1Env,
		StackName: jsii.String(pjPrefix + "-EdgeStackA"),
	}, pjPrefix, isRetain)

	mainStackA := NewMainStackA(app, pjPrefix+"-MainStackA", &awscdk.StackProps{
		Env:       usEast1Env,
		StackName: jsii.String(pjPrefix + "-MainStackA"),
	}, pjPrefix, isRetain)
	mainStackA.AddDependency(edgeStackA, jsii.String("MainStackA depends on EdgeStackA"))

	// Pattern B: crossRegionReferences (EdgeB in us-east-1, MainB in ap-northeast-1)
	edgeStackB, edgeBucketB := NewEdgeStackB(app, pjPrefix+"-EdgeStackB", &awscdk.StackProps{
		Env:                   usEast1Env,
		StackName:             jsii.String(pjPrefix + "-EdgeStackB"),
		CrossRegionReferences: jsii.Bool(true),
	}, pjPrefix, isRetain)

	mainStackB := NewMainStackB(app, pjPrefix+"-MainStackB", &awscdk.StackProps{
		Env:                   apNortheast1Env,
		StackName:             jsii.String(pjPrefix + "-MainStackB"),
		CrossRegionReferences: jsii.Bool(true),
	}, pjPrefix, isRetain, edgeBucketB.BucketArn())
	_ = mainStackB
	_ = edgeStackB

	app.Synth(nil)
}
