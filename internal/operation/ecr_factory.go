package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/go-to-k/delstack/pkg/client"
)

type EcrOperatorFactory struct {
	config aws.Config
}

func NewEcrOperatorFactory(config aws.Config) *EcrOperatorFactory {
	return &EcrOperatorFactory{config}
}

func (f *EcrOperatorFactory) CreateEcrOperator() *EcrOperator {
	return NewEcrOperator(
		f.createEcrClient(),
	)
}

func (f *EcrOperatorFactory) createEcrClient() *client.Ecr {
	sdkEcrClient := ecr.NewFromConfig(f.config)

	return client.NewEcr(
		sdkEcrClient,
	)
}
