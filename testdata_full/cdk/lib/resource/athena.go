package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsathena"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewAthena(scope constructs.Construct, resourcePrefix string) {
	awsathena.NewCfnWorkGroup(scope, jsii.String("AthenaWorkGroup"), &awsathena.CfnWorkGroupProps{
		Name:  jsii.String(resourcePrefix + "-AthenaWorkGroup"),
		State: jsii.String("ENABLED"),
	})
}
