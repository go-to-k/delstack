package operation

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestCloudFormationStackOperator_DeleteCloudFormationStack(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx         context.Context
		stackName   *string
		isRootStack bool
	}

	cases := []struct {
		name                         string
		args                         args
		prepareMockCloudFormationFn  func(m *client.MockICloudFormation)
		prepareMockOperatorManagerFn func(m *MockIOperatorManager)
		want                         error
		wantErr                      bool
	}{
		{
			name: "delete stack successfully for root stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         nil,
			wantErr:                      false,
		},
		{
			name: "delete stack successfully for child stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         nil,
			wantErr:                      false,
		},
		{
			name: "delete stack failure for root stack for TerminationProtection is enabled stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "CREATE_COMPLETE",
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("TerminationProtectionError: test"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for TerminationProtection is enabled stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "CREATE_COMPLETE",
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("TerminationProtectionError: test"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for root stack for not DELETE_FAILED stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "UPDATE_ROLLBACK_COMPLETE",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for not DELETE_FAILED stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "UPDATE_ROLLBACK_COMPLETE",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for root stack for delete stack error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(fmt.Errorf("DeleteStackError"))
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("DeleteStackError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for delete stack error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(fmt.Errorf("DeleteStackError"))
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("DeleteStackError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for root stack for describe stacks error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("DescribeStacksError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for describe stacks error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("DescribeStacksError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for root stack for describe stacks but not exists",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("NotExistsError: test not found"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for describe stacks but not exists",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         nil,
			wantErr:                      false,
		},
		{
			name: "delete stack failure for root stack for describe stacks error after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("DescribeStacksError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for describe stacks error after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("DescribeStacksError"),
			wantErr:                      true,
		},
		{
			name: "delete stack success for root stack for no resources after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         nil,
			wantErr:                      false,
		},
		{
			name: "delete stack success for child stack for no resources after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         nil,
			wantErr:                      false,
		},
		{
			name: "delete stack failure for root stack for list stack resources error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{},
					fmt.Errorf("ListStackResourcesError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("ListStackResourcesError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for child stack for list stack resources error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{},
					fmt.Errorf("ListStackResourcesError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {},
			want:                         fmt.Errorf("ListStackResourcesError"),
			wantErr:                      true,
		},
		{
			name: "delete stack failure for root stack for operator manager check resource counts error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(fmt.Errorf("CheckResourceCountsError"))
			},
			want:    fmt.Errorf("CheckResourceCountsError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for operator manager check resource counts error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(fmt.Errorf("CheckResourceCountsError"))
			},
			want:    fmt.Errorf("CheckResourceCountsError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for operator manager delete resource collection error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(fmt.Errorf("DeleteResourceCollectionError"))
			},
			want:    fmt.Errorf("DeleteResourceCollectionError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for operator manager delete resource collection error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(fmt.Errorf("DeleteResourceCollectionError"))
			},
			want:    fmt.Errorf("DeleteResourceCollectionError"),
			wantErr: true,
		},
		{
			name: "delete stack success for root stack for delete stack at last",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1", "LogicalResourceId2"}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1", "LogicalResourceId2"})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack success for child stack for delete stack at last",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1", "LogicalResourceId2"}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1", "LogicalResourceId2"})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack failure for root stack for delete stack error at last",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1", "LogicalResourceId2"}).Return(fmt.Errorf("DeleteStackError"))
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1", "LogicalResourceId2"})
			},
			want:    fmt.Errorf("DeleteStackError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for delete stack error at last",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
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
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1", "LogicalResourceId2"}).Return(fmt.Errorf("DeleteStackError"))
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1", "LogicalResourceId2"})
			},
			want:    fmt.Errorf("DeleteStackError"),
			wantErr: true,
		},
		{
			name: "delete stack success for root stack after multiple loop iterations",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				// First iteration
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId1"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("PhysicalResourceId1"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1"}).Return(nil)

				// Second iteration - stack still DELETE_FAILED
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId2"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::IAM::Role"),
							PhysicalResourceId: aws.String("PhysicalResourceId2"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId2"}).Return(nil)

				// Final iteration - stack deleted
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				// First iteration
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1"})

				// Second iteration
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId2"})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack success for child stack after multiple loop iterations",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				// First iteration
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId1"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("PhysicalResourceId1"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1"}).Return(nil)

				// Second iteration - stack still DELETE_FAILED
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId2"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::IAM::Role"),
							PhysicalResourceId: aws.String("PhysicalResourceId2"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId2"}).Return(nil)

				// Final iteration - stack deleted
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				// First iteration
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1"})

				// Second iteration
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId2"})
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete stack failure for root stack for unexpected stack status in loop",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId1"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("PhysicalResourceId1"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1"}).Return(nil)

				// After delete, stack status is not DELETE_FAILED
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: "UPDATE_ROLLBACK_COMPLETE",
						},
					},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1"})
			},
			want:    fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for unexpected stack status in loop",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId1"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("PhysicalResourceId1"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1"}).Return(nil)

				// After delete, stack status is not DELETE_FAILED
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: "UPDATE_ROLLBACK_COMPLETE",
						},
					},
					nil,
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1"})
			},
			want:    fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for describe stacks error in loop",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId1"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("PhysicalResourceId1"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1"}).Return(nil)

				// DescribeStacks error in loop
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1"})
			},
			want:    fmt.Errorf("DescribeStacksError"),
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for describe stacks error in loop",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("LogicalResourceId1"),
							ResourceStatus:     "DELETE_FAILED",
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("PhysicalResourceId1"),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1"}).Return(nil)

				// DescribeStacks error in loop
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			prepareMockOperatorManagerFn: func(m *MockIOperatorManager) {
				m.EXPECT().SetOperatorCollection(aws.String("test"), gomock.Any()).Do(
					func(stackName *string, stackResourceSummaries []types.StackResourceSummary) {},
				)
				m.EXPECT().CheckResourceCounts().Return(nil)
				m.EXPECT().DeleteResourceCollection(gomock.Any()).Return(nil)
				m.EXPECT().GetLogicalResourceIds().Return([]string{"LogicalResourceId1"})
			},
			want:    fmt.Errorf("DescribeStacksError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)
			operatorManagerMock := NewMockIOperatorManager(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)
			tt.prepareMockOperatorManagerFn(operatorManagerMock)

			s3Mock := client.NewMockIS3(ctrl)
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, s3Mock)

			err := cloudformationStackOperator.DeleteCloudFormationStack(tt.args.ctx, tt.args.stackName, tt.args.isRootStack, operatorManagerMock)
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

