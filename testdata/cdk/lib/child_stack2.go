package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Define properties for ChildStack2 object
type ChildStack2Props struct {
	awscdk.NestedStackProps
	PjPrefix string
}

// NewChildStack2 creates a second child stack
func NewChildStack2(scope constructs.Construct, id string, props *ChildStack2Props) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	// Create nested stack
	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	// Get project prefix
	pjPrefix := props.PjPrefix

	// Create further nested stack
	descendTwoStack := NewDescendStack2(stack, "DescendTwoStack", &DescendStack2Props{
		PjPrefix: pjPrefix,
	})

	// Create resources within the nested stack

	// Create S3 bucket
	childTwoS3Bucket := awss3.NewBucket(stack, jsii.String("ChildTwoS3Bucket"), &awss3.BucketProps{
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	})

	// Create IAM roles
	childTwoLambdaRole := awsiam.NewRole(stack, jsii.String("ChildTwoLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-child2-role-1"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	childTwoLambdaRole2 := awsiam.NewRole(stack, jsii.String("ChildTwoLambdaRole2"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-child2-role-2"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	// Create IAM policy and attach to both roles
	childTwoLambdaPolicy := awsiam.NewPolicy(stack, jsii.String("ChildTwoLambdaPolicy"), &awsiam.PolicyProps{
		PolicyName: jsii.String(pjPrefix + "-child2-policy"),
		Roles:      &[]awsiam.IRole{childTwoLambdaRole, childTwoLambdaRole2},
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
	awscdk.NewCfnOutput(stack, jsii.String("ChildTwoS3BucketName"), &awscdk.CfnOutputProps{
		Value: childTwoS3Bucket.BucketName(),
	})

	// Avoid unused variable warnings (use appropriately in actual code)
	_ = descendTwoStack
	_ = childTwoLambdaPolicy

	return stack
}
