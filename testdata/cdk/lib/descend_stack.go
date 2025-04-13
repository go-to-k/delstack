package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type DescendStackProps struct {
	awscdk.NestedStackProps
	PjPrefix string
}

func NewDescendStack(scope constructs.Construct, id string, props *DescendStackProps) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	NewS3DirectoryBucket(stack, props.PjPrefix+"-Descend")
	NewIamGroup(stack)

	return stack
}
