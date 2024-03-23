package operation

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "github.com/golang/mock/gomock"
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
			want:                         fmt.Errorf("TerminationProtectionIsEnabled: test"),
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
			want:                         fmt.Errorf("TerminationProtectionIsEnabled: test"),
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
			want:                         fmt.Errorf("NotExistsError: test stack not found."),
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

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1", "LogicalResourceId2"}).Return(nil)
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

				m.EXPECT().DeleteStack(gomock.Any(), aws.String("test"), []string{"LogicalResourceId1", "LogicalResourceId2"}).Return(nil)
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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)
			operatorManagerMock := NewMockIOperatorManager(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)
			tt.prepareMockOperatorManagerFn(operatorManagerMock)

			targetResourceTypes := []string{
				"AWS::S3::Bucket",
				"AWS::IAM::Role",
				"AWS::ECR::Repository",
				"AWS::Backup::BackupVault",
				"AWS::CloudFormation::Stack",
				"Custom::",
			}

			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, targetResourceTypes)

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
				err: fmt.Errorf("TerminationProtectionIsEnabled: test"),
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
				err: fmt.Errorf("TerminationProtectionIsEnabled: test"),
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
				err: fmt.Errorf("NotExistsError: test stack not found."),
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

			targetResourceTypes := []string{
				"AWS::S3::Bucket",
				"AWS::IAM::Role",
				"AWS::ECR::Repository",
				"AWS::Backup::BackupVault",
				"AWS::CloudFormation::Stack",
				"Custom::",
			}
			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, targetResourceTypes)

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

// TODO: Tests for GetStackNamesSorted with complicated order

// TODO: Add new tests for empty error handlings
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
			name: "list stacks with RootId filtered by keyword and no stacks do not have RootId found successfully",
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
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)."),
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
			name: "list stacks filtered by keyword but no stacks found successfully",
			args: args{
				ctx:     ctx,
				keyword: "TestStack",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return([]types.Stack{}, nil)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)."),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword but empty keyword and no stacks found successfully",
			args: args{
				ctx:     ctx,
				keyword: "",
			},
			prepareMockCloudFormationFn: func(m *client.MockICloudFormation) {
				m.EXPECT().DescribeStacks(gomock.Any(), nil).Return([]types.Stack{}, nil)
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword ()."),
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
			name: "list stacks filtered by keyword with XxxInProgress stacks successfully",
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
				err:            fmt.Errorf("NotExistsError: No stacks matching the keyword (TestStack)."),
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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cloudformationMock := client.NewMockICloudFormation(ctrl)

			tt.prepareMockCloudFormationFn(cloudformationMock)

			targetResourceTypes := []string{
				"AWS::S3::Bucket",
				"AWS::IAM::Role",
				"AWS::ECR::Repository",
				"AWS::Backup::BackupVault",
				"AWS::CloudFormation::Stack",
				"Custom::",
			}

			cloudformationStackOperator := NewCloudFormationStackOperator(aws.Config{}, cloudformationMock, targetResourceTypes)

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
