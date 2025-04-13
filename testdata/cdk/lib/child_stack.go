package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type ChildStackProps struct {
	awscdk.NestedStackProps
	PjPrefix string
}

func NewChildStack(scope constructs.Construct, id string, props *ChildStackProps) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	NewS3Bucket(stack)
	NewS3DirectoryBucket(stack, props.PjPrefix+"-Child")
	NewIamGroup(stack)
	NewCustomResources(stack)

	NewDescendStack(stack, "Descend", &DescendStackProps{
		PjPrefix: props.PjPrefix,
	})
	NewDescendStack3(stack, "DescendThree", &DescendStack3Props{
		PjPrefix: props.PjPrefix,
	})

	return stack
}
