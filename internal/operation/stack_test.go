package operation

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
)

/*
	Test Cases
*/

func TestStackOperator_DeleteStack(t *testing.T) {
	io.NewLogger(false)
	ctx := context.Background()

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
			want:    fmt.Errorf("DescribeStacksError"),
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
			name: "delete stack failure for root stack for describe stacks but not exists",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				isRootStack:         true,
				clientMock:          describeStacksNotExistsErrorMock,
				operatorManagerMock: mockOperatorManager,
			},
			want:    fmt.Errorf("NotExistsError: test stack not found."),
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
			targetResourceTypes := []string{
				"AWS::S3::Bucket",
				"AWS::IAM::Role",
				"AWS::ECR::Repository",
				"AWS::Backup::BackupVault",
				"AWS::CloudFormation::Stack",
				"Custom::",
			}
			cloudformationOperator := NewStackOperator(aws.Config{}, tt.args.clientMock, targetResourceTypes)

			err := cloudformationOperator.DeleteStackResources(tt.args.ctx, tt.args.stackName, tt.args.isRootStack, tt.args.operatorManagerMock)
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

func TestStackOperator_deleteRootStack(t *testing.T) {
	io.NewLogger(false)
	ctx := context.Background()

	mock := NewMockCloudFormation()
	terminationProtectionIsEnabledMock := NewTerminationProtectionIsEnabledMockCloudFormation()
	notDeleteFailedMock := NewNotDeleteFailedMockCloudFormation()
	allErrorMock := NewAllErrorMockCloudFormation()
	describeStacksErrorMock := NewDescribeStacksErrorMockCloudFormation()
	describeStacksNotExistsErrorMock := NewDescribeStacksNotExistsErrorMockCloudFormation()

	type args struct {
		ctx         context.Context
		stackName   *string
		isRootStack bool
		clientMock  client.ICloudFormation
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
			name: "delete stack successfully for root stack",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: true,
				clientMock:  mock,
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
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: true,
				clientMock:  terminationProtectionIsEnabledMock,
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
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: true,
				clientMock:  notDeleteFailedMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for all errors",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: true,
				clientMock:  allErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for root stack for describe stacks",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: true,
				clientMock:  describeStacksErrorMock,
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
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: true,
				clientMock:  describeStacksNotExistsErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("NotExistsError: test stack not found."),
			},
			wantErr: true,
		},
		{
			name: "delete stack successfully for child stack",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: false,
				clientMock:  mock,
			},
			want: want{
				got: false,
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete stack failure for child stack for TerminationProtection is enabled stack",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: false,
				clientMock:  terminationProtectionIsEnabledMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("TerminationProtectionIsEnabled: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for not DELETE_FAILED stack",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: false,
				clientMock:  notDeleteFailedMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but UPDATE_ROLLBACK_COMPLETE: test"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for all errors",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: false,
				clientMock:  allErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack failure for child stack for describe stacks",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: false,
				clientMock:  describeStacksErrorMock,
			},
			want: want{
				got: false,
				err: fmt.Errorf("DescribeStacksError"),
			},
			wantErr: true,
		},
		{
			name: "delete stack successfully for child stack for the stack already deleted",
			args: args{
				ctx:         ctx,
				stackName:   aws.String("test"),
				isRootStack: false,
				clientMock:  describeStacksNotExistsErrorMock,
			},
			want: want{
				got: true,
				err: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			targetResourceTypes := []string{
				"AWS::S3::Bucket",
				"AWS::IAM::Role",
				"AWS::ECR::Repository",
				"AWS::Backup::BackupVault",
				"AWS::CloudFormation::Stack",
				"Custom::",
			}
			cloudformationOperator := NewStackOperator(aws.Config{}, tt.args.clientMock, targetResourceTypes)

			got, err := cloudformationOperator.deleteStackNormally(tt.args.ctx, tt.args.stackName, tt.args.isRootStack)
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

func TestStackOperator_ListStacksFilteredByKeyword(t *testing.T) {
	io.NewLogger(false)
	ctx := context.Background()

	mock := NewMockCloudFormation()
	allErrorMock := NewAllErrorMockCloudFormation()
	listStacksErrorMock := NewListStacksErrorMockCloudFormation()
	listStacksEmptyMock := NewListStacksEmptyMockCloudFormation()

	type args struct {
		ctx        context.Context
		keyword    string
		clientMock client.ICloudFormation
	}

	type want struct {
		filteredStacks []string
		err            error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list stacks filtered by keyword successfully",
			args: args{
				ctx:        ctx,
				keyword:    "TestStack",
				clientMock: mock,
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
				ctx:        ctx,
				keyword:    "",
				clientMock: mock,
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
				ctx:        ctx,
				keyword:    "TestStack",
				clientMock: listStacksEmptyMock,
			},
			want: want{
				filteredStacks: []string{},
				err:            nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks filtered by keyword but empty keyword and no stacks found successfully",
			args: args{
				ctx:        ctx,
				keyword:    "",
				clientMock: listStacksEmptyMock,
			},
			want: want{
				filteredStacks: []string{},
				err:            nil,
			},
			wantErr: false,
		},
		{
			name: "list stacks filtered by keyword failure for all errors",
			args: args{
				ctx:        ctx,
				keyword:    "TestStack",
				clientMock: allErrorMock,
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("ListStacksError"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword failure for list stacks errors",
			args: args{
				ctx:        ctx,
				keyword:    "TestStack",
				clientMock: listStacksErrorMock,
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("ListStacksError"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword but empty keyword failure for all errors",
			args: args{
				ctx:        ctx,
				keyword:    "",
				clientMock: allErrorMock,
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("ListStacksError"),
			},
			wantErr: true,
		},
		{
			name: "list stacks filtered by keyword but empty keyword failure for list stacks errors",
			args: args{
				ctx:        ctx,
				keyword:    "",
				clientMock: listStacksErrorMock,
			},
			want: want{
				filteredStacks: []string{},
				err:            fmt.Errorf("ListStacksError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			targetResourceTypes := []string{
				"AWS::S3::Bucket",
				"AWS::IAM::Role",
				"AWS::ECR::Repository",
				"AWS::Backup::BackupVault",
				"AWS::CloudFormation::Stack",
				"Custom::",
			}

			cloudformationOperator := NewStackOperator(aws.Config{}, tt.args.clientMock, targetResourceTypes)

			output, err := cloudformationOperator.ListStacksFilteredByKeyword(tt.args.ctx, &tt.args.keyword)
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
