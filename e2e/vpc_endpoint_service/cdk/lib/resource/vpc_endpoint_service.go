package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// VpcEndpointServiceResources is the provider side of the topology: a VPC, an
// internal Network Load Balancer, and the AWS PrivateLink endpoint service hosted
// on it. The consumer VPC endpoint that actually blocks the deletion is created
// out of band by deploy.go, because a consumer endpoint inside this stack would be
// deleted by CloudFormation in the correct order and never block anything.
//
// AcceptanceRequired is false so the connection reaches `Available` on its own,
// and the deploying account is an allowed principal so it can connect to its own
// endpoint service.
type VpcEndpointServiceResources struct {
	Vpc                 awsec2.Vpc
	NetworkLoadBalancer awselasticloadbalancingv2.NetworkLoadBalancer
	EndpointService     awsec2.VpcEndpointService
}

func NewVpcEndpointServiceStack(scope constructs.Construct) VpcEndpointServiceResources {
	vpc := awsec2.NewVpc(scope, jsii.String("Vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("10.194.0.0/16")),
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(0),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// An endpoint service needs a load balancer, but no listener or target is
	// required for a connection to reach the `Available` state.
	nlb := awselasticloadbalancingv2.NewNetworkLoadBalancer(scope, jsii.String("Nlb"), &awselasticloadbalancingv2.NetworkLoadBalancerProps{
		Vpc:            vpc,
		InternetFacing: jsii.Bool(false),
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	})

	accountRootArn := awscdk.Fn_Sub(jsii.String("arn:${AWS::Partition}:iam::${AWS::AccountId}:root"), nil)

	endpointService := awsec2.NewVpcEndpointService(scope, jsii.String("EndpointService"), &awsec2.VpcEndpointServiceProps{
		VpcEndpointServiceLoadBalancers: &[]awsec2.IVpcEndpointServiceLoadBalancer{nlb},
		AcceptanceRequired:              jsii.Bool(false),
		AllowedPrincipals: &[]awsiam.ArnPrincipal{
			awsiam.NewArnPrincipal(accountRootArn),
		},
	})

	return VpcEndpointServiceResources{
		Vpc:                 vpc,
		NetworkLoadBalancer: nlb,
		EndpointService:     endpointService,
	}
}
