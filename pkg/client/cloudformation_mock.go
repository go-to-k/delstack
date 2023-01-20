package client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
	Mocks for client
*/

var _ ICloudFormation = (*MockCloudFormation)(nil)
var _ ICloudFormation = (*TerminationProtectionIsEnabledMockCloudFormation)(nil)
var _ ICloudFormation = (*NotDeleteFailedMockCloudFormation)(nil)
var _ ICloudFormation = (*AllErrorMockCloudFormation)(nil)
var _ ICloudFormation = (*DeleteStackErrorMockCloudFormation)(nil)
var _ ICloudFormation = (*DescribeStacksErrorMockCloudFormation)(nil)
var _ ICloudFormation = (*DescribeStacksNotExistsErrorMockCloudFormation)(nil)
var _ ICloudFormation = (*ListStackResourcesErrorMockCloudFormation)(nil)
var _ ICloudFormation = (*ListStacksErrorMockCloudFormation)(nil)
var _ ICloudFormation = (*ListStacksEmptyMockCloudFormation)(nil)

type MockCloudFormation struct{}

func NewMockCloudFormation() *MockCloudFormation {
	return &MockCloudFormation{}
}

func (m *MockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *MockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *MockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *MockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type TerminationProtectionIsEnabledMockCloudFormation struct{}

func NewTerminationProtectionIsEnabledMockCloudFormation() *TerminationProtectionIsEnabledMockCloudFormation {
	return &TerminationProtectionIsEnabledMockCloudFormation{}
}

func (m *TerminationProtectionIsEnabledMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *TerminationProtectionIsEnabledMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *TerminationProtectionIsEnabledMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *TerminationProtectionIsEnabledMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type NotDeleteFailedMockCloudFormation struct{}

func NewNotDeleteFailedMockCloudFormation() *NotDeleteFailedMockCloudFormation {
	return &NotDeleteFailedMockCloudFormation{}
}

func (m *NotDeleteFailedMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *NotDeleteFailedMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *NotDeleteFailedMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "UPDATE_ROLLBACK_COMPLETE",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "UPDATE_ROLLBACK_COMPLETE",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *NotDeleteFailedMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type AllErrorMockCloudFormation struct{}

func NewAllErrorMockCloudFormation() *AllErrorMockCloudFormation {
	return &AllErrorMockCloudFormation{}
}

func (m *AllErrorMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *AllErrorMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *AllErrorMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

func (m *AllErrorMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{}

	return output, fmt.Errorf("ListStacksError")
}

type DeleteStackErrorMockCloudFormation struct{}

func NewDeleteStackErrorMockCloudFormation() *DeleteStackErrorMockCloudFormation {
	return &DeleteStackErrorMockCloudFormation{}
}

func (m *DeleteStackErrorMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *DeleteStackErrorMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *DeleteStackErrorMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *DeleteStackErrorMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type DescribeStacksErrorMockCloudFormation struct{}

func NewDescribeStacksErrorMockCloudFormation() *DescribeStacksErrorMockCloudFormation {
	return &DescribeStacksErrorMockCloudFormation{}
}

func (m *DescribeStacksErrorMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *DescribeStacksErrorMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *DescribeStacksErrorMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *DescribeStacksErrorMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type DescribeStacksNotExistsErrorMockCloudFormation struct{}

func NewDescribeStacksNotExistsErrorMockCloudFormation() *DescribeStacksNotExistsErrorMockCloudFormation {
	return &DescribeStacksNotExistsErrorMockCloudFormation{}
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, false, nil
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *DescribeStacksNotExistsErrorMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type ListStackResourcesErrorMockCloudFormation struct{}

func NewListStackResourcesErrorMockCloudFormation() *ListStackResourcesErrorMockCloudFormation {
	return &ListStackResourcesErrorMockCloudFormation{}
}

func (m *ListStackResourcesErrorMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStackResourcesErrorMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStackResourcesErrorMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

func (m *ListStackResourcesErrorMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{
		{
			StackName:   aws.String("TestStack1"),
			StackStatus: types.StackStatusCreateComplete,
		},
		{
			StackName:   aws.String("TestStack2"),
			StackStatus: types.StackStatusCreateComplete,
		},
	}

	return output, nil
}

type ListStacksErrorMockCloudFormation struct{}

func NewListStacksErrorMockCloudFormation() *ListStacksErrorMockCloudFormation {
	return &ListStacksErrorMockCloudFormation{}
}

func (m *ListStacksErrorMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStacksErrorMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStacksErrorMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *ListStacksErrorMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{}

	return output, fmt.Errorf("ListStacksError")
}

type ListStacksEmptyMockCloudFormation struct{}

func NewListStacksEmptyMockCloudFormation() *ListStacksEmptyMockCloudFormation {
	return &ListStacksEmptyMockCloudFormation{}
}

func (m *ListStacksEmptyMockCloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStacksEmptyMockCloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStacksEmptyMockCloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("LogicalResourceId1"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::CloudFormation::Stack"),
			PhysicalResourceId: aws.String("PhysicalResourceId1"),
		},
		{
			LogicalResourceId:  aws.String("LogicalResourceId2"),
			ResourceStatus:     "DELETE_FAILED",
			ResourceType:       aws.String("AWS::S3::Bucket"),
			PhysicalResourceId: aws.String("PhysicalResourceId2"),
		},
	}

	return output, nil
}

func (m *ListStacksEmptyMockCloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{}

	return output, nil
}
