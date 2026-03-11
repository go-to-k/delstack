package nest

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ChildStackProps struct {
	awscdk.NestedStackProps
	PjPrefix string
	Vpc      awsec2.Vpc
}

func NewChildStack(scope constructs.Construct, id string, props *ChildStackProps) awscdk.NestedStack {
	var sprops awscdk.NestedStackProps
	if props != nil {
		sprops = props.NestedStackProps
	}
	stack := awscdk.NewNestedStack(scope, &id, &sprops)

	// Lambda function in nested stack with VPC (IPv6 disabled)
	awslambda.NewFunction(stack, jsii.String("ChildLambdaVpc"), &awslambda.FunctionProps{
		FunctionName: jsii.String(props.PjPrefix + "-Child-LambdaVpc"),
		Runtime:      awslambda.Runtime_PYTHON_3_12(),
		Handler:      jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(`
def handler(event, context):
    return {
        'statusCode': 200,
        'body': 'Hello from child VPC Lambda!'
    }
`)),
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
		Vpc:     props.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	})

	// Lambda function in nested stack with VPC (IPv6 enabled)
	awslambda.NewFunction(stack, jsii.String("ChildLambdaVpcIpv6"), &awslambda.FunctionProps{
		FunctionName:             jsii.String(props.PjPrefix + "-Child-LambdaVpcIpv6"),
		Runtime:                  awslambda.Runtime_PYTHON_3_12(),
		Handler:                  jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(`
def handler(event, context):
    return {
        'statusCode': 200,
        'body': 'Hello from child VPC Lambda with IPv6!'
    }
`)),
		Timeout:                  awscdk.Duration_Seconds(jsii.Number(30)),
		Vpc:                      props.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		Ipv6AllowedForDualStack: jsii.Bool(true),
	})

	return stack
}
