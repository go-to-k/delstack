package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/customresources"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NewLambdaResources creates required Lambda resources and CloudWatch log groups
func NewLambdaResources(scope constructs.Construct, pjPrefix string, iamResources map[string]awscdk.IResource) {
	// Create CloudWatch log group
	rootLogGroup := awslogs.NewLogGroup(scope, jsii.String("RootLogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String(pjPrefix + "-Root-log"),
		Retention:     awslogs.RetentionDays_TWO_WEEKS,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	// Get the Lambda role
	lambdaRole, ok := iamResources["RootLambdaRole"].(awsiam.Role)
	if !ok {
		panic("RootLambdaRole not found or not a Role")
	}

	// Create Lambda function
	lambdaCode := `
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

	rootResourcePolicyLambda := awslambda.NewFunction(scope, jsii.String("RootResourcePolicyLambdaForLogs"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PYTHON_3_9(),
		Handler:      jsii.String("index.handler"),
		Code:         awslambda.Code_FromInline(jsii.String(lambdaCode)),
		Role:         lambdaRole,
		FunctionName: jsii.String(pjPrefix + "-resource-policy-lambda"),
	})

	// Create custom resource
	provider := customresources.NewProvider(scope, jsii.String("RootCustomResourceProvider"), &customresources.ProviderProps{
		OnEventHandler: rootResourcePolicyLambda,
	})

	awscdk.NewCustomResource(scope, jsii.String("RootAddResourcePolicy"), &awscdk.CustomResourceProps{
		ServiceToken: provider.ServiceToken(),
		Properties: &map[string]interface{}{
			"CloudWatchLogsLogGroupArn": []interface{}{rootLogGroup.LogGroupArn()},
			"PolicyName":                pjPrefix + "RootResourcePolicyForDNSLog",
			"ServiceName":               "route53.amazonaws.com",
			"ServiceTimeout":            "5",
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
}
