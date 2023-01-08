package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
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
var _ IOperator = (*MockStackOperator)(nil)

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

	stackOperator := NewMockStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrOperator := NewMockEcrOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operators = append(operators, stackOperator)
	operators = append(operators, bucketOperator)
	operators = append(operators, roleOperator)
	operators = append(operators, ecrOperator)
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

	stackOperator := NewMockStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrOperator := NewMockEcrOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operators = append(operators, stackOperator)
	operators = append(operators, bucketOperator)
	operators = append(operators, roleOperator)
	operators = append(operators, ecrOperator)
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

	stackOperator := NewErrorMockStackOperator()
	bucketOperator := NewMockBucketOperator()
	roleOperator := NewMockRoleOperator()
	ecrOperator := NewMockEcrOperator()
	backupVaultOperator := NewMockBackupVaultOperator()
	customOperator := NewMockCustomOperator()

	operators = append(operators, stackOperator)
	operators = append(operators, bucketOperator)
	operators = append(operators, roleOperator)
	operators = append(operators, ecrOperator)
	operators = append(operators, backupVaultOperator)
	operators = append(operators, customOperator)

	return operators
}

func (m *OperatorDeleteResourcesMockOperatorCollection) RaiseUnsupportedResourceError() error {
	return fmt.Errorf("UnsupportedResourceError")
}

/*
	Test Cases
*/
func TestOperatorManager_getOperatorResourcesLength(t *testing.T) {
	io.NewLogger(false)

	mock := NewMockOperatorCollection()

	type args struct {
		ctx  context.Context
		mock IOperatorCollection
	}

	cases := []struct {
		name string
		args args
		want int
	}{
		{
			name: "get operator resources length successfully",
			args: args{
				ctx:  context.Background(),
				mock: mock,
			},
			want: 6,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			got := operatorManager.getOperatorResourcesLength()
			if got != tt.want {
				t.Errorf("got = %#v, want %#v", got, tt.want)
				return
			}
		})
	}
}

func TestOperatorManager_CheckResourceCounts(t *testing.T) {
	io.NewLogger(false)

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
				ctx:  context.Background(),
				mock: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "check resource counts failure",
			args: args{
				ctx:  context.Background(),
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

func TestOperatorManager_DeleteResourceCollection(t *testing.T) {
	io.NewLogger(false)

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
				ctx:  context.Background(),
				mock: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resource collection failure",
			args: args{
				ctx:  context.Background(),
				mock: operatorDeleteResourcesMock,
			},
			want:    fmt.Errorf("ErrorDeleteResources"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			operatorManager := NewOperatorManager(tt.args.mock)

			err := operatorManager.DeleteResourceCollection(tt.args.ctx)
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
