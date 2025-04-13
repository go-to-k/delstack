package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewIamGroup(scope constructs.Construct) {
	awsiam.NewGroup(scope, jsii.String("IamGroup"), &awsiam.GroupProps{})
}