func TestCloudFormationStackOperator_deleteStackNormally(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx         context.Context
		stackName   *string
		isRootStack bool
	}

	type want struct {
		got bool
		err error
	}

	cases := []struct {
		name                        string
		args                        args
		prepareMockCloudFormationFn func(m *client.MockICloudFormation)
		want                        want
		wantErr                     bool
	}{
		{
			name: "delete stack successfully for root stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)
			},
			want: want{
				got: false,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete stack successfully for child stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)
			},
			want: want{
				got: false,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete stack failure for root stack for TerminationProtection is enabled stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "CREATE_COMPLETE",
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("TerminationProtectionError: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for TerminationProtection is enabled stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "CREATE_COMPLETE",
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("TerminationProtectionError: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for not DELETE_FAILED stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "UPDATE_ROLLBACK_COMPLETE",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)
			},
			want: want{
				got: false,
				err: fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for not DELETE_FAILED stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "UPDATE_ROLLBACK_COMPLETE",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				).AnyTimes()

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)
			},
			want: want{
				got: false,
				err: fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for delete stack error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(fmt.Errorf("DeleteStackError"))
			},
			want: want{
				got: false,
				err: fmt.Errorf("DeleteStackError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for delete stack error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(fmt.Errorf("DeleteStackError"))
			},
			want: want{
				got: false,
				err: fmt.Errorf("DeleteStackError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for describe stacks error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for describe stacks error",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for describe stacks but not exists",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("NotExistsError: test not found"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for already deleted",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			want: want{
				got: true,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete stack failure for root stack for describe stacks error after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for describe stacks error after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack success for root stack for no resources after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: true,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			want: want{
				got: true,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete stack success for child stack for no resources after delete stack",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_FAILED",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{}).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			want: want{
				got: true,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete stack failure for root stack with operation in progress",
			args: args{
				ctx:         context.Background(),
				stackName:   aws.String("test"),
				isRootStack: false,
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("test"),
							StackStatus:                 "DELETE_IN_PROGRESS",
							EnableTerminationProtection: aws.Bool(false),
						},
					},
					nil,
				)
			},
			want: want{
				got: false,
				err: fmt.Errorf("OperationInProgressError: Stacks with XxxInProgress cannot be deleted, but DELETE_IN_PROGRESS: test"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)

			s3Mock := client.NewMockIS3(ctrl)
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, s3Mock)

			got, err := cloudformationStackOperator.deleteStackNormally(tt.args.ctx, tt.args.stackName, tt.args.isRootStack)
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

func TestCloudFormationStackOperator_GetSortedStackNames(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx        context.Context
		stackNames []string
	}

	type want struct {
		sortedStackNames []string
		err              error
	}

	cases := []struct {
		name                        string
		args                        args
		prepareMockCloudFormationFn func(m *client.MockICloudFormation)
		want                        want
		wantErr                     bool
	}{
		{
			name: "sort stacks with ascending order successfully",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack4"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-10 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{"Stack5", "Stack4", "Stack3", "Stack2", "Stack1"},
				err:              nil,
			},
			wantErr: false,
		},
		{
			name: "sort stacks with descending order successfully",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack5", "Stack4", "Stack3", "Stack2", "Stack1"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack4"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-10 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{"Stack5", "Stack4", "Stack3", "Stack2", "Stack1"},
				err:              nil,
			},
			wantErr: false,
		},
		{
			name: "sort stacks with random order successfully",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack3", "Stack4", "Stack1", "Stack2", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack4"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-10 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{"Stack5", "Stack4", "Stack3", "Stack2", "Stack1"},
				err:              nil,
			},
			wantErr: false,
		},
		{
			name: "sort stacks failure for non existent stacks",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("NotExistsError: Stack2, Stack4 not found"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for EnableTerminationProtection stacks",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("Stack2"),
							StackStatus:                 types.StackStatusCreateComplete,
							CreationTime:                aws.Time(time.Now().Add(-30 * time.Minute)),
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("Stack4"),
							StackStatus:                 types.StackStatusCreateComplete,
							CreationTime:                aws.Time(time.Now().Add(-10 * time.Minute)),
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("TerminationProtectionError: Stack2, Stack4"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for stacks in progress",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusUpdateInProgress,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack4"),
							StackStatus:  types.StackStatusRollbackInProgress,
							CreationTime: aws.Time(time.Now().Add(-10 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("OperationInProgressError: Stacks with XxxInProgress cannot be deleted, but UPDATE_IN_PROGRESS: Stack2, ROLLBACK_IN_PROGRESS: Stack4"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for stacks with some errors (not found, termination protection, and excepted status)",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusUpdateInProgress,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("Stack3"),
							StackStatus:                 types.StackStatusCreateComplete,
							CreationTime:                aws.Time(time.Now().Add(-20 * time.Minute)),
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("NotExistsError: Stack4 not found"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for stacks with some errors (not found and excepted status)",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusUpdateInProgress,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack3"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-20 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("NotExistsError: Stack4 not found"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for stacks with some errors (not found and termination protection)",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("Stack3"),
							StackStatus:                 types.StackStatusCreateComplete,
							CreationTime:                aws.Time(time.Now().Add(-20 * time.Minute)),
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("NotExistsError: Stack4 not found"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for stacks with some errors (termination protection and excepted status)",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2", "Stack3", "Stack4", "Stack5"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-40 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack2"),
							StackStatus:  types.StackStatusUpdateInProgress,
							CreationTime: aws.Time(time.Now().Add(-30 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack3")).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("Stack3"),
							StackStatus:                 types.StackStatusCreateComplete,
							CreationTime:                aws.Time(time.Now().Add(-20 * time.Minute)),
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack4")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack4"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-10 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack5")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack5"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now()),
						},
					},
					nil,
				)
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("TerminationProtectionError: Stack3"),
			},
			wantErr: true,
		},
		{
			name: "sort stacks failure for DescribeStacksError",
			args: args{
				ctx:        ctx,
				stackNames: []string{"Stack1", "Stack2"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack1")).Return(
					[]types.Stack{
						{
							StackName:    aws.String("Stack1"),
							StackStatus:  types.StackStatusCreateComplete,
							CreationTime: aws.Time(time.Now().Add(-10 * time.Minute)),
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("Stack2")).Return([]types.Stack{}, fmt.Errorf("DescribeStacksError"))
			},
			want: want{
				sortedStackNames: []string{},
				err:              fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)

			s3Mock := client.NewMockIS3(ctrl)
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, s3Mock)

			output, err := cloudformationStackOperator.GetSortedStackNames(tt.args.ctx, tt.args.stackNames)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.sortedStackNames) {
				t.Errorf("output = %#v, want %#v", output, tt.want.sortedStackNames)
			}
		})
	}
}

func TestCloudFormationStackOperator_ListStacksFilteredByKeyword(t *testing.T) {
	io.NewLogger(false)
	ctx := context.Background()

	type args struct {
		ctx     context.Context
		keyword string
	}

	type want struct {
		filteredStacks []string
		err            error
	}

	cases := []struct {
		name                        string
		args                        args
		prepareMockCloudFormationFn func(m *client.MockICloudFormation)
		want                        want
		wantErr                     bool
	}{
		{
			name: "list stacks filtered by keyword successfully",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateComplete,
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{
					"TestStack1",
					"TestStack2",
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks with RootId filtered by keyword successfully",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateComplete,
							RootId:      aws.String("test-stack-root"),
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{
					"TestStack2",
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks with RootId filtered by keyword and no stacks do not have RootId failure",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateComplete,
							RootId:      aws.String("test-stack-root"),
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusCreateComplete,
							RootId:      aws.String("test-stack-root"),
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by lower keyword successfully",
			args: args{
				ctx:     ctx,
				keyword: "teststack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateComplete,
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{
					"TestStack1",
					"TestStack2",
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks filtered by upper keyword successfully",
			args: args{
				ctx:     ctx,
				keyword: "TESTSTACK",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateComplete,
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{
					"TestStack1",
					"TestStack2",
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks filtered by keyword but empty keyword successfully",
			args: args{
				ctx:     ctx,
				keyword: "",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateComplete,
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{
					"TestStack1",
					"TestStack2",
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks filtered by keyword but no stacks failure",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return([]types.Stack{}, nil)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword but empty keyword and no stacks failure",
			args: args{
				ctx:     ctx,
				keyword: "",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return([]types.Stack{}, nil)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword ()"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword failure for describe stacks errors",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return([]types.Stack{}, fmt.Errorf("DescribeStacksError"))
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword but empty keyword failure for describe stacks errors",
			args: args{
				ctx:     ctx,
				keyword: "",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return([]types.Stack{}, fmt.Errorf("DescribeStacksError"))
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword with XxxInProgress stacks failure",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:   aws.String("TestStack1"),
							StackStatus: types.StackStatusCreateInProgress,
						},
						{
							StackName:   aws.String("TestStack2"),
							StackStatus: types.StackStatusUpdateRollbackCompleteCleanupInProgress,
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword with EnableTerminationProtection stacks successfully",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("TestStack1"),
							StackStatus:                 types.StackStatusCreateComplete,
							EnableTerminationProtection: aws.Bool(false),
						},
						{
							StackName:                   aws.String("TestStack2"),
							StackStatus:                 types.StackStatusCreateComplete,
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{
					"TestStack1",
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks filtered by keyword with EnableTerminationProtection stacks but empty failure",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return(
					[]types.Stack{
						{
							StackName:                   aws.String("TestStack1"),
							StackStatus:                 types.StackStatusCreateComplete,
							EnableTerminationProtection: aws.Bool(true),
						},
						{
							StackName:                   aws.String("TestStack2"),
							StackStatus:                 types.StackStatusCreateComplete,
							EnableTerminationProtection: aws.Bool(true),
						},
					},
					nil,
				)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)

			s3Mock := client.NewMockIS3(ctrl)
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, s3Mock)

			output, err := cloudformationStackOperator.ListStacksFilteredByKeyword(tt.args.ctx, &tt.args.keyword)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.filteredStacks) {
				t.Errorf("output = %#v, want %#v", output, tt.want.filteredStacks)
			}
		})
	}
}

func TestCloudFormationStackOperator_RemoveDeletionPolicy(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx       context.Context
		stackName *string
	}

	cases := []struct {
		name                        string
		args                        args
		prepareMockCloudFormationFn func(m *client.MockICloudFormation)
		prepareMockS3Fn             func(m *client.MockIS3)
		want                        error
		wantErr                     bool
	}{
		{
			name: "remove deletion policy successfully",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters: []types.Parameter{
								{
									ParameterKey:   aws.String("Key1"),
									ParameterValue: aws.String("Value1"),
								},
							},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					aws.String(`{
						"Resources": {
							"Resource1": {
								"Type": "AWS::S3::Bucket",
								"DeletionPolicy": "Retain"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().UpdateStack(gomock.Any(), aws.String("test"), gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy successfully for no deletion policy",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters: []types.Parameter{
								{
									ParameterKey:   aws.String("Key1"),
									ParameterValue: aws.String("Value1"),
								},
							},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					aws.String(`{
						"Resources": {
							"Resource1": {
								"Type": "AWS::S3::Bucket"
							}
						}
					}`),
					nil,
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy successfully for nested stack",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters: []types.Parameter{
								{
									ParameterKey:   aws.String("Key1"),
									ParameterValue: aws.String("Value1"),
								},
							},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("NestedStack"),
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("test-nested"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					aws.String(`{
						"Resources": {
							"NestedStack": {
								"Type": "AWS::CloudFormation::Stack",
								"DeletionPolicy": "Retain"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().UpdateStack(gomock.Any(), aws.String("test"), gomock.Any(), gomock.Any()).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-nested")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-nested"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters: []types.Parameter{
								{
									ParameterKey:   aws.String("Key1"),
									ParameterValue: aws.String("Value1"),
								},
							},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-nested")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-nested")).Return(
					aws.String(`{
						"Resources": {
							"Resource1": {
								"Type": "AWS::S3::Bucket",
								"DeletionPolicy": "Retain"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().UpdateStack(gomock.Any(), aws.String("test-nested"), gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy successfully for no template changes with nested stack",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("NestedStack"),
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("test-nested"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					aws.String(`{
						"Resources": {
							"NestedStack": {
								"Type": "AWS::CloudFormation::Stack"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-nested")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-nested"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-nested")).Return(
					[]types.StackResourceSummary{},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-nested")).Return(
					aws.String(`{
						"Resources": {
							"Resource1": {
								"Type": "AWS::S3::Bucket",
								"DeletionPolicy": "Retain"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().UpdateStack(gomock.Any(), aws.String("test-nested"), gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy failure for describe stacks error",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					fmt.Errorf("DescribeStacksError"),
				)
			},
			want:    fmt.Errorf("DescribeStacksError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy failure for not exists",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{},
					nil,
				)
			},
			want:    fmt.Errorf("NotExistsError: test"),
			wantErr: true,
		},
		{
			name: "remove deletion policy successfully for rollback complete",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusRollbackComplete,
						},
					},
					nil,
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy failure for list stack resources error",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{},
					fmt.Errorf("ListStackResourcesError"),
				)
			},
			want:    fmt.Errorf("ListStackResourcesError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy failure for get template error",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					nil,
					fmt.Errorf("GetTemplateError"),
				)
			},
			want:    fmt.Errorf("GetTemplateError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy failure for update stack error",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters: []types.Parameter{
								{
									ParameterKey:   aws.String("Key1"),
									ParameterValue: aws.String("Value1"),
								},
							},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					aws.String(`{
						"Resources": {
							"Resource1": {
								"Type": "AWS::S3::Bucket",
								"DeletionPolicy": "Retain"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().UpdateStack(gomock.Any(), aws.String("test"), gomock.Any(), gomock.Any()).Return(fmt.Errorf("UpdateStackError"))
			},
			want:    fmt.Errorf("UpdateStackError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy failure for nested stack error",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test"),
							StackStatus: types.StackStatusCreateComplete,
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("NestedStack"),
							ResourceType:       aws.String("AWS::CloudFormation::Stack"),
							PhysicalResourceId: aws.String("test-nested"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test")).Return(
					aws.String(`{
						"Resources": {
							"NestedStack": {
								"Type": "AWS::CloudFormation::Stack",
								"DeletionPolicy": "Retain"
							}
						}
					}`),
					nil,
				)

				m.EXPECT().UpdateStack(gomock.Any(), aws.String("test"), gomock.Any(), gomock.Any()).Return(nil)

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-nested")).Return(
					nil,
					fmt.Errorf("DescribeStacksError"),
				)
			},
			want:    fmt.Errorf("DescribeStacksError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy with large template using S3",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				// Create a large template that exceeds 51200 bytes
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				// Add padding to exceed 51200 bytes
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1] // Remove last comma
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large")).Return(
					aws.String(largeTemplate),
					nil,
				)

				m.EXPECT().UpdateStackWithTemplateURL(gomock.Any(), aws.String("test-large"), gomock.Any(), gomock.Any()).Return(nil)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().DeleteObjects(gomock.Any(), gomock.Any(), gomock.Any()).Return([]s3types.Error{}, nil)
				m.EXPECT().DeleteBucket(gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy with large template using S3 but UpdateStack fails - should still cleanup S3",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-fail"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				// Create a large template that exceeds 51200 bytes
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-fail")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-fail"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-fail/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-fail")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-fail")).Return(
					aws.String(largeTemplate),
					nil,
				)

				// UpdateStack fails
				m.EXPECT().UpdateStackWithTemplateURL(gomock.Any(), aws.String("test-large-fail"), gomock.Any(), gomock.Any()).Return(fmt.Errorf("UpdateStackError"))
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				// Even if UpdateStack fails, DeleteObjects and DeleteBucket should still be called (via defer)
				m.EXPECT().DeleteObjects(gomock.Any(), gomock.Any(), gomock.Any()).Return([]s3types.Error{}, nil)
				m.EXPECT().DeleteBucket(gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    fmt.Errorf("TemplateS3UpdateError: failed to update stack with large template via S3: UpdateStackError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy with large template - PutObject fails should cleanup bucket",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-putobject-fail"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				// Create a large template
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-putobject-fail")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-putobject-fail"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-putobject-fail/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-putobject-fail")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-putobject-fail")).Return(
					aws.String(largeTemplate),
					nil,
				)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				// PutObject fails
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("PutObjectError"))
				// Bucket should be cleaned up (via defer in uploadTemplateToS3)
				m.EXPECT().DeleteBucket(gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    fmt.Errorf("TemplateS3UploadError: failed to upload template to S3: PutObjectError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy with large template - CreateBucket fails",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-createbucket-fail"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-createbucket-fail")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-createbucket-fail"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-createbucket-fail/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-createbucket-fail")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-createbucket-fail")).Return(
					aws.String(largeTemplate),
					nil,
				)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				// CreateBucket fails
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(fmt.Errorf("CreateBucketError"))
			},
			want:    fmt.Errorf("TemplateS3UploadError: failed to create S3 bucket: CreateBucketError"),
			wantErr: true,
		},
		{
			name: "remove deletion policy with large template - missing account ID",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-no-account"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-no-account")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-no-account"),
							StackId:     aws.String("invalid-arn"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-no-account")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-no-account")).Return(
					aws.String(largeTemplate),
					nil,
				)
			},
			want:    fmt.Errorf("TemplateS3UploadError: failed to extract account ID from stack ARN"),
			wantErr: true,
		},
		{
			name: "remove deletion policy with large template - DeleteObjects fails but operation succeeds with warning",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-deleteobjects-fail"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-deleteobjects-fail")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-deleteobjects-fail"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-deleteobjects-fail/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-deleteobjects-fail")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-deleteobjects-fail")).Return(
					aws.String(largeTemplate),
					nil,
				)

				m.EXPECT().UpdateStackWithTemplateURL(gomock.Any(), aws.String("test-large-deleteobjects-fail"), gomock.Any(), gomock.Any()).Return(nil)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				// DeleteObjects fails
				m.EXPECT().DeleteObjects(gomock.Any(), gomock.Any(), gomock.Any()).Return([]s3types.Error{}, fmt.Errorf("DeleteObjectsError"))
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy with large template - DeleteBucket fails but operation succeeds with warning",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-deletebucket-fail"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-deletebucket-fail")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-deletebucket-fail"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-deletebucket-fail/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-deletebucket-fail")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-deletebucket-fail")).Return(
					aws.String(largeTemplate),
					nil,
				)

				m.EXPECT().UpdateStackWithTemplateURL(gomock.Any(), aws.String("test-large-deletebucket-fail"), gomock.Any(), gomock.Any()).Return(nil)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().DeleteObjects(gomock.Any(), gomock.Any(), gomock.Any()).Return([]s3types.Error{}, nil)
				// DeleteBucket fails
				m.EXPECT().DeleteBucket(gomock.Any(), gomock.Any()).Return(fmt.Errorf("DeleteBucketError"))
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy with large template - DeleteObjects returns errors but operation succeeds with warning",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-deleteobjects-errors"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-deleteobjects-errors")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-deleteobjects-errors"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-deleteobjects-errors/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-deleteobjects-errors")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-deleteobjects-errors")).Return(
					aws.String(largeTemplate),
					nil,
				)

				m.EXPECT().UpdateStackWithTemplateURL(gomock.Any(), aws.String("test-large-deleteobjects-errors"), gomock.Any(), gomock.Any()).Return(nil)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				// DeleteObjects returns error list
				m.EXPECT().DeleteObjects(gomock.Any(), gomock.Any(), gomock.Any()).Return([]s3types.Error{
					{
						Key:     aws.String("test.template"),
						Code:    aws.String("InternalError"),
						Message: aws.String("Internal error occurred"),
					},
				}, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "remove deletion policy with large template - both DeleteObjects and DeleteBucket fail but operation succeeds with warning",
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-large-both-delete-fail"),
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				largeTemplate := `{"Resources":{"Resource1":{"Type":"AWS::S3::Bucket","DeletionPolicy":"Retain","Properties":{"BucketName":"test-bucket","Tags":[`
				for i := 0; i < 5000; i++ {
					largeTemplate += `{"Key":"tag` + fmt.Sprintf("%d", i) + `","Value":"value` + fmt.Sprintf("%d", i) + `"},`
				}
				largeTemplate = largeTemplate[:len(largeTemplate)-1]
				largeTemplate += `]}}}}`

				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("test-large-both-delete-fail")).Return(
					[]types.Stack{
						{
							StackName:   aws.String("test-large-both-delete-fail"),
							StackId:     aws.String("arn:aws:cloudformation:us-east-1:123456789012:stack/test-large-both-delete-fail/guid"),
							StackStatus: types.StackStatusCreateComplete,
							Parameters:  []types.Parameter{},
						},
					},
					nil,
				)

				m.EXPECT().ListStackResources(gomock.Any(), aws.String("test-large-both-delete-fail")).Return(
					[]types.StackResourceSummary{
						{
							LogicalResourceId:  aws.String("Resource1"),
							ResourceType:       aws.String("AWS::S3::Bucket"),
							PhysicalResourceId: aws.String("test-bucket"),
						},
					},
					nil,
				)

				m.EXPECT().GetTemplate(gomock.Any(), aws.String("test-large-both-delete-fail")).Return(
					aws.String(largeTemplate),
					nil,
				)

				m.EXPECT().UpdateStackWithTemplateURL(gomock.Any(), aws.String("test-large-both-delete-fail"), gomock.Any(), gomock.Any()).Return(nil)
			},
			prepareMockS3Fn: func(m *client.MockIS3) {
				m.EXPECT().CreateBucket(gomock.Any(), gomock.Any()).Return(nil)
				m.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				// Both DeleteObjects and DeleteBucket would fail, but DeleteObjects is called first
				m.EXPECT().DeleteObjects(gomock.Any(), gomock.Any(), gomock.Any()).Return([]s3types.Error{}, fmt.Errorf("DeleteObjectsError"))
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)

			s3Mock := client.NewMockIS3(ctrl)
			if tt.prepareMockS3Fn != nil {
				tt.prepareMockS3Fn(s3Mock)
			}
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{Region: "us-east-1"}, cloudformationMock, s3Mock)

			err := cloudformationStackOperator.RemoveDeletionPolicy(tt.args.ctx, tt.args.stackName)
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

func TestCloudFormationStackOperator_BuildDependencyGraph(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx        context.Context
		stackNames []string
	}

	cases := []struct {
		name                        string
		args                        args
		prepareMockCloudFormationFn func(m *client.MockICloudFormation)
		wantDependencies            map[string]map[string]struct{}
		wantErr                     bool
		wantErrMsg                  string
	}{
		{
			name: "no dependencies between stacks",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a", "stack-b"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs:   []types.Output{},
						},
					},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-b")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-b"),
							Outputs:   []types.Output{},
						},
					},
					nil,
				)
			},
			wantDependencies: map[string]map[string]struct{}{},
			wantErr:          false,
		},
		{
			name: "simple dependency (B depends on A)",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a", "stack-b"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					[]string{"stack-b"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-b")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-b"),
							Outputs:   []types.Output{},
						},
					},
					nil,
				)
			},
			wantDependencies: map[string]map[string]struct{}{
				"stack-b": {
					"stack-a": {},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple outputs referenced (deduplication test)",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a", "stack-b"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA1"),
									OutputValue: aws.String("value-a1"),
									ExportName:  aws.String("export-a-1"),
								},
								{
									OutputKey:   aws.String("ExportA2"),
									OutputValue: aws.String("value-a2"),
									ExportName:  aws.String("export-a-2"),
								},
								{
									OutputKey:   aws.String("ExportA3"),
									OutputValue: aws.String("value-a3"),
									ExportName:  aws.String("export-a-3"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a-1")).Return(
					[]string{"stack-b"},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a-2")).Return(
					[]string{"stack-b"},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a-3")).Return(
					[]string{"stack-b"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-b")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-b"),
							Outputs:   []types.Output{},
						},
					},
					nil,
				)
			},
			wantDependencies: map[string]map[string]struct{}{
				"stack-b": {
					"stack-a": {},
				},
			},
			wantErr: false,
		},
		{
			name: "external stack reference causes error",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a", "stack-b"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					[]string{"stack-b", "external-stack"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-b")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-b"),
							Outputs:   []types.Output{},
						},
					},
					nil,
				)
			},
			wantDependencies: nil,
			wantErr:          true,
			wantErrMsg:       "deletion would break dependencies for non-target stacks:\nStack 'stack-a' exports 'export-a' which is imported by non-target stack(s) 'external-stack'",
		},
		{
			name: "multiple external stack references cause error",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a", "stack-b"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					[]string{"external-stack-1", "external-stack-2"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-b")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-b"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportB"),
									OutputValue: aws.String("value-b"),
									ExportName:  aws.String("export-b"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-b")).Return(
					[]string{"external-stack-3"},
					nil,
				)
			},
			wantDependencies: nil,
			wantErr:          true,
			wantErrMsg:       "deletion would break dependencies for non-target stacks:\nStack 'stack-a' exports 'export-a' which is imported by non-target stack(s) 'external-stack-1', 'external-stack-2'\nStack 'stack-b' exports 'export-b' which is imported by non-target stack(s) 'external-stack-3'",
		},
		{
			name: "single stack with external reference causes error",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					[]string{"external-stack"},
					nil,
				)
			},
			wantDependencies: nil,
			wantErr:          true,
			wantErrMsg:       "deletion would break dependencies for non-target stacks:\nStack 'stack-a' exports 'export-a' which is imported by non-target stack(s) 'external-stack'",
		},
		{
			name: "diamond dependency",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a", "stack-b", "stack-c", "stack-d"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					[]string{"stack-b", "stack-c"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-b")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-b"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportB"),
									OutputValue: aws.String("value-b"),
									ExportName:  aws.String("export-b"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-b")).Return(
					[]string{"stack-d"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-c")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-c"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportC"),
									OutputValue: aws.String("value-c"),
									ExportName:  aws.String("export-c"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-c")).Return(
					[]string{"stack-d"},
					nil,
				)
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-d")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-d"),
							Outputs:   []types.Output{},
						},
					},
					nil,
				)
			},
			wantDependencies: map[string]map[string]struct{}{
				"stack-b": {
					"stack-a": {},
				},
				"stack-c": {
					"stack-a": {},
				},
				"stack-d": {
					"stack-b": {},
					"stack-c": {},
				},
			},
			wantErr: false,
		},
		{
			name: "error on DescribeStacks",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					nil,
					fmt.Errorf("DescribeStacks error"),
				)
			},
			wantDependencies: nil,
			wantErr:          true,
		},
		{
			name: "export is not imported by any stack (should not error)",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					[]string{},
					nil,
				)
			},
			wantDependencies: map[string]map[string]struct{}{},
			wantErr:          false,
		},
		{
			name: "error on ListImports (non-validation error)",
			args: args{
				ctx:        context.Background(),
				stackNames: []string{"stack-a"},
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), aws.String("stack-a")).Return(
					[]types.Stack{
						{
							StackName: aws.String("stack-a"),
							Outputs: []types.Output{
								{
									OutputKey:   aws.String("ExportA"),
									OutputValue: aws.String("value-a"),
									ExportName:  aws.String("export-a"),
								},
							},
						},
					},
					nil,
				)
				m.EXPECT().ListImports(gomock.Any(), aws.String("export-a")).Return(
					nil,
					fmt.Errorf("Some other ListImports error"),
				)
			},
			wantDependencies: nil,
			wantErr:          true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)

			s3Mock := client.NewMockIS3(ctrl)
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, s3Mock)

			graph, err := cloudformationStackOperator.BuildDependencyGraph(tt.args.ctx, tt.args.stackNames)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrMsg != "" {
				if err.Error() != tt.wantErrMsg {
					t.Errorf("error message = %v, want %v", err.Error(), tt.wantErrMsg)
				}
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(graph.dependencies, tt.wantDependencies) {
					t.Errorf("dependencies = %v, want %v", graph.dependencies, tt.wantDependencies)
				}
			}
		})
	}
}
