package operation

import (
	"context"
	"fmt"

	"github.com/go-to-k/delstack/pkg/client"
)

var _ client.IEcr = (*MockEcr)(nil)
var _ client.IEcr = (*DeleteRepositoryErrorMockEcr)(nil)
var _ client.IEcr = (*CheckEcrExistsErrorMockEcr)(nil)
var _ client.IEcr = (*CheckEcrNotExistsMockEcr)(nil)

/*
	Mocks for client
*/

type MockEcr struct{}

func NewMockEcr() *MockEcr {
	return &MockEcr{}
}

func (m *MockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return nil
}

func (m *MockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type DeleteRepositoryErrorMockEcr struct{}

func NewDeleteRepositoryErrorMockEcr() *DeleteRepositoryErrorMockEcr {
	return &DeleteRepositoryErrorMockEcr{}
}

func (m *DeleteRepositoryErrorMockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return fmt.Errorf("DeleteRepositoryError")
}

func (m *DeleteRepositoryErrorMockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type CheckEcrExistsErrorMockEcr struct{}

func NewCheckEcrExistsErrorMockEcr() *CheckEcrExistsErrorMockEcr {
	return &CheckEcrExistsErrorMockEcr{}
}

func (m *CheckEcrExistsErrorMockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return nil
}

func (m *CheckEcrExistsErrorMockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, fmt.Errorf("DescribeRepositoriesError")
}

type CheckEcrNotExistsMockEcr struct{}

func NewCheckEcrNotExistsMockEcr() *CheckEcrNotExistsMockEcr {
	return &CheckEcrNotExistsMockEcr{}
}

func (m *CheckEcrNotExistsMockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return nil
}

func (m *CheckEcrNotExistsMockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, nil
}
