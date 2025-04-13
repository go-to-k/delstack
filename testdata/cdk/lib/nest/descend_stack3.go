package nest

import (
	"cdk/lib/resource"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type DescendStack3Props struct {
	awscdk.NestedStackProps
	PjPrefix string
}

func NewDescendStack3(scope constructs.Construct, id string, props *DescendStack3Props) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}

	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	resource.NewCustomResources(stack)
	resource.NewDynamoDB(stack)

	return stack
}
