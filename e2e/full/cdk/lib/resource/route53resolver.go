package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53resolver"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// The VPC is deleted outside of CFn by deploy.go, so the ResolverConfig fails to delete (issue #682).
func NewRoute53ResolverConfig(scope constructs.Construct) {
	vpc := awsec2.NewCfnVPC(scope, jsii.String("ResolverConfigVpc"), &awsec2.CfnVPCProps{
		CidrBlock: jsii.String("10.0.0.0/16"),
	})

	awsroute53resolver.NewCfnResolverConfig(scope, jsii.String("ResolverConfig"), &awsroute53resolver.CfnResolverConfigProps{
		AutodefinedReverseFlag: jsii.String("DISABLE"),
		ResourceId:             vpc.Ref(),
	})
}
