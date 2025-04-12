package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Define properties for ChildStack object
type ChildStackProps struct {
	awscdk.NestedStackProps
	PjPrefix string
}

// NewChildStack creates a child stack
func NewChildStack(scope constructs.Construct, id string, props *ChildStackProps) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	// Create nested stack
	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	// Get project prefix
	pjPrefix := props.PjPrefix

	// Create further nested stacks
	descendStack := NewDescendStack(stack, "DescendStack", &DescendStackProps{
		PjPrefix: pjPrefix,
	})

	descendThreeStack := NewDescendStack3(stack, "DescendThreeStack", &DescendStack3Props{
		PjPrefix: pjPrefix,
	})

	// Create resources within the nested stack

	// Create S3 bucket
	childS3Bucket := awss3.NewBucket(stack, jsii.String("ChildS3Bucket"), &awss3.BucketProps{
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Versioned:         jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	})

	// Create IAM group
	childIamGroup := awsiam.NewGroup(stack, jsii.String("ChildIamGroup"), &awsiam.GroupProps{
		GroupName: jsii.String(pjPrefix + "-child-group"),
	})

	// Create IAM roles
	childLambdaRole := awsiam.NewRole(stack, jsii.String("ChildLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-child-role-1"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	childLambdaRole2 := awsiam.NewRole(stack, jsii.String("ChildLambdaRole2"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-child-role-2"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	// Create IAM policy and attach to both roles
	childLambdaPolicy := awsiam.NewPolicy(stack, jsii.String("ChildLambdaPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String(pjPrefix + "-child-policy"),
		Roles:      &[]awsiam.IRole{childLambdaRole, childLambdaRole2},
		Document: awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
			Statements: &[]awsiam.PolicyStatement{
				awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
					Effect: awsiam.Effect_ALLOW,
					Actions: jsii.Strings(
						"logs:CreateLogGroup",
						"logs:CreateLogStream",
						"logs:PutLogEvents",
						"logs:PutResourcePolicy",
						"logs:DeleteResourcePolicy",
					),
					Resources: jsii.Strings("*"),
				}),
			},
		}),
	})

	// Add outputs
	awscdk.NewCfnOutput(stack, jsii.String("ChildS3BucketName"), &awscdk.CfnOutputProps{
		Value: childS3Bucket.BucketName(),
	})

	// Avoid unused variable warnings (use appropriately in actual code)
	_ = childIamGroup
	_ = childLambdaPolicy
	_ = descendStack
	_ = descendThreeStack

	return stack
}
