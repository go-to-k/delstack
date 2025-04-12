package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"

	"cdk/lib" // Import custom library
)

type TestStackProps struct {
	awscdk.StackProps
	PjPrefix string
}

func NewTestStack(scope constructs.Construct, id string, props *TestStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Add project prefix as stack parameter
	pjPrefixParam := awscdk.NewCfnParameter(stack, jsii.String("PJPrefix"), &awscdk.CfnParameterProps{
		Type:        jsii.String("String"),
		Description: jsii.String("Project name prefix"),
		Default:     jsii.String(props.PjPrefix),
	})

	// Get parameter value
	pjPrefix := pjPrefixParam.ValueAsString()

	// Create ECR repositories
	ecrResources := lib.NewECRRepositories(stack, *pjPrefix)

	// Create S3 resources
	s3Resources := lib.NewS3Resources(stack, *pjPrefix)

	// Create IAM resources
	iamResources := lib.NewIAMResources(stack, *pjPrefix)

	// Create Lambda related resources
	lambdaResources := lib.NewLambdaResources(stack, *pjPrefix, iamResources)

	// Create DynamoDB resources
	dynamoResources := lib.NewDynamoDBResources(stack, *pjPrefix)

	// Create AWS Backup resources
	backupResources := lib.NewBackupResources(stack, *pjPrefix, iamResources)

	// Create first nested stack
	childStack := lib.NewChildStack(stack, "ChildStack", &lib.ChildStackProps{
		PjPrefix: *pjPrefix,
	})

	// Create second nested stack
	childStack2 := lib.NewChildStack2(stack, "ChildTwoStack", &lib.ChildStack2Props{
		PjPrefix: *pjPrefix,
	})

	// Add outputs
	awscdk.NewCfnOutput(stack, jsii.String("ECR1Arn"), &awscdk.CfnOutputProps{
		Value: ecrResources["ECR1"].(awsecr.Repository).RepositoryArn(),
	})

	awscdk.NewCfnOutput(stack, jsii.String("S3BucketName"), &awscdk.CfnOutputProps{
		Value: jsii.String(*pjPrefix + "-root--use1-az4--x-s3"),
	})

	// Use variables to avoid unused variable warnings (use appropriately in actual code)
	_ = s3Resources
	_ = lambdaResources
	_ = dynamoResources
	_ = backupResources
	_ = childStack
	_ = childStack2

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Get stack name from environment variables or use default value
	pjPrefix := os.Getenv("PJ_PREFIX")
	if pjPrefix == "" {
		pjPrefix = "dev-delstack-test"
	}

	stackName := pjPrefix + "-TestStack"

	NewTestStack(app, stackName, &TestStackProps{
		StackProps: awscdk.StackProps{
			Env:       env(),
			StackName: jsii.String(stackName),
		},
		PjPrefix: pjPrefix,
	})

	app.Synth(nil)
}

// Set up the environment (account+region)
func env() *awscdk.Environment {
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	region := os.Getenv("CDK_DEFAULT_REGION")

	if region == "" {
		region = "us-east-1" // Default region
	}

	if account != "" && region != "" {
		return &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String(region),
		}
	}

	// Use current settings if environment variables are not set
	return &awscdk.Environment{
		Account: jsii.String(account),
		Region:  jsii.String(region),
	}
}
