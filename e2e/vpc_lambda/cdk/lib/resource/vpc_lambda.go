package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewVpcLambdaStack creates a minimal VPC + VPC-attached Lambda topology used to
// reproduce orphan AWS Lambda VPC ENIs blocking Subnet/SecurityGroup deletion.
// PRIVATE_ISOLATED + 0 NAT Gateways keeps the cost low; Lambda VPC ENIs are still
// provisioned by AWS Lambda when the function is invoked.
func NewVpcLambdaStack(scope constructs.Construct, pjPrefix string) awslambda.Function {
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
		Description:      jsii.String("Security group for VPC Lambda used in orphan ENI E2E"),
		AllowAllOutbound: jsii.Bool(true),
	})

	functionName := pjPrefix + "-VpcLambda"

	return awslambda.NewFunction(scope, jsii.String("VpcLambda"), &awslambda.FunctionProps{
		FunctionName: jsii.String(functionName),
		Runtime:      awslambda.Runtime_PYTHON_3_12(),
		Handler:      jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(`
def handler(event, context):
    return {'statusCode': 200, 'body': 'ok'}
`)),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Vpc:     vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups: &[]awsec2.ISecurityGroup{lambdaSg},
	})
}
