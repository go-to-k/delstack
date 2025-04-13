package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewCustomResource(scope constructs.Construct) {
	logGroup := awslogs.NewLogGroup(scope, jsii.String("LogGroup"), &awslogs.LogGroupProps{
		Retention:     awslogs.RetentionDays_ONE_DAY,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	resourcePolicyLambda := awslambda.NewFunction(scope, jsii.String("ResourcePolicyLambdaForLogs"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PYTHON_3_13(),
		Handler: jsii.String("index.handler"),
		Code:    awslambda.Code_FromInline(jsii.String(getCode())),
	})

	provider := customresources.NewProvider(scope, jsii.String("CustomResourceProvider"), &customresources.ProviderProps{
		OnEventHandler: resourcePolicyLambda,
	})

	awscdk.NewCustomResource(scope, jsii.String("AddResourcePolicy"), &awscdk.CustomResourceProps{
		ServiceToken: provider.ServiceToken(),
		Properties: &map[string]interface{}{
			"CloudWatchLogsLogGroupArn": []interface{}{logGroup.LogGroupArn()},
			"PolicyName":                "ResourcePolicyForDNSLog",
			"ServiceName":               "route53.amazonaws.com",
			"ServiceTimeout":            "5",
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
}

func getLambdaCode() string {
	return `
import json
import cfnresponse
import boto3
from botocore.exceptions import ClientError

client = boto3.client("logs")

def PutPolicy(arns, policyname, service):
	arn_str = '","'.join(arns)
	arn = "[\"" + arn_str + "\"]"

	response = client.put_resource_policy(
		policyName=policyname,
		policyDocument="{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"" + service + "\"},\"Action\":[\"logs:CreateLogStream\",\"logs:PutLogEvents\"],\"Resource\":"+ arn + "}]}",
	)
	return

def DeletePolicy(policyname):
	response = client.delete_resource_policy(
		policyName=policyname
	)
	return

def handler(event, context):

	CloudWatchLogsLogGroupArns = event['ResourceProperties']['CloudWatchLogsLogGroupArn']
	PolicyName = event['ResourceProperties']['PolicyName']
	ServiceName = event['ResourceProperties']['ServiceName']

	responseData = {}

	try:
		if event['RequestType'] == "Delete":
			# DeletePolicy(PolicyName)
			responseData['Data'] = "FAILED"
			status=cfnresponse.FAILED
		if event['RequestType'] == "Create":
			# PutPolicy(CloudWatchLogsLogGroupArns, PolicyName, ServiceName)
			responseData['Data'] = "SUCCESS"
			status=cfnresponse.SUCCESS
	except ClientError as e:
		responseData['Data'] = "FAILED"
		status=cfnresponse.FAILED
		print("Unexpected error: %s" % e)

	cfnresponse.send(event, context, status, responseData, "CustomResourcePhysicalID")
`
}
