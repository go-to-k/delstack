package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/go-to-k/delstack/pkg/client"
)

type RoleOperatorFactory struct {
	config aws.Config
}

func NewRoleOperatorFactory(config aws.Config) *RoleOperatorFactory {
	return &RoleOperatorFactory{config}
}

func (factory *RoleOperatorFactory) CreateRoleOperator() *RoleOperator {
	return NewRoleOperator(
		factory.createIamClient(),
	)
}

func (factory *RoleOperatorFactory) createIamClient() *client.Iam {
	sdkIamClient := iam.NewFromConfig(factory.config)

	return client.NewIam(
		sdkIamClient,
	)
}
