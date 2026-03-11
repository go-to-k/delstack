package preprocessor

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/go-to-k/delstack/internal/operation"
	"github.com/go-to-k/delstack/pkg/client"
)

func NewLambdaVPCDetacherFromConfig(config aws.Config) *LambdaVPCDetacher {
	sdkLambdaClient := lambda.NewFromConfig(config, func(o *lambda.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})
	sdkLambdaWaiter := lambda.NewFunctionUpdatedV2Waiter(sdkLambdaClient)

	sdkCfnClient := cloudformation.NewFromConfig(config, func(o *cloudformation.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})
	sdkCfnDeleteWaiter := cloudformation.NewStackDeleteCompleteWaiter(sdkCfnClient)
	sdkCfnUpdateWaiter := cloudformation.NewStackUpdateCompleteWaiter(sdkCfnClient)

	sdkEC2Client := ec2.NewFromConfig(config, func(o *ec2.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewLambdaVPCDetacher(
		client.NewLambdaClient(
			sdkLambdaClient,
			sdkLambdaWaiter,
		),
		client.NewCloudFormation(
			sdkCfnClient,
			sdkCfnDeleteWaiter,
			sdkCfnUpdateWaiter,
		),
		client.NewEC2Client(sdkEC2Client),
	)
}
