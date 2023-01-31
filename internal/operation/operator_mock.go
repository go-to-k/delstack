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
var _ IOperator = (*MockS3BucketOperator)(nil)
var _ IOperator = (*MockIamRoleOperator)(nil)
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

type MockS3BucketOperator struct{}

func NewMockS3BucketOperator() *MockS3BucketOperator {
	return &MockS3BucketOperator{}
}

func (m *MockS3BucketOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockS3BucketOperator) GetResourcesLength() int {
	return 1
}

func (m *MockS3BucketOperator) DeleteResources(ctx context.Context) error {
	return nil
}

type MockIamRoleOperator struct{}

func NewMockIamRoleOperator() *MockIamRoleOperator {
	return &MockIamRoleOperator{}
}

func (m *MockIamRoleOperator) AddResource(resource *types.StackResourceSummary) {}

func (m *MockIamRoleOperator) GetResourcesLength() int {
	return 1
}

func (m *MockIamRoleOperator) DeleteResources(ctx context.Context) error {
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
