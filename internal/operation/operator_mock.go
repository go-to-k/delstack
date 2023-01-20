package operation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
	Mocks for each Operator
*/

var _ IOperator = (*MockStackOperator)(nil)
var _ IOperator = (*ErrorMockStackOperator)(nil)
var _ IOperator = (*MockBucketOperator)(nil)
var _ IOperator = (*MockRoleOperator)(nil)
var _ IOperator = (*MockEcrOperator)(nil)
var _ IOperator = (*MockBackupVaultOperator)(nil)
var _ IOperator = (*MockCustomOperator)(nil)

type MockStackOperator struct{}

func NewMockStackOperator() *MockStackOperator {
	return &MockStackOperator{}
}

func (m *MockStackOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockStackOperator) GetResourcesLength() int {
	return 1
}

func (m *MockStackOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type ErrorMockStackOperator struct{}

func NewErrorMockStackOperator() *ErrorMockStackOperator {
	return &ErrorMockStackOperator{}
}

func (m *ErrorMockStackOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *ErrorMockStackOperator) GetResourcesLength() int {
	return 1
}

func (m *ErrorMockStackOperator) DeleteResources(ctx context.Context) error {
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

type MockEcrOperator struct{}

func NewMockEcrOperator() *MockEcrOperator {
	return &MockEcrOperator{}
}

func (m *MockEcrOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockEcrOperator) GetResourcesLength() int {
	return 1
}

func (m *MockEcrOperator) DeleteResources(ctx context.Context) error {
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
