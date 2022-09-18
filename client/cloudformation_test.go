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
	"github.com/go-to-k/delstack/logger"
)

/*
	Mocks for SDK Waiter
*/
var _ ICloudFormationSDKWaiter = (*mockCloudFormationSDKWaiter)(nil)
var _ ICloudFormationSDKWaiter = (*otherErrorMockCloudFormationSDKWaiter)(nil)
var _ ICloudFormationSDKWaiter = (*failureErrorMockCloudFormationSDKWaiter)(nil)

type mockCloudFormationSDKWaiter struct{}

func NewMockCloudFormationSDKWaiter() *mockCloudFormationSDKWaiter {
	return &mockCloudFormationSDKWaiter{}
}

func (m *mockCloudFormationSDKWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	return nil
}

type otherErrorMockCloudFormationSDKWaiter struct{}

func NewOtherErrorMockCloudFormationSDKWaiter() *otherErrorMockCloudFormationSDKWaiter {
	return &otherErrorMockCloudFormationSDKWaiter{}
}

func (m *otherErrorMockCloudFormationSDKWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	return fmt.Errorf("WaitError")
}

type failureErrorMockCloudFormationSDKWaiter struct{}

func NewFailureErrorMockCloudFormationSDKWaiter() *failureErrorMockCloudFormationSDKWaiter {
	return &failureErrorMockCloudFormationSDKWaiter{}
}

func (m *failureErrorMockCloudFormationSDKWaiter) Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error {
	return fmt.Errorf("waiter state transitioned to Failure")
}

/*
	Mocks for SDK Client
*/
var _ ICloudFormationSDKClient = (*mockCloudFormationSDKClient)(nil)
var _ ICloudFormationSDKClient = (*errorMockCloudFormationSDKClient)(nil)
var _ ICloudFormationSDKClient = (*notExistsMockCloudFormationSDKClient)(nil)

type mockCloudFormationSDKClient struct{}

func NewMockCloudFormationSDKClient() *mockCloudFormationSDKClient {
	return &mockCloudFormationSDKClient{}
}

func (m *mockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, nil
}

func (m *mockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
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

func (m *mockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
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

type errorMockCloudFormationSDKClient struct{}

func NewErrorMockCloudFormationSDKClient() *errorMockCloudFormationSDKClient {
	return &errorMockCloudFormationSDKClient{}
}

func (m *errorMockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, fmt.Errorf("DeleteStackError")
}

func (m *errorMockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
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

func (m *errorMockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
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

type notExistsMockCloudFormationSDKClient struct{}

func NewNotExistsMockCloudFormationSDKClient() *notExistsMockCloudFormationSDKClient {
	return &notExistsMockCloudFormationSDKClient{}
}

func (m *notExistsMockCloudFormationSDKClient) DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error) {
	return nil, nil
}

func (m *notExistsMockCloudFormationSDKClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return &cloudformation.DescribeStacksOutput{}, fmt.Errorf("does not exist")
}

func (m *notExistsMockCloudFormationSDKClient) ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error) {
	return nil, nil
}

/*
	Test Cases
*/
func TestDeleteStack(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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

func TestDescribeStacks(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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

func TestListStackResources(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
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
