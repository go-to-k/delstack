package resource

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewLambdaVpcAttached creates a Lambda function attached to VPC
func NewLambdaVpcAttached(scope constructs.Construct, pjPrefix string, vpc awsec2.Vpc, ipv6Enabled bool) awslambda.Function {
	var functionName string
	if ipv6Enabled {
		functionName = pjPrefix + "-LambdaVpcIpv6"
	} else {
		functionName = pjPrefix + "-LambdaVpc"
	}

	functionProps := &awslambda.FunctionProps{
		FunctionName: jsii.String(functionName),
		Runtime:      awslambda.Runtime_PYTHON_3_12(),
		Handler:      jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(`
def handler(event, context):
    return {
        'statusCode': 200,
        'body': 'Hello from VPC Lambda!'
    }
`)),
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
		Vpc:     vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	}

	if ipv6Enabled {
		functionProps.Ipv6AllowedForDualStack = jsii.Bool(true)
	}

	function := awslambda.NewFunction(scope, jsii.String(functionName), functionProps)

	return function
}

// NewLambdaNoVpc creates a Lambda function without VPC
func NewLambdaNoVpc(scope constructs.Construct, pjPrefix string) awslambda.Function {
	functionName := pjPrefix + "-LambdaNoVpc"

	function := awslambda.NewFunction(scope, jsii.String(functionName), &awslambda.FunctionProps{
		FunctionName: jsii.String(functionName),
		Runtime:      awslambda.Runtime_PYTHON_3_12(),
		Handler:      jsii.String("index.handler"),
		Code: awslambda.Code_FromInline(jsii.String(`
def handler(event, context):
    return {
        'statusCode': 200,
        'body': 'Hello from non-VPC Lambda!'
    }
`)),
		Timeout: awscdk.Duration_Seconds(jsii.Number(30)),
	})

	return function
}
