package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"
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

/*
	Test Cases
*/
func TestCloudFormation_DeleteStack(t *testing.T) {
	ctx := context.Background()
	mockWaiter := NewMockCloudFormationSDKWaiter()
	failureErrorMockWaiter := NewFailureErrorMockCloudFormationSDKWaiter()
	otherErrorMockWaiter := NewOtherErrorMockCloudFormationSDKWaiter()
	mock := NewMockCloudFormationSDKClient()
	errorMock := NewErrorMockCloudFormationSDKClient()

	type args struct {
		ctx             context.Context
		stackName       *string
		retainResources []string
		mockClient      ICloudFormationSDKClient
		mockWaiter      ICloudFormationSDKWaiter
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete stack successfully",
			args: args{
				ctx:             ctx,
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				mockClient:      mock,
				mockWaiter:      mockWaiter,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack successfully including non retainResources",
			args: args{
				ctx:             ctx,
				stackName:       aws.String("test"),
				retainResources: []string{},
				mockClient:      mock,
				mockWaiter:      mockWaiter,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack successfully for transitioned to Failure",
			args: args{
				ctx:             ctx,
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				mockClient:      mock,
				mockWaiter:      failureErrorMockWaiter,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack failure",
			args: args{
				ctx:             ctx,
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				mockClient:      errorMock,
				mockWaiter:      mockWaiter,
			},
			want:    fmt.Errorf("DeleteStackError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for other errors",
			args: args{
				ctx:             ctx,
				stackName:       aws.String("test"),
				retainResources: []string{"test1", "test2"},
				mockClient:      mock,
				mockWaiter:      otherErrorMockWaiter,
			},
			want:    fmt.Errorf("WaitError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cloudformationClient := NewCloudFormation(tt.args.mockClient, tt.args.mockWaiter)

			err := cloudformationClient.DeleteStack(tt.args.stackName, tt.args.retainResources)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestCloudFormation_DescribeStacks(t *testing.T) {
	ctx := context.Background()
	mockWaiter := NewMockCloudFormationSDKWaiter()
	mock := NewMockCloudFormationSDKClient()
	errorMock := NewErrorMockCloudFormationSDKClient()
	notExistsMock := NewNotExistsMockCloudFormationSDKClient()

	type args struct {
		ctx        context.Context
		stackName  *string
		mockClient ICloudFormationSDKClient
		mockWaiter ICloudFormationSDKWaiter
	}

	type want struct {
		output *cloudformation.DescribeStacksOutput
		exists bool
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "describe stacks successfully",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: mock,
				mockWaiter: mockWaiter,
			},
			want: want{
				output: &cloudformation.DescribeStacksOutput{
					Stacks: []types.Stack{
						{
							StackName:   aws.String("StackName"),
							StackStatus: "DELETE_FAILED",
						},
					},
				},
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "describe stacks failure",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: errorMock,
				mockWaiter: mockWaiter,
			},
			want: want{
				output: &cloudformation.DescribeStacksOutput{
					Stacks: []types.Stack{
						{
							StackName:   aws.String("StackName"),
							StackStatus: "DELETE_FAILED",
						},
					},
				},
				exists: true,
				err:    fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "describe stacks but not exist",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: notExistsMock,
				mockWaiter: mockWaiter,
			},
			want: want{
				output: &cloudformation.DescribeStacksOutput{},
				exists: false,
				err:    nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cloudformationClient := NewCloudFormation(tt.args.mockClient, tt.args.mockWaiter)

			output, exists, err := cloudformationClient.DescribeStacks(tt.args.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
				return
			}
			if exists != tt.want.exists {
				t.Errorf("exists = %#v, want %#v", exists, tt.want.exists)
			}
		})
	}
}

func TestCloudFormation_waitDeleteStack(t *testing.T) {
	ctx := context.Background()
	mockWaiter := NewMockCloudFormationSDKWaiter()
	failureErrorMockWaiter := NewFailureErrorMockCloudFormationSDKWaiter()
	otherErrorMockWaiter := NewOtherErrorMockCloudFormationSDKWaiter()
	mock := NewMockCloudFormationSDKClient()

	type args struct {
		ctx        context.Context
		stackName  *string
		mockClient ICloudFormationSDKClient
		mockWaiter ICloudFormationSDKWaiter
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "wait successfully",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: mock,
				mockWaiter: mockWaiter,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "wait failure for other error",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: mock,
				mockWaiter: otherErrorMockWaiter,
			},
			want:    fmt.Errorf("WaitError"),
			wantErr: true,
		},
		{
			name: "wait failure for transitioned to Failure",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: mock,
				mockWaiter: failureErrorMockWaiter,
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cloudformationClient := NewCloudFormation(tt.args.mockClient, tt.args.mockWaiter)

			err := cloudformationClient.waitDeleteStack(tt.args.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestCloudFormation_ListStackResources(t *testing.T) {
	ctx := context.Background()
	mockWaiter := NewMockCloudFormationSDKWaiter()
	mock := NewMockCloudFormationSDKClient()
	errorMock := NewErrorMockCloudFormationSDKClient()

	type args struct {
		ctx        context.Context
		stackName  *string
		mockClient ICloudFormationSDKClient
		mockWaiter ICloudFormationSDKWaiter
	}

	type want struct {
		output []types.StackResourceSummary
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list stack resources successfully",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: mock,
				mockWaiter: mockWaiter,
			},
			want: want{
				output: []types.StackResourceSummary{
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
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stack resources failure",
			args: args{
				ctx:        ctx,
				stackName:  aws.String("test"),
				mockClient: errorMock,
				mockWaiter: mockWaiter,
			},
			want: want{
				output: []types.StackResourceSummary{},
				err:    fmt.Errorf("ListStackResourcesError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			cloudformationClient := NewCloudFormation(tt.args.mockClient, tt.args.mockWaiter)

			output, err := cloudformationClient.ListStackResources(tt.args.stackName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}
