package operation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
	Mocks for each Operator
*/

var _ IOperator = (*MockCloudFormationStackOperator)(nil)
var _ IOperator = (*ErrorMockCloudFormationStackOperator)(nil)
var _ IOperator = (*MockBucketOperator)(nil)
var _ IOperator = (*MockRoleOperator)(nil)
var _ IOperator = (*MockEcrRepositoryOperator)(nil)
var _ IOperator = (*MockBackupVaultOperator)(nil)
var _ IOperator = (*MockCustomOperator)(nil)

type MockCloudFormationStackOperator struct{}

func NewMockCloudFormationStackOperator() *MockCloudFormationStackOperator {
	return &MockCloudFormationStackOperator{}
}

func (m *MockCloudFormationStackOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockCloudFormationStackOperator) GetResourcesLength() int {
	return 1
}

func (m *MockCloudFormationStackOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type ErrorMockCloudFormationStackOperator struct{}

func NewErrorMockCloudFormationStackOperator() *ErrorMockCloudFormationStackOperator {
	return &ErrorMockCloudFormationStackOperator{}
}

func (m *ErrorMockCloudFormationStackOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *ErrorMockCloudFormationStackOperator) GetResourcesLength() int {
	return 1
}

func (m *ErrorMockCloudFormationStackOperator) DeleteResources(ctx context.Context) error {
	return fmt.Errorf("ErrorDeleteResources")
}

type MockBucketOperator struct{}

func NewMockBucketOperator() *MockBucketOperator {
	return &MockBucketOperator{}
}

func (m *MockBucketOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockBucketOperator) GetResourcesLength() int {
	return 1
}

func (m *MockBucketOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type MockRoleOperator struct{}

func NewMockRoleOperator() *MockRoleOperator {
	return &MockRoleOperator{}
}

func (m *MockRoleOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockRoleOperator) GetResourcesLength() int {
	return 1
}

func (m *MockRoleOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type MockEcrRepositoryOperator struct{}

func NewMockEcrRepositoryOperator() *MockEcrRepositoryOperator {
	return &MockEcrRepositoryOperator{}
}

func (m *MockEcrRepositoryOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockEcrRepositoryOperator) GetResourcesLength() int {
	return 1
}

func (m *MockEcrRepositoryOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type MockBackupVaultOperator struct{}

func NewMockBackupVaultOperator() *MockBackupVaultOperator {
	return &MockBackupVaultOperator{}
}

func (m *MockBackupVaultOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockBackupVaultOperator) GetResourcesLength() int {
	return 1
}

func (m *MockBackupVaultOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type MockCustomOperator struct{}

func NewMockCustomOperator() *MockCustomOperator {
	return &MockCustomOperator{}
}

func (m *MockCustomOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockCustomOperator) GetResourcesLength() int {
	return 1
}

func (m *MockCustomOperator) DeleteResources(ctx context.Context) error {
	return nil
}
