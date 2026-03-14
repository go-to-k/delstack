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
	protectionRemover := newDeletionProtectionRemoverFromConfig(config, forceMode)

	composite := NewCompositePreprocessor(
		[]IPreprocessor{protectionRemover},
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

func newDeletionProtectionRemoverFromConfig(config aws.Config, forceMode bool) *DeletionProtectionRemover {
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
	)
}
