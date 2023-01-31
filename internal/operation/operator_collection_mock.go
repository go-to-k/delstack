package operation

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
	Mocks for OperatorCollection
*/

var _ IOperatorCollection = (*MockOperatorCollection)(nil)
var _ IOperatorCollection = (*IncorrectResourceCountsMockOperatorCollection)(nil)
var _ IOperatorCollection = (*OperatorDeleteResourcesMockOperatorCollection)(nil)

type MockOperatorCollection struct{}

func NewMockOperatorCollection() *MockOperatorCollection {
	return &MockOperatorCollection{}
}

func (m *MockOperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *MockOperatorCollection) GetLogicalResourceIds() []string {
	return []string{
		"logicalResourceId1",
		"logicalResourceId2",
		"logicalResourceId3",
		"logicalResourceId4",
		"logicalResourceId5",
		"logicalResourceId6",
	}
}

func (m *MockOperatorCollection) GetOperators() []IOperator {
	var operators []IOperator

	cloudFormationStackOperator := NewMockCloudFormationStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrRepositoryOperator := NewMockEcrRepositoryOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operators = append(operators, cloudFormationStackOperator)
	operators = append(operators, bucketOperator)
	operators = append(operators, roleOperator)
	operators = append(operators, ecrRepositoryOperator)
	operators = append(operators, backupVaultOperator)
	operators = append(operators, customOperator)

	return operators
}

func (m *MockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}

type IncorrectResourceCountsMockOperatorCollection struct{}

func NewIncorrectResourceCountsMockOperatorCollection() *IncorrectResourceCountsMockOperatorCollection {
	return &IncorrectResourceCountsMockOperatorCollection{}
}

func (m *IncorrectResourceCountsMockOperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *IncorrectResourceCountsMockOperatorCollection) GetLogicalResourceIds() []string {
	return []string{
		"logicalResourceId1",
		"logicalResourceId2",
	}
}

func (m *IncorrectResourceCountsMockOperatorCollection) GetOperators() []IOperator {
	var operators []IOperator

	cloudFormationStackOperator := NewMockCloudFormationStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrRepositoryOperator := NewMockEcrRepositoryOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operators = append(operators, cloudFormationStackOperator)
	operators = append(operators, bucketOperator)
	operators = append(operators, roleOperator)
	operators = append(operators, ecrRepositoryOperator)
	operators = append(operators, backupVaultOperator)
	operators = append(operators, customOperator)

	return operators
}

func (m *IncorrectResourceCountsMockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}

type OperatorDeleteResourcesMockOperatorCollection struct{}

func NewOperatorDeleteResourcesMockOperatorCollection() *OperatorDeleteResourcesMockOperatorCollection {
	return &OperatorDeleteResourcesMockOperatorCollection{}
}

func (m *OperatorDeleteResourcesMockOperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *OperatorDeleteResourcesMockOperatorCollection) GetLogicalResourceIds() []string {
	return []string{
		"logicalResourceId1",
		"logicalResourceId2",
		"logicalResourceId3",
		"logicalResourceId4",
		"logicalResourceId5",
		"logicalResourceId6",
	}
}

func (m *OperatorDeleteResourcesMockOperatorCollection) GetOperators() []IOperator {
	var operators []IOperator

	cloudFormationStackOperator := NewErrorMockCloudFormationStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrRepositoryOperator := NewMockEcrRepositoryOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operators = append(operators, cloudFormationStackOperator)
	operators = append(operators, bucketOperator)
	operators = append(operators, roleOperator)
	operators = append(operators, ecrRepositoryOperator)
	operators = append(operators, backupVaultOperator)
	operators = append(operators, customOperator)

	return operators
}

func (m *OperatorDeleteResourcesMockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}
