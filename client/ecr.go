package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type IEcr interface {
	DeleteRepository(repositoryName *string) error
}

var _ IEcr = (*Ecr)(nil)

type IEcrSDKClient interface {
	DeleteRepository(ctx context.Context, params *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error)
}

type Ecr struct {
	client IEcrSDKClient
}

func NewEcr(client IEcrSDKClient) *Ecr {
	return &Ecr{
		client,
	}
}

func (ecrClient *Ecr) DeleteRepository(repositoryName *string) error {
	input := &ecr.DeleteRepositoryInput{
		RepositoryName: repositoryName,
		Force:          true,
	}

	_, err := ecrClient.client.DeleteRepository(context.TODO(), input)

	return err
}
