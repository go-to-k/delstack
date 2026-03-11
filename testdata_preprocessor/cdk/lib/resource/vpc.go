package resource

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewVpc(scope constructs.Construct, pjPrefix string) awsec2.Vpc {
	vpc := awsec2.NewVpc(scope, jsii.String("Vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("10.193.0.0/16")),
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(0), // No NAT Gateway to reduce costs
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
	})

	// Enable IPv6 for the VPC to support dual-stack Lambda functions
	ipv6CidrBlock := awsec2.NewCfnVPCCidrBlock(scope, jsii.String("VpcIpv6CidrBlock"), &awsec2.CfnVPCCidrBlockProps{
		VpcId:                       vpc.VpcId(),
		AmazonProvidedIpv6CidrBlock: jsii.Bool(true),
	})

	// Add IPv6 CIDR blocks to all private subnets
	for i, subnet := range *vpc.IsolatedSubnets() {
		cidrBlock := awsec2.NewCfnSubnetCidrBlock(scope, jsii.String(fmt.Sprintf("PrivateSubnetIpv6CidrBlock%d", i)), &awsec2.CfnSubnetCidrBlockProps{
			SubnetId:      subnet.SubnetId(),
			Ipv6CidrBlock: awscdk.Fn_Select(jsii.Number(i), awscdk.Fn_Cidr(awscdk.Fn_Select(jsii.Number(0), vpc.VpcIpv6CidrBlocks()), jsii.Number(256), jsii.String("64"))),
		})
		cidrBlock.Node().AddDependency(ipv6CidrBlock)
	}

	return vpc
}
