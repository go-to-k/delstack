package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewAlb creates an ELBv2 Application Load Balancer with deletion protection enabled.
func NewAlb(scope constructs.Construct, vpc awsec2.Vpc) awselasticloadbalancingv2.ApplicationLoadBalancer {
	alb := awselasticloadbalancingv2.NewApplicationLoadBalancer(scope, jsii.String("Alb"), &awselasticloadbalancingv2.ApplicationLoadBalancerProps{
		Vpc: vpc,
		InternetFacing:   jsii.Bool(true),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PUBLIC,
		},
		// Enable deletion protection
		DeletionProtection: jsii.Bool(true),
	})

	return alb
}
