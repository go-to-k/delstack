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
	resource.NewS3TableBucket(stack, props.PjPrefix+"-Descend") // can only contain [2 AWS::S3Tables::TableBucket] : Table bucket can only have up to 10 buckets created per AWS account (per region), and we want to be able to make up to 5 stacks
	resource.NewIamGroup(stack)                                 // can only contain [2 AWS::IAM::Group] in this CDK app: 1 IAM user (DelstackTestUser) can only belong to 10 IAM groups, and we want to be able to make up to 5 stacks

	return stack
}
