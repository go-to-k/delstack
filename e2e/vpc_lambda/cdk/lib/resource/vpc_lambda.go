package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// VpcLambdaResources is the topology required to reproduce issue #637 in
// E2E. We do NOT deploy a real VPC Lambda function: Hyperplane ENI release
// timing is non-deterministic, which makes the orphan-ENI condition flaky.
// Instead, deploy.go creates synthetic ENIs (description prefix
// "AWS Lambda VPC ENI-...") attached to this Subnet / SecurityGroup so the
// CloudFormation Subnet / SecurityGroup deletion always fails with the same
// dependency error a real orphan Lambda VPC ENI would produce.
type VpcLambdaResources struct {
	Vpc           awsec2.Vpc
	LambdaSg      awsec2.SecurityGroup
	PrivateSubnet awsec2.ISubnet
}

func NewVpcLambdaStack(scope constructs.Construct) VpcLambdaResources {
	vpc := awsec2.NewVpc(scope, jsii.String("Vpc"), &awsec2.VpcProps{
		IpAddresses: awsec2.IpAddresses_Cidr(jsii.String("10.193.0.0/16")),
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

	lambdaSg := awsec2.NewSecurityGroup(scope, jsii.String("LambdaSg"), &awsec2.SecurityGroupProps{
		Vpc:              vpc,
		Description:      jsii.String("Security group used to host synthetic Lambda VPC ENIs in E2E"),
		AllowAllOutbound: jsii.Bool(true),
	})

	subnets := vpc.IsolatedSubnets()
	return VpcLambdaResources{
		Vpc:           vpc,
		LambdaSg:      lambdaSg,
		PrivateSubnet: (*subnets)[0],
	}
}
