package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/go-to-k/delstack/client"
)

type StackOperatorFactory struct {
	config aws.Config
}

func NewStackOperatorFactory(config aws.Config) *StackOperatorFactory {
	return &StackOperatorFactory{config}
}

func (factory *StackOperatorFactory) CreateStackOperator() Operator {
	return NewStackOperator(
		factory.config,
		factory.createCloudFormationClient(),
	)
}

func (factory *StackOperatorFactory) createCloudFormationClient() client.ICloudFormation {
	sdkCfnClient := cloudformation.NewFromConfig(factory.config)
	sdkCfnWaiter := cloudformation.NewStackDeleteCompleteWaiter(sdkCfnClient)

	return client.NewCloudFormation(
		factory.config,
		sdkCfnClient,
		sdkCfnWaiter,
	)
}
