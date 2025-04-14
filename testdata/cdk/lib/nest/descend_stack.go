package nest

import (
	"cdk/lib/resource"

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

	resource.NewS3DirectoryBucket(stack, props.PjPrefix+"-Descend")
	resource.NewIamGroup(stack)

	return stack
}
