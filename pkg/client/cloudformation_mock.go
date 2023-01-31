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

var _ ICloudformation = (*MockCloudformation)(nil)
var _ ICloudformation = (*TerminationProtectionIsEnabledMockCloudformation)(nil)
var _ ICloudformation = (*NotDeleteFailedMockCloudformation)(nil)
var _ ICloudformation = (*AllErrorMockCloudformation)(nil)
var _ ICloudformation = (*DeleteStackErrorMockCloudformation)(nil)
var _ ICloudformation = (*DescribeStacksErrorMockCloudformation)(nil)
var _ ICloudformation = (*DescribeStacksNotExistsErrorMockCloudformation)(nil)
var _ ICloudformation = (*ListStackResourcesErrorMockCloudformation)(nil)
var _ ICloudformation = (*ListStacksErrorMockCloudformation)(nil)
var _ ICloudformation = (*ListStacksEmptyMockCloudformation)(nil)

type MockCloudformation struct{}

func NewMockCloudformation() *MockCloudformation {
	return &MockCloudformation{}
}

func (m *MockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *MockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *MockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *MockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type TerminationProtectionIsEnabledMockCloudformation struct{}

func NewTerminationProtectionIsEnabledMockCloudformation() *TerminationProtectionIsEnabledMockCloudformation {
	return &TerminationProtectionIsEnabledMockCloudformation{}
}

func (m *TerminationProtectionIsEnabledMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *TerminationProtectionIsEnabledMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *TerminationProtectionIsEnabledMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *TerminationProtectionIsEnabledMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type NotDeleteFailedMockCloudformation struct{}

func NewNotDeleteFailedMockCloudformation() *NotDeleteFailedMockCloudformation {
	return &NotDeleteFailedMockCloudformation{}
}

func (m *NotDeleteFailedMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *NotDeleteFailedMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *NotDeleteFailedMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *NotDeleteFailedMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type AllErrorMockCloudformation struct{}

func NewAllErrorMockCloudformation() *AllErrorMockCloudformation {
	return &AllErrorMockCloudformation{}
}

func (m *AllErrorMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *AllErrorMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *AllErrorMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

func (m *AllErrorMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{}

	return output, fmt.Errorf("ListStacksError")
}

type DeleteStackErrorMockCloudformation struct{}

func NewDeleteStackErrorMockCloudformation() *DeleteStackErrorMockCloudformation {
	return &DeleteStackErrorMockCloudformation{}
}

func (m *DeleteStackErrorMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return fmt.Errorf("DeleteStackError")
}

func (m *DeleteStackErrorMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *DeleteStackErrorMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *DeleteStackErrorMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type DescribeStacksErrorMockCloudformation struct{}

func NewDescribeStacksErrorMockCloudformation() *DescribeStacksErrorMockCloudformation {
	return &DescribeStacksErrorMockCloudformation{}
}

func (m *DescribeStacksErrorMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *DescribeStacksErrorMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, true, fmt.Errorf("DescribeStacksError")
}

func (m *DescribeStacksErrorMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *DescribeStacksErrorMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type DescribeStacksNotExistsErrorMockCloudformation struct{}

func NewDescribeStacksNotExistsErrorMockCloudformation() *DescribeStacksNotExistsErrorMockCloudformation {
	return &DescribeStacksNotExistsErrorMockCloudformation{}
}

func (m *DescribeStacksNotExistsErrorMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *DescribeStacksNotExistsErrorMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	output := &cloudformation.DescribeStacksOutput{}
	return output, false, nil
}

func (m *DescribeStacksNotExistsErrorMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *DescribeStacksNotExistsErrorMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type ListStackResourcesErrorMockCloudformation struct{}

func NewListStackResourcesErrorMockCloudformation() *ListStackResourcesErrorMockCloudformation {
	return &ListStackResourcesErrorMockCloudformation{}
}

func (m *ListStackResourcesErrorMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStackResourcesErrorMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStackResourcesErrorMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	output := []types.StackResourceSummary{}

	return output, fmt.Errorf("ListStackResourcesError")
}

func (m *ListStackResourcesErrorMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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

type ListStacksErrorMockCloudformation struct{}

func NewListStacksErrorMockCloudformation() *ListStacksErrorMockCloudformation {
	return &ListStacksErrorMockCloudformation{}
}

func (m *ListStacksErrorMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStacksErrorMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStacksErrorMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *ListStacksErrorMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{}

	return output, fmt.Errorf("ListStacksError")
}

type ListStacksEmptyMockCloudformation struct{}

func NewListStacksEmptyMockCloudformation() *ListStacksEmptyMockCloudformation {
	return &ListStacksEmptyMockCloudformation{}
}

func (m *ListStacksEmptyMockCloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	return nil
}

func (m *ListStacksEmptyMockCloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
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

func (m *ListStacksEmptyMockCloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (m *ListStacksEmptyMockCloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	output := []types.StackSummary{}

	return output, nil
}
