package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/go-to-k/delstack/pkg/client"
)

type StackOperatorFactory struct {
	config aws.Config
}

func NewStackOperatorFactory(config aws.Config) *StackOperatorFactory {
	return &StackOperatorFactory{config}
}

func (f *StackOperatorFactory) CreateStackOperator(targetResourceTypes []string) *StackOperator {
	return NewStackOperator(
		f.config,
		f.createCloudFormationClient(),
		targetResourceTypes,
	)
}

func (f *StackOperatorFactory) createCloudFormationClient() *client.CloudFormation {
	sdkCfnClient := cloudformation.NewFromConfig(f.config)
	sdkCfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(sdkCfnClient)

	return client.NewCloudFormation(
		sdkCfnClient,
		sdkCfnWaiter,
	)
}
