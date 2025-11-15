package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Example 3: Complex Dependencies (Multiple Levels of Parallelism)
//
// Stack Configuration:
//   A (Export: ExportA)
//   B (Export: ExportB)
//   C (Import: ExportA, Export: ExportC)
//   D (Import: ExportA, Export: ExportD)
//   E (Import: ExportB, Export: ExportE)
//   F (Import: ExportC, ExportD, ExportE)
//
// Dependencies:
//   C → A
//   D → A
//   E → B
//   F → C
//   F → D
//   F → E
//
// Dynamic Deletion Flow:
//   Initial queue: [F]  (reverse in-degree 0)
//
//   Step 1: Start deleting F
//   Step 2: F completes → decrease C's, D's, E's reverse in-degree (all 1→0) → add C, D, E to queue
//   Step 3: Start deleting C, D, E concurrently
//   Step 4: E completes → decrease B's reverse in-degree (1→0) → add B to queue
//   Step 5: Start deleting B (doesn't wait for C, D!)
//   Step 6: C completes → decrease A's reverse in-degree (2→1)
//   Step 7: D completes → decrease A's reverse in-degree (1→0) → add A to queue
//   Step 8: Start deleting A
//   Step 9: B and A complete → all done

type StackAProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewStackA(scope constructs.Construct, id string, props *StackAProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create a minimal S3 bucket
	bucket := awss3.NewBucket(stack, jsii.String("BucketA"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-stack-a-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Export bucket ARN
	awscdk.NewCfnOutput(stack, jsii.String("ExportA"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(props.PjPrefix + "-ExportA"),
	})

	return stack
}

type StackBProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewStackB(scope constructs.Construct, id string, props *StackBProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create a minimal S3 bucket
	bucket := awss3.NewBucket(stack, jsii.String("BucketB"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-stack-b-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Export bucket ARN
	awscdk.NewCfnOutput(stack, jsii.String("ExportB"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(props.PjPrefix + "-ExportB"),
	})

	return stack
}

type StackCProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewStackC(scope constructs.Construct, id string, props *StackCProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Import from Stack A
	importedBucketArn := awscdk.Fn_ImportValue(jsii.String(props.PjPrefix + "-ExportA"))

	// Create a minimal S3 bucket with a tag referencing the imported value
	bucket := awss3.NewBucket(stack, jsii.String("BucketC"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-stack-c-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(bucket).Add(jsii.String("DependsOn"), importedBucketArn, nil)

	// Export bucket ARN
	awscdk.NewCfnOutput(stack, jsii.String("ExportC"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(props.PjPrefix + "-ExportC"),
	})

	return stack
}

type StackDProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewStackD(scope constructs.Construct, id string, props *StackDProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Import from Stack A
	importedBucketArn := awscdk.Fn_ImportValue(jsii.String(props.PjPrefix + "-ExportA"))

	// Create a minimal S3 bucket with a tag referencing the imported value
	bucket := awss3.NewBucket(stack, jsii.String("BucketD"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-stack-d-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(bucket).Add(jsii.String("DependsOn"), importedBucketArn, nil)

	// Export bucket ARN
	awscdk.NewCfnOutput(stack, jsii.String("ExportD"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(props.PjPrefix + "-ExportD"),
	})

	return stack
}

type StackEProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewStackE(scope constructs.Construct, id string, props *StackEProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Import from Stack B
	importedBucketArn := awscdk.Fn_ImportValue(jsii.String(props.PjPrefix + "-ExportB"))

	// Create a minimal S3 bucket with a tag referencing the imported value
	bucket := awss3.NewBucket(stack, jsii.String("BucketE"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-stack-e-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	awscdk.Tags_Of(bucket).Add(jsii.String("DependsOn"), importedBucketArn, nil)

	// Export bucket ARN
	awscdk.NewCfnOutput(stack, jsii.String("ExportE"), &awscdk.CfnOutputProps{
		Value:      bucket.BucketArn(),
		ExportName: jsii.String(props.PjPrefix + "-ExportE"),
	})

	return stack
}

type StackFProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewStackF(scope constructs.Construct, id string, props *StackFProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Import from Stack C, D, E
	importedBucketArnC := awscdk.Fn_ImportValue(jsii.String(props.PjPrefix + "-ExportC"))
	importedBucketArnD := awscdk.Fn_ImportValue(jsii.String(props.PjPrefix + "-ExportD"))
	importedBucketArnE := awscdk.Fn_ImportValue(jsii.String(props.PjPrefix + "-ExportE"))

	// Create a minimal S3 bucket
	awss3.NewBucket(stack, jsii.String("BucketF"), &awss3.BucketProps{
		BucketName:    jsii.String(props.PjPrefix + "-stack-f-bucket"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Create outputs to ensure dependencies
	awscdk.NewCfnOutput(stack, jsii.String("DependsOnC"), &awscdk.CfnOutputProps{
		Value:       importedBucketArnC,
		Description: jsii.String("Depends on Stack C"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("DependsOnD"), &awscdk.CfnOutputProps{
		Value:       importedBucketArnD,
		Description: jsii.String("Depends on Stack D"),
	})
	awscdk.NewCfnOutput(stack, jsii.String("DependsOnE"), &awscdk.CfnOutputProps{
		Value:       importedBucketArnE,
		Description: jsii.String("Depends on Stack E"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	pjPrefix := app.Node().TryGetContext(jsii.String("PJ_PREFIX")).(string)
	if pjPrefix == "" {
		pjPrefix = "delstack-dependency"
	}

	// Create all 6 stacks
	stackA := NewStackA(app, pjPrefix+"-Stack-A", &StackAProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix + "-Stack-A"),
		},
		PjPrefix: pjPrefix,
	})

	stackB := NewStackB(app, pjPrefix+"-Stack-B", &StackBProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix + "-Stack-B"),
		},
		PjPrefix: pjPrefix,
	})

	stackC := NewStackC(app, pjPrefix+"-Stack-C", &StackCProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix + "-Stack-C"),
		},
		PjPrefix: pjPrefix,
	})
	stackC.AddDependency(stackA, jsii.String("Stack C depends on Stack A"))

	stackD := NewStackD(app, pjPrefix+"-Stack-D", &StackDProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix + "-Stack-D"),
		},
		PjPrefix: pjPrefix,
	})
	stackD.AddDependency(stackA, jsii.String("Stack D depends on Stack A"))

	stackE := NewStackE(app, pjPrefix+"-Stack-E", &StackEProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix + "-Stack-E"),
		},
		PjPrefix: pjPrefix,
	})
	stackE.AddDependency(stackB, jsii.String("Stack E depends on Stack B"))

	stackF := NewStackF(app, pjPrefix+"-Stack-F", &StackFProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(pjPrefix + "-Stack-F"),
		},
		PjPrefix: pjPrefix,
	})
	stackF.AddDependency(stackC, jsii.String("Stack F depends on Stack C"))
	stackF.AddDependency(stackD, jsii.String("Stack F depends on Stack D"))
	stackF.AddDependency(stackE, jsii.String("Stack F depends on Stack E"))

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
