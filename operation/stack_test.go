package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
)

/*
	Mocks for OperatorManager
*/
var _ IOperatorManager = (*mockOperatorManager)(nil)
var _ IOperatorManager = (*allErrorMockOperatorManager)(nil)
var _ IOperatorManager = (*checkResourceCountsErrorMockOperatorManager)(nil)
var _ IOperatorManager = (*deleteResourceCollectionErrorMockOperatorManager)(nil)

type mockOperatorManager struct{}

func NewMockOperatorManager() *mockOperatorManager {
	return &mockOperatorManager{}
}

func (m *mockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *mockOperatorManager) CheckResourceCounts() error {
	return nil
}

func (m *mockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *mockOperatorManager) DeleteResourceCollection() error {
	return nil
}

type allErrorMockOperatorManager struct{}

func NewAllErrorMockOperatorManager() *allErrorMockOperatorManager {
	return &allErrorMockOperatorManager{}
}

func (m *allErrorMockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *allErrorMockOperatorManager) CheckResourceCounts() error {
	return fmt.Errorf("CheckResourceCountsError")
}

func (m *allErrorMockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *allErrorMockOperatorManager) DeleteResourceCollection() error {
	return fmt.Errorf("DeleteResourceCollectionError")
}

type checkResourceCountsErrorMockOperatorManager struct{}

func NewCheckResourceCountsErrorMockOperatorManager() *checkResourceCountsErrorMockOperatorManager {
	return &checkResourceCountsErrorMockOperatorManager{}
}

func (m *checkResourceCountsErrorMockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *checkResourceCountsErrorMockOperatorManager) CheckResourceCounts() error {
	return fmt.Errorf("CheckResourceCountsError")
}

func (m *checkResourceCountsErrorMockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *checkResourceCountsErrorMockOperatorManager) DeleteResourceCollection() error {
	return nil
}

type deleteResourceCollectionErrorMockOperatorManager struct{}

func NewDeleteResourceCollectionErrorMockOperatorManager() *deleteResourceCollectionErrorMockOperatorManager {
	return &deleteResourceCollectionErrorMockOperatorManager{}
}

func (m *deleteResourceCollectionErrorMockOperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
}

func (m *deleteResourceCollectionErrorMockOperatorManager) CheckResourceCounts() error {
	return nil
}

func (m *deleteResourceCollectionErrorMockOperatorManager) GetLogicalResourceIds() []string {
	return []string{"logicalResourceId1", "logicalResourceId2"}
}

func (m *deleteResourceCollectionErrorMockOperatorManager) DeleteResourceCollection() error {
	return fmt.Errorf("DeleteResourceCollectionError")
}

/*
	Mocks for client
*/
var _ client.ICloudFormation = (*mockCloudFormation)(nil)
var _ client.ICloudFormation = (*terminationProtectionIsEnabledMockCloudFormation)(nil)
var _ client.ICloudFormation = (*notDeleteFailedMockCloudFormation)(nil)
var _ client.ICloudFormation = (*allErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*deleteStackErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*describeStacksErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*describeStacksNotExistsErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*listStackResourcesErrorMockCloudFormation)(nil)

type mockCloudFormation struct{}

func NewMockCloudFormation() *mockCloudFormation {
	return &mockCloudFormation{}
}

func (m *mockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *mockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:                   aws.String("StackName"),
				StackStatus:                 "DELETE_FAILED",
				EnableTerminationProtection: aws.Bool(false),
			},
		},
	}
	return output, true, nil
}

func (m *mockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.S3_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type terminationProtectionIsEnabledMockCloudFormation struct{}

func NewTerminationProtectionIsEnabledMockCloudFormation() *terminationProtectionIsEnabledMockCloudFormation {
	return &terminationProtectionIsEnabledMockCloudFormation{}
}

func (m *terminationProtectionIsEnabledMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *terminationProtectionIsEnabledMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:                   aws.String("StackName"),
				StackStatus:                 "CREATE_COMPLETE",
				EnableTerminationProtection: aws.Bool(true),
			},
		},
	}
	return output, true, nil
}

func (m *terminationProtectionIsEnabledMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.S3_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type notDeleteFailedMockCloudFormation struct{}

func NewNotDeleteFailedMockCloudFormation() *notDeleteFailedMockCloudFormation {
	return &notDeleteFailedMockCloudFormation{}
}

func (m *notDeleteFailedMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *notDeleteFailedMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:                   aws.String("StackName"),
				StackStatus:                 "UPDATE_ROLLBACK_COMPLETE",
				EnableTerminationProtection: aws.Bool(false),
			},
		},
	}
	return output, true, nil
}

func (m *notDeleteFailedMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "UPDATE_ROLLBACK_COMPLETE",
			ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "UPDATE_ROLLBACK_COMPLETE",
			ResourceType:       aws.String(resourcetype.S3_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type allErrorMockCloudFormation struct{}

func NewAllErrorMockCloudFormation() *allErrorMockCloudFormation {
	return &allErrorMockCloudFormation{}
}

func (m *allErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *allErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *allErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

type deleteStackErrorMockCloudFormation struct{}

func NewDeleteStackErrorMockCloudFormation() *deleteStackErrorMockCloudFormation {
	return &deleteStackErrorMockCloudFormation{}
}

func (m *deleteStackErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *deleteStackErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:                   aws.String("StackName"),
				StackStatus:                 "DELETE_FAILED",
				EnableTerminationProtection: aws.Bool(false),
			},
		},
	}
	return output, true, nil
}

