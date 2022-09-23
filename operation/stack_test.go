package operation

import (
	"context"
	"fmt"
	"reflect"
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

func (m *MockOperatorManager) DeleteResourceCollection() error {
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

func (m *AllErrorMockOperatorManager) DeleteResourceCollection() error {
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

func (m *CheckResourceCountsErrorMockOperatorManager) DeleteResourceCollection() error {
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

func (m *DeleteResourceCollectionErrorMockOperatorManager) DeleteResourceCollection() error {
	return fmt.Errorf("DeleteResourceCollectionError")
}

/*
	Mocks for client
*/
var _ client.ICloudFormation = (*MockCloudFormation)(nil)
var _ client.ICloudFormation = (*TerminationProtectionIsEnabledMockCloudFormation)(nil)
var _ client.ICloudFormation = (*NotDeleteFailedMockCloudFormation)(nil)
var _ client.ICloudFormation = (*AllErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*DeleteStackErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*DescribeStacksErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*DescribeStacksNotExistsErrorMockCloudFormation)(nil)
var _ client.ICloudFormation = (*ListStackResourcesErrorMockCloudFormation)(nil)

type MockCloudFormation struct{}

func NewMockCloudFormation() *MockCloudFormation {
	return &MockCloudFormation{}
}

func (m *MockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *MockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *MockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			ResourceType:       aws.String(resourcetype.S3_BUCKET),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type TerminationProtectionIsEnabledMockCloudFormation struct{}

func NewTerminationProtectionIsEnabledMockCloudFormation() *TerminationProtectionIsEnabledMockCloudFormation {
	return &TerminationProtectionIsEnabledMockCloudFormation{}
}

func (m *TerminationProtectionIsEnabledMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *TerminationProtectionIsEnabledMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *TerminationProtectionIsEnabledMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			ResourceType:       aws.String(resourcetype.S3_BUCKET),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type NotDeleteFailedMockCloudFormation struct{}

func NewNotDeleteFailedMockCloudFormation() *NotDeleteFailedMockCloudFormation {
	return &NotDeleteFailedMockCloudFormation{}
}

func (m *NotDeleteFailedMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *NotDeleteFailedMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *NotDeleteFailedMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			ResourceType:       aws.String(resourcetype.S3_BUCKET),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type AllErrorMockCloudFormation struct{}

func NewAllErrorMockCloudFormation() *AllErrorMockCloudFormation {
	return &AllErrorMockCloudFormation{}
}

func (m *AllErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *AllErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *AllErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

type DeleteStackErrorMockCloudFormation struct{}

func NewDeleteStackErrorMockCloudFormation() *DeleteStackErrorMockCloudFormation {
	return &DeleteStackErrorMockCloudFormation{}
}

func (m *DeleteStackErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *DeleteStackErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *DeleteStackErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			ResourceType:       aws.String(resourcetype.S3_BUCKET),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type DescribeStacksErrorMockCloudFormation struct{}

func NewDescribeStacksErrorMockCloudFormation() *DescribeStacksErrorMockCloudFormation {
	return &DescribeStacksErrorMockCloudFormation{}
}

func (m *DescribeStacksErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *DescribeStacksErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *DescribeStacksErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			ResourceType:       aws.String(resourcetype.S3_BUCKET),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type DescribeStacksNotExistsErrorMockCloudFormation struct{}

func NewDescribeStacksNotExistsErrorMockCloudFormation() *DescribeStacksNotExistsErrorMockCloudFormation {
	return &DescribeStacksNotExistsErrorMockCloudFormation{}
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, false, nil
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			ResourceType:       aws.String(resourcetype.S3_BUCKET),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

type ListStackResourcesErrorMockCloudFormation struct{}

func NewListStackResourcesErrorMockCloudFormation() *ListStackResourcesErrorMockCloudFormation {
	return &ListStackResourcesErrorMockCloudFormation{}
}

func (m *ListStackResourcesErrorMockCloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStackResourcesErrorMockCloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStackResourcesErrorMockCloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
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
			want:    fmt.Errorf("NotExistsError: test"),
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

func Test_deleteRootStack(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()

	mock := NewMockCloudFormation()
	terminationProtectionIsEnabledMock := NewTerminationProtectionIsEnabledMockCloudFormation()
	notDeleteFailedMock := NewNotDeleteFailedMockCloudFormation()
	allErrorMock := NewAllErrorMockCloudFormation()
	describeStacksErrorMock := NewDescribeStacksErrorMockCloudFormation()
	describeStacksNotExistsErrorMock := NewDescribeStacksNotExistsErrorMockCloudFormation()

	type args struct {
		ctx        context.Context
		stackName  *string
		clientMock client.ICloudFormation
	}

	type want struct {
		got bool
		err error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "delete root stack successfully",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				clientMock: mock,
			},
			want: want{
				got: false,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete root stack failure for TerminationProtection is enabled stack",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				clientMock: terminationProtectionIsEnabledMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("TerminationProtectionIsEnabled: test"),
			},
			wantErr: true,
		},
		{
			name: "delete root stack failure for not DELETE_FAILED stack",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				clientMock: notDeleteFailedMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			},
			wantErr: true,
		},
		{
			name: "delete root stack failure for all errors",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				clientMock: allErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete root stack failure for describe stacks",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				clientMock: describeStacksErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete root stack failure for describe stacks not exist",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				clientMock: describeStacksNotExistsErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("NotExistsError: test"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			targetResourceTypes := resourcetype.GetResourceTypes()
			cloudformationOperator := NewStackOperator(aws.Config{}, tt.args.clientMock, targetResourceTypes)

			got, err := cloudformationOperator.deleteRootStack(tt.args.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("output = %#v, want %#v", got, tt.want.got)
			}
		})
	}
}
