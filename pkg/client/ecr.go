//go:generate mockgen -source=./ecr.go -destination=./ecr_mock.go -package=client
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

type Ecr struct {
	client *ecr.Client
}

func NewEcr(client *ecr.Client) *Ecr {
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
	if err != nil {
		return &ClientError{
			ResourceName: repositoryName,
			Err:          err,
		}
	}
	return nil
}

func (e *Ecr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	var nextToken *string

	for {
		select {
		case <-ctx.Done():
			return false, &ClientError{
				ResourceName: repositoryName,
				Err:          ctx.Err(),
			}
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
			return false, &ClientError{
				ResourceName: repositoryName,
				Err:          err,
			}
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
