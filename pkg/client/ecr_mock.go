package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

var _ IEcrSDKClient = (*MockEcrSDKClient)(nil)
var _ IEcrSDKClient = (*ErrorMockEcrSDKClient)(nil)
var _ IEcrSDKClient = (*NotExistsMockForDescribeRepositoriesEcrSDKClient)(nil)

/*
	Mocks for SDK Client
*/

type MockEcrSDKClient struct{}

func NewMockEcrSDKClient() *MockEcrSDKClient {
	return &MockEcrSDKClient{}
}

func (m *MockEcrSDKClient) DeleteRepository(ctx context.Context, params *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error) {
	return nil, nil
}

func (m *MockEcrSDKClient) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	output := &ecr.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			{
				RepositoryName: aws.String("test"),
			},
		},
	}
	return output, nil
}

type ErrorMockEcrSDKClient struct{}

func NewErrorMockEcrSDKClient() *ErrorMockEcrSDKClient {
	return &ErrorMockEcrSDKClient{}
}

func (m *ErrorMockEcrSDKClient) DeleteRepository(ctx context.Context, params *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error) {
	return nil, fmt.Errorf("DeleteRepositoryError")
}

func (m *ErrorMockEcrSDKClient) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	return nil, fmt.Errorf("DescribeRepositoriesError")
}

type NotExistsMockForDescribeRepositoriesEcrSDKClient struct{}

func NewNotExistsMockForDescribeRepositoriesEcrSDKClient() *NotExistsMockForDescribeRepositoriesEcrSDKClient {
	return &NotExistsMockForDescribeRepositoriesEcrSDKClient{}
}

func (m *NotExistsMockForDescribeRepositoriesEcrSDKClient) DeleteRepository(ctx context.Context, params *ecr.DeleteRepositoryInput, optFns ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error) {
	return nil, nil
}

func (m *NotExistsMockForDescribeRepositoriesEcrSDKClient) DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	return nil, fmt.Errorf("does not exist")
}
