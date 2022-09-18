package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-to-k/delstack/client"
)

type RoleOperatorFactory struct {
	config aws.Config
}

func NewRoleOperatorFactory(config aws.Config) *RoleOperatorFactory {
	return &RoleOperatorFactory{config}
}

func (factory *RoleOperatorFactory) CreateRoleOperator() Operator {
	return NewRoleOperator(
		factory.createIamClient(),
	)
}

func (factory *RoleOperatorFactory) createIamClient() client.IIam {
	sdkIamClient := iam.NewFromConfig(factory.config)

	return client.NewIam(
		factory.config,
		sdkIamClient,
	)
}
