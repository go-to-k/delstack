package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type ECR struct {
	client *ecr.Client
}

func NewECR(config aws.Config) *ECR {
	client := ecr.NewFromConfig(config)
	return &ECR{
		client,
	}
}

func (ecrClient *ECR) DeleteRepository(repositoryName *string) error {
	input := &ecr.DeleteRepositoryInput{
		RepositoryName: repositoryName,
		Force:          true,
	}

	_, err := ecrClient.client.DeleteRepository(context.TODO(), input)

	return err
}