func (m *deleteStackErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.S3_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type describeStacksErrorMockCloudFormation struct{}

func NewDescribeStacksErrorMockCloudFormation() *describeStacksErrorMockCloudFormation {
	return &describeStacksErrorMockCloudFormation{}
}

func (m *describeStacksErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *describeStacksErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *describeStacksErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.S3_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type describeStacksNotExistsErrorMockCloudFormation struct{}

func NewDescribeStacksNotExistsErrorMockCloudFormation() *describeStacksNotExistsErrorMockCloudFormation {
	return &describeStacksNotExistsErrorMockCloudFormation{}
}

func (m *describeStacksNotExistsErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *describeStacksNotExistsErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, false, fmt.Errorf("does not exist")
}

func (m *describeStacksNotExistsErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String(resourcetype.S3_STACK),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type listStackResourcesErrorMockCloudFormation struct{}

func NewListStackResourcesErrorMockCloudFormation() *listStackResourcesErrorMockCloudFormation {
	return &listStackResourcesErrorMockCloudFormation{}
}

func (m *listStackResourcesErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *listStackResourcesErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:                   aws.String("StackName"),
				StackStatus:                 "DELETE_FAILED",
				EnableTerminationProtection: aws.Bool(false),
			},
		},
	}
	return output, true, nil
}

func (m *listStackResourcesErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

/*
	Test Cases
*/
func TestDeleteStack(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()

	mock := NewMockCloudFormation()
	terminationProtectionIsEnabledMock := NewTerminationProtectionIsEnabledMockCloudFormation()
	notDeleteFailedMock := NewNotDeleteFailedMockCloudFormation()
	allErrorMock := NewAllErrorMockCloudFormation()
	deleteStackErrorMock := NewDeleteStackErrorMockCloudFormation()
	describeStacksErrorMock := NewDescribeStacksErrorMockCloudFormation()
	describeStacksNotExistsErrorMock := NewDescribeStacksNotExistsErrorMockCloudFormation()
	listStackResourcesErrorMock := NewListStackResourcesErrorMockCloudFormation()

	mockOperatorManager := NewMockOperatorManager()
	allErrorMockOperatorManager := NewAllErrorMockOperatorManager()
	checkResourceCountsErrorMockOperatorManager := NewCheckResourceCountsErrorMockOperatorManager()
	deleteResourceCollectionErrorMockOperatorManager := NewDeleteResourceCollectionErrorMockOperatorManager()

	type args struct {
		ctx                 context.Context
		stackName           *string
		isRootStack         bool
		clientMock          client.ICloudFormation
		operatorManagerMock IOperatorManager
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete stack successfully for root stack",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          mock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack successfully for child stack",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          mock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack failure for TerminationProtection is enabled stack",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          terminationProtectionIsEnabledMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("TerminationProtectionIsEnabled: test"),
			wantErr: true,
		},
		{
			name: "delete stack failure for not DELETE_FAILED stack",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          notDeleteFailedMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for all errors",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          allErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("DescribeStacksError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for all errors",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          allErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("ListStackResourcesError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for delete Stack",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          deleteStackErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("DeleteStackError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for delete Stack",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          deleteStackErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("DeleteStackError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for describe stacks",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          describeStacksErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("DescribeStacksError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for describe stacks not exist",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          describeStacksNotExistsErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("does not exist"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for list stack resources",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          listStackResourcesErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("ListStackResourcesError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for list stack resources",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          listStackResourcesErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("ListStackResourcesError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for operator manager all errors",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          mock,
				operatorManagerMock: allErrorMockOperatorManager,
			},
			want:    fmt.Errorf("CheckResourceCountsError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for operator manager all errors",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          mock,
				operatorManagerMock: allErrorMockOperatorManager,
			},
			want:    fmt.Errorf("CheckResourceCountsError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for operator manager check resource counts",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          mock,
				operatorManagerMock: checkResourceCountsErrorMockOperatorManager,
			},
			want:    fmt.Errorf("CheckResourceCountsError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for operator manager check resource counts",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          mock,
				operatorManagerMock: checkResourceCountsErrorMockOperatorManager,
			},
			want:    fmt.Errorf("CheckResourceCountsError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for operator manager delete resource collection",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          mock,
				operatorManagerMock: deleteResourceCollectionErrorMockOperatorManager,
			},
			want:    fmt.Errorf("DeleteResourceCollectionError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for operator manager delete resource collection",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         false,
				clientMock:          mock,
				operatorManagerMock: deleteResourceCollectionErrorMockOperatorManager,
			},
			want:    fmt.Errorf("DeleteResourceCollectionError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			targetResourceTypes := resourcetype.GetResourceTypes()
			cloudformationOperator := NewStackOperator(aws.Config{}, tt.args.clientMock, targetResourceTypes)

			err := cloudformationOperator.DeleteStackResources(tt.args.stackName, tt.args.isRootStack, tt.args.operatorManagerMock)
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
