package client

import (
	"context"
	"log"

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

func (ecrClient *ECR) DeleteRepository(repositoryName string) error {
	input := &ecr.DeleteRepositoryInput{
		RepositoryName: &repositoryName,
		Force:          true,
	}

	_, err := ecrClient.client.DeleteRepository(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed delete the ECR Repository, %v", err)
		return err
	}

	return nil
}
