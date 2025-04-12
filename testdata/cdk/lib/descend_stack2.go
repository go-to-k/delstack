package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Define properties for DescendStack2 object
type DescendStack2Props struct {
	awscdk.NestedStackProps
	PjPrefix string
}

// NewDescendStack2 creates a nested stack that is nested within ChildTwoStack
func NewDescendStack2(scope constructs.Construct, id string, props *DescendStack2Props) awscdk.NestedStack {
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
	descendTwoLambdaRole := awsiam.NewRole(stack, jsii.String("DescendTwoLambdaRole"), &awsiam.RoleProps{
		RoleName:  jsii.String(pjPrefix + "-descend2-role-1"),
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})

	// Add outputs
	awscdk.NewCfnOutput(stack, jsii.String("DescendTwoRoleName"), &awscdk.CfnOutputProps{
		Value: descendTwoLambdaRole.RoleName(),
	})

	return stack
}
