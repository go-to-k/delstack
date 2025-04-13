package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type ChildStack2Props struct {
	awscdk.NestedStackProps
	PjPrefix string
}

func NewChildStack2(scope constructs.Construct, id string, props *ChildStack2Props) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	NewS3Bucket(stack)
	NewCustomResource(stack)

	NewDescendStack2(stack, "DescendTwo", &DescendStack2Props{
		PjPrefix: props.PjPrefix,
	})

	return stack
}
