package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Define properties for DescendStack object
type DescendStackProps struct {
	awscdk.NestedStackProps
	PjPrefix string
}

// NewDescendStack creates a nested stack that is further nested within a child stack
func NewDescendStack(scope constructs.Construct, id string, props *DescendStackProps) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	// Create nested stack
	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	// Get project prefix
	pjPrefix := props.PjPrefix

	// Create resources within the nested stack

	// Create IAM role
	descendLambdaRole := awsiam.NewRole(stack, jsii.String("DescendLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-descend-role-1"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	// Add outputs
	awscdk.NewCfnOutput(stack, jsii.String("DescendRoleName"), &awscdk.CfnOutputProps{
		Value: descendLambdaRole.RoleName(),
	})

	return stack
}
