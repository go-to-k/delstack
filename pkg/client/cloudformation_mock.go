package client

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

/*
	Mocks for SDK Waiter
*/

var _ ICloudFormationSDKWaiter = (*MockCloudFormationSDKWaiter)(nil)
var _ ICloudFormationSDKWaiter = (*OtherErrorMockCloudFormationSDKWaiter)(nil)
var _ ICloudFormationSDKWaiter = (*FailureErrorMockCloudFormationSDKWaiter)(nil)

type MockCloudFormationSDKWaiter struct{}

func NewMockCloudFormationSDKWaiter() *MockCloudFormationSDKWaiter {
	return &MockCloudFormationSDKWaiter{}
}

func (m *MockCloudFormationSDKWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	return nil
}

type OtherErrorMockCloudFormationSDKWaiter struct{}

func NewOtherErrorMockCloudFormationSDKWaiter() *OtherErrorMockCloudFormationSDKWaiter {
	return &OtherErrorMockCloudFormationSDKWaiter{}
}

func (m *OtherErrorMockCloudFormationSDKWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	return fmt.Errorf("WaitError")
}

type FailureErrorMockCloudFormationSDKWaiter struct{}

func NewFailureErrorMockCloudFormationSDKWaiter() *FailureErrorMockCloudFormationSDKWaiter {
	return &FailureErrorMockCloudFormationSDKWaiter{}
}

func (m *FailureErrorMockCloudFormationSDKWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	return fmt.Errorf("waiter state transitioned to Failure")
}

/*
	Mocks for SDK Client
*/

var _ ICloudFormationSDKClient = (*MockCloudFormationSDKClient)(nil)
var _ ICloudFormationSDKClient = (*ErrorMockCloudFormationSDKClient)(nil)
var _ ICloudFormationSDKClient = (*NotExistsMockCloudFormationSDKClient)(nil)
var _ ICloudFormationSDKClient = (*EmptyMockCloudFormationSDKClient)(nil)

type MockCloudFormationSDKClient struct{}

func NewMockCloudFormationSDKClient() *MockCloudFormationSDKClient {
	return &MockCloudFormationSDKClient{}
}

func (m *MockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, nil
}

func (m *MockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:   aws.String("StackName"),
				StackStatus: "DELETE_FAILED",
			},
		},
	}
	return output, nil
}

func (m *MockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
	output := &cloudformation.ListStackResourcesOutput{
		StackResourceSummaries: []types.StackResourceSummary{
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
		},
	}
	return output, nil
}

func (m *MockCloudFormationSDKClient) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	output := &cloudformation.ListStacksOutput{
		StackSummaries: []types.StackSummary{
			{
				StackName:   aws.String("TestStack1"),
				StackStatus: types.StackStatusCreateComplete,
			},
			{
				StackName:   aws.String("TestStack2"),
				StackStatus: types.StackStatusCreateComplete,
			},
		},
	}
	return output, nil
}

type ErrorMockCloudFormationSDKClient struct{}

func NewErrorMockCloudFormationSDKClient() *ErrorMockCloudFormationSDKClient {
	return &ErrorMockCloudFormationSDKClient{}
}

func (m *ErrorMockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, fmt.Errorf("DeleteStackError")
}

func (m *ErrorMockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:   aws.String("StackName"),
				StackStatus: "DELETE_FAILED",
			},
		},
	}
	return output, fmt.Errorf("DescribeStacksError")
}

func (m *ErrorMockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
	output := &cloudformation.ListStackResourcesOutput{
		StackResourceSummaries: []types.StackResourceSummary{
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
		},
	}
	return output, fmt.Errorf("ListStackResourcesError")
}

func (m *ErrorMockCloudFormationSDKClient) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	output := &cloudformation.ListStacksOutput{
		StackSummaries: []types.StackSummary{
			{
				StackName:   aws.String("TestStack1"),
				StackStatus: types.StackStatusCreateComplete,
			},
			{
				StackName:   aws.String("TestStack2"),
				StackStatus: types.StackStatusCreateComplete,
			},
		},
	}
	return output, fmt.Errorf("ListStacksError")
}

type NotExistsMockCloudFormationSDKClient struct{}

func NewNotExistsMockCloudFormationSDKClient() *NotExistsMockCloudFormationSDKClient {
	return &NotExistsMockCloudFormationSDKClient{}
}

func (m *NotExistsMockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, nil
}

func (m *NotExistsMockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return &cloudformation.DescribeStacksOutput{}, fmt.Errorf("does not exist")
}

func (m *NotExistsMockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
	return nil, nil
}

func (m *NotExistsMockCloudFormationSDKClient) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	return nil, nil
}

type EmptyMockCloudFormationSDKClient struct{}

func NewEmptyMockCloudFormationSDKClient() *EmptyMockCloudFormationSDKClient {
	return &EmptyMockCloudFormationSDKClient{}
}

func (m *EmptyMockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, nil
}

func (m *EmptyMockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	output := &cloudformation.DescribeStacksOutput{
		Stacks: []types.Stack{
			{
				StackName:   aws.String("StackName"),
				StackStatus: "DELETE_FAILED",
			},
		},
	}
	return output, nil
}

func (m *EmptyMockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
	output := &cloudformation.ListStackResourcesOutput{
		StackResourceSummaries: []types.StackResourceSummary{
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
		},
	}
	return output, nil
}

func (m *EmptyMockCloudFormationSDKClient) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	output := &cloudformation.ListStacksOutput{
		StackSummaries: []types.StackSummary{},
	}
	return output, nil
}
