package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewIamUser(scope constructs.Construct) {
	awsiam.NewUser(scope, jsii.String("IamUser"), &awsiam.UserProps{})
}
