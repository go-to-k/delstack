package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/logger"
)

/*
	Mocks for each Operator
*/
var _ IOperator = (*mockStackOperator)(nil)

type mockStackOperator struct{}

func NewMockStackOperator() *mockStackOperator {
	return &mockStackOperator{}
}

func (m *mockStackOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *mockStackOperator) GetResourcesLength() int {
	return 1
}

func (m *mockStackOperator) DeleteResources() error {
	return nil
}

type errorMockStackOperator struct{}

func NewErrorMockStackOperator() *errorMockStackOperator {
	return &errorMockStackOperator{}
}

func (m *errorMockStackOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *errorMockStackOperator) GetResourcesLength() int {
	return 1
}

func (m *errorMockStackOperator) DeleteResources() error {
	return fmt.Errorf("ErrorDeleteResources")
}

type mockBucketOperator struct{}

func NewMockBucketOperator() *mockBucketOperator {
	return &mockBucketOperator{}
}

func (m *mockBucketOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *mockBucketOperator) GetResourcesLength() int {
	return 1
}

func (m *mockBucketOperator) DeleteResources() error {
	return nil
}

type mockRoleOperator struct{}

func NewMockRoleOperator() *mockRoleOperator {
	return &mockRoleOperator{}
}

func (m *mockRoleOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *mockRoleOperator) GetResourcesLength() int {
	return 1
}

func (m *mockRoleOperator) DeleteResources() error {
	return nil
}

type mockEcrOperator struct{}

func NewMockEcrOperator() *mockEcrOperator {
	return &mockEcrOperator{}
}

func (m *mockEcrOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *mockEcrOperator) GetResourcesLength() int {
	return 1
}

func (m *mockEcrOperator) DeleteResources() error {
	return nil
}

type mockBackupVaultOperator struct{}

func NewMockBackupVaultOperator() *mockBackupVaultOperator {
	return &mockBackupVaultOperator{}
}

func (m *mockBackupVaultOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *mockBackupVaultOperator) GetResourcesLength() int {
	return 1
}

func (m *mockBackupVaultOperator) DeleteResources() error {
	return nil
}

type mockCustomOperator struct{}

func NewMockCustomOperator() *mockCustomOperator {
	return &mockCustomOperator{}
}

func (m *mockCustomOperator) AddResources(resource *types.StackResourceSummary) {}

func (m *mockCustomOperator) GetResourcesLength() int {
	return 1
}

func (m *mockCustomOperator) DeleteResources() error {
	return nil
}

/*
	Mocks for OperatorCollection
*/
var _ IOperatorCollection = (*mockOperatorCollection)(nil)
var _ IOperatorCollection = (*incorrectResourceCountsMockOperatorCollection)(nil)

type mockOperatorCollection struct{}

func NewMockOperatorCollection() *mockOperatorCollection {
	return &mockOperatorCollection{}
}

func (m *mockOperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *mockOperatorCollection) GetLogicalResourceIds() []string {
	return []string{
		"logicalResourceId1",
		"logicalResourceId2",
		"logicalResourceId3",
		"logicalResourceId4",
		"logicalResourceId5",
		"logicalResourceId6",
	}
}

func (m *mockOperatorCollection) GetOperatorList() []IOperator {
	var operatorList []IOperator

	stackOperator := NewMockStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrOperator := NewMockEcrOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operatorList = append(operatorList, stackOperator)
	operatorList = append(operatorList, bucketOperator)
	operatorList = append(operatorList, roleOperator)
	operatorList = append(operatorList, ecrOperator)
	operatorList = append(operatorList, backupVaultOperator)
	operatorList = append(operatorList, customOperator)

	return operatorList
}

func (m *mockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}

type incorrectResourceCountsMockOperatorCollection struct{}

func NewIncorrectResourceCountsMockOperatorCollection() *incorrectResourceCountsMockOperatorCollection {
	return &incorrectResourceCountsMockOperatorCollection{}
}

func (m *incorrectResourceCountsMockOperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *incorrectResourceCountsMockOperatorCollection) GetLogicalResourceIds() []string {
	return []string{
		"logicalResourceId1",
		"logicalResourceId2",
	}
}

func (m *incorrectResourceCountsMockOperatorCollection) GetOperatorList() []IOperator {
	var operatorList []IOperator

	stackOperator := NewMockStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrOperator := NewMockEcrOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operatorList = append(operatorList, stackOperator)
	operatorList = append(operatorList, bucketOperator)
	operatorList = append(operatorList, roleOperator)
	operatorList = append(operatorList, ecrOperator)
	operatorList = append(operatorList, backupVaultOperator)
	operatorList = append(operatorList, customOperator)

	return operatorList
}

func (m *incorrectResourceCountsMockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}

type operatorDeleteResourcesMockOperatorCollection struct{}

func NewOperatorDeleteResourcesMockOperatorCollection() *operatorDeleteResourcesMockOperatorCollection {
	return &operatorDeleteResourcesMockOperatorCollection{}
}

func (m *operatorDeleteResourcesMockOperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *operatorDeleteResourcesMockOperatorCollection) GetLogicalResourceIds() []string {
	return []string{
		"logicalResourceId1",
		"logicalResourceId2",
		"logicalResourceId3",
		"logicalResourceId4",
		"logicalResourceId5",
		"logicalResourceId6",
	}
}

func (m *operatorDeleteResourcesMockOperatorCollection) GetOperatorList() []IOperator {
	var operatorList []IOperator

	stackOperator := NewErrorMockStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrOperator := NewMockEcrOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operatorList = append(operatorList, stackOperator)
	operatorList = append(operatorList, bucketOperator)
	operatorList = append(operatorList, roleOperator)
	operatorList = append(operatorList, ecrOperator)
	operatorList = append(operatorList, backupVaultOperator)
	operatorList = append(operatorList, customOperator)

	return operatorList
}

func (m *operatorDeleteResourcesMockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}

/*
	Test Cases
*/
func TestCheckResourceCounts(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()

	mock := NewMockOperatorCollection()
	incorrectResourceCountsMock := NewIncorrectResourceCountsMockOperatorCollection()

	type args struct {
		ctx  context.Context
		mock IOperatorCollection
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "check resource counts successfully",
			args: args{
				ctx:  ctx,
				mock: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "check resource counts failure",
			args: args{
				ctx:  ctx,
				mock: incorrectResourceCountsMock,
			},
			want:    fmt.Errorf("UnsupportedResourceError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			err := operatorManager.CheckResourceCounts()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestDeleteResourceCollection(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()

	mock := NewMockOperatorCollection()
	operatorDeleteResourcesMock := NewOperatorDeleteResourcesMockOperatorCollection()

	type args struct {
		ctx  context.Context
		mock IOperatorCollection
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete resource collection successfully",
			args: args{
				ctx:  ctx,
				mock: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resource collection failure",
			args: args{
				ctx:  ctx,
				mock: operatorDeleteResourcesMock,
			},
			want:    fmt.Errorf("ErrorDeleteResources"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			err := operatorManager.DeleteResourceCollection()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
