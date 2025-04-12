package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Define properties for DescendStack3 object
type DescendStack3Props struct {
	awscdk.NestedStackProps
	PjPrefix string
}

// NewDescendStack3 creates a third nested stack that is nested within a child stack
func NewDescendStack3(scope constructs.Construct, id string, props *DescendStack3Props) awscdk.NestedStack {
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
	descendThreeLambdaRole := awsiam.NewRole(stack, jsii.String("DescendThreeLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-descend3-role-1"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	// Add outputs
	awscdk.NewCfnOutput(stack, jsii.String("DescendThreeRoleName"), &awscdk.CfnOutputProps{
		Value: descendThreeLambdaRole.RoleName(),
	})

	return stack
}
