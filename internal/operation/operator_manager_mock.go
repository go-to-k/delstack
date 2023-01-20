package operation

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
	Mocks for OperatorManager
*/

var _ IOperatorManager = (*MockOperatorManager)(nil)
var _ IOperatorManager = (*AllErrorMockOperatorManager)(nil)
var _ IOperatorManager = (*CheckResourceCountsErrorMockOperatorManager)(nil)
var _ IOperatorManager = (*DeleteResourceCollectionErrorMockOperatorManager)(nil)

type MockOperatorManager struct{}

func NewMockOperatorManager() *MockOperatorManager {
	return &MockOperatorManager{}
}

func (m *MockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *MockOperatorManager) CheckResourceCounts() error {
	return nil
}

func (m *MockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *MockOperatorManager) DeleteResourceCollection(ctx context.Context) error {
	return nil
}

type AllErrorMockOperatorManager struct{}

func NewAllErrorMockOperatorManager() *AllErrorMockOperatorManager {
	return &AllErrorMockOperatorManager{}
}

func (m *AllErrorMockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *AllErrorMockOperatorManager) CheckResourceCounts() error {
	return fmt.Errorf("CheckResourceCountsError")
}

func (m *AllErrorMockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *AllErrorMockOperatorManager) DeleteResourceCollection(ctx context.Context) error {
	return fmt.Errorf("DeleteResourceCollectionError")
}

type CheckResourceCountsErrorMockOperatorManager struct{}

func NewCheckResourceCountsErrorMockOperatorManager() *CheckResourceCountsErrorMockOperatorManager {
	return &CheckResourceCountsErrorMockOperatorManager{}
}

func (m *CheckResourceCountsErrorMockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *CheckResourceCountsErrorMockOperatorManager) CheckResourceCounts() error {
	return fmt.Errorf("CheckResourceCountsError")
}

func (m *CheckResourceCountsErrorMockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *CheckResourceCountsErrorMockOperatorManager) DeleteResourceCollection(ctx context.Context) error {
	return nil
}

type DeleteResourceCollectionErrorMockOperatorManager struct{}

func NewDeleteResourceCollectionErrorMockOperatorManager() *DeleteResourceCollectionErrorMockOperatorManager {
	return &DeleteResourceCollectionErrorMockOperatorManager{}
}

func (m *DeleteResourceCollectionErrorMockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *DeleteResourceCollectionErrorMockOperatorManager) CheckResourceCounts() error {
	return nil
}

func (m *DeleteResourceCollectionErrorMockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *DeleteResourceCollectionErrorMockOperatorManager) DeleteResourceCollection(ctx context.Context) error {
	return fmt.Errorf("DeleteResourceCollectionError")
}
