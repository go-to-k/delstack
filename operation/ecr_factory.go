package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/go-to-k/delstack/client"
)

type EcrOperatorFactory struct {
	config aws.Config
}

func NewEcrOperatorFactory(config aws.Config) *EcrOperatorFactory {
	return &EcrOperatorFactory{config}
}

func (factory *EcrOperatorFactory) CreateEcrOperator() *EcrOperator {
	return NewEcrOperator(
		factory.createEcrClient(),
	)
}

func (factory *EcrOperatorFactory) createEcrClient() *client.Ecr {
	sdkEcrClient := ecr.NewFromConfig(factory.config)

	return client.NewEcr(
		factory.config,
		sdkEcrClient,
	)
}
