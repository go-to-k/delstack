package preprocessor

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
)

func NewRecursivePreprocessorFromConfig(config aws.Config, forceMode bool) *RecursivePreprocessor {
	sdkCfnClient := cloudformation.NewFromConfig(config, func(o *cloudformation.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})
	sdkCfnDeleteWaiter := cloudformation.NewStackDeleteCompleteWaiter(sdkCfnClient)
	sdkCfnUpdateWaiter := cloudformation.NewStackUpdateCompleteWaiter(sdkCfnClient)
	cfnClient := client.NewCloudFormation(
		sdkCfnClient,
		sdkCfnDeleteWaiter,
		sdkCfnUpdateWaiter,
	)

	lambdaVPCDetacher := newLambdaVPCDetacherFromConfig(config)
	sdkRoute53Client := route53.NewFromConfig(config, func(o *route53.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})
	route53Client := client.NewRoute53(sdkRoute53Client)

	protectionRemover := newDeletionProtectionRemoverFromConfig(config, forceMode, route53Client)
	recordSetProtectionRemover := NewRoute53RecordSetProtectionRemover(forceMode, cfnClient, route53Client)

	composite := NewCompositePreprocessor(
		[]IPreprocessor{protectionRemover, recordSetProtectionRemover},
		[]IPreprocessor{lambdaVPCDetacher},
	)

	return NewRecursivePreprocessor(cfnClient, composite)
}

func newLambdaVPCDetacherFromConfig(config aws.Config) *LambdaVPCDetacher {
	sdkLambdaClient := lambda.NewFromConfig(config, func(o *lambda.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})
	sdkLambdaWaiter := lambda.NewFunctionUpdatedV2Waiter(sdkLambdaClient)

	sdkEC2Client := ec2.NewFromConfig(config, func(o *ec2.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewLambdaVPCDetacher(
		client.NewLambdaClient(
			sdkLambdaClient,
			sdkLambdaWaiter,
		),
		client.NewEC2Client(sdkEC2Client),
	)
}

func newDeletionProtectionRemoverFromConfig(config aws.Config, forceMode bool, route53Client client.IRoute53) *DeletionProtectionRemover {
	sdkEC2Client := ec2.NewFromConfig(config, func(o *ec2.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	sdkRDSClient := rds.NewFromConfig(config, func(o *rds.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	sdkCognitoClient := cognitoidentityprovider.NewFromConfig(config, func(o *cognitoidentityprovider.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	sdkLogsClient := cloudwatchlogs.NewFromConfig(config, func(o *cloudwatchlogs.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	sdkELBV2Client := elasticloadbalancingv2.NewFromConfig(config, func(o *elasticloadbalancingv2.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewDeletionProtectionRemover(
		forceMode,
		client.NewEC2Client(sdkEC2Client),
		client.NewRDS(sdkRDSClient),
		client.NewCognito(sdkCognitoClient),
		client.NewCloudWatchLogs(sdkLogsClient),
		client.NewELBV2(sdkELBV2Client),
		route53Client,
	)
}
