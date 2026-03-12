package preprocessor

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
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

func NewAgentCoreVPCDetacherFromConfig(config aws.Config) *AgentCoreVPCDetacher {
	sdkAgentCoreClient := bedrockagentcorecontrol.NewFromConfig(config, func(o *bedrockagentcorecontrol.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	sdkEC2Client := ec2.NewFromConfig(config, func(o *ec2.Options) {
		o.RetryMaxAttempts = operation.SDKRetryMaxAttempts
		o.RetryMode = aws.RetryModeStandard
	})

	return NewAgentCoreVPCDetacher(
		client.NewAgentCoreClient(sdkAgentCoreClient),
		client.NewEC2Client(sdkEC2Client),
	)
}
