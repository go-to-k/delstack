package client

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type IEcr interface {
	DeleteRepository(ctx context.Context, repositoryName *string) error
	CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error)
}

var _ IEcr = (*Ecr)(nil)

type IEcrSDKClient interface {
	DeleteRepository(ctx context.Context, params *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error)
	DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
}

type Ecr struct {
	client IEcrSDKClient
}

func NewEcr(client IEcrSDKClient) *Ecr {
	return &Ecr{
		client,
	}
}

func (e *Ecr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	input := &ecr.DeleteRepositoryInput{
		RepositoryName: repositoryName,
		Force:          true,
	}

	_, err := e.client.DeleteRepository(ctx, input)

	return err
}

func (e *Ecr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	var nextToken *string

	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		input := &ecr.DescribeRepositoriesInput{
			NextToken: nextToken,
			RepositoryNames: []string{
				*repositoryName,
			},
		}

		output, err := e.client.DescribeRepositories(ctx, input)
		if err != nil && strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		if err != nil {
			return false, err
		}

		for _, repository := range output.Repositories {
			if *repository.RepositoryName == *repositoryName {
				return true, nil
			}
		}

		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return false, nil
}
