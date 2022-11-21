package operation

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/logger"
)

var targetResourceTypesForAllServices = []string{
	"AWS::S3::Bucket",
	"AWS::IAM::Role",
	"AWS::ECR::Repository",
	"AWS::Backup::BackupVault",
	"AWS::CloudFormation::Stack",
	"Custom::",
}

var targetResourceTypesForPartialServices = []string{
	"AWS::S3::Bucket",
	"AWS::IAM::Role",
	"Custom::",
}

/*
	Test Cases
*/
func TestOperatorCollection_SetOperatorCollection(t *testing.T) {
	logger.NewLogger(false)
	ctx := context.TODO()

	type args struct {
		ctx                    context.Context
		stackName              *string
		targetResourceTypes    []string
		stackResourceSummaries []types.StackResourceSummary
	}

	type want struct {
		logicalResourceIdsLength        int
		unsupportedStackResourcesLength int
	}

	cases := []struct {
		name string
		args args
		want want
	}{
		{
			name: "resource counts check 1 for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				stackResourceSummaries: []types.StackResourceSummary{
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
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::IAM::Role"),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::ECR::Repository"),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId5"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::Backup::BackupVault"),
						PhysicalResourceId: aws.String("PhysicalResourceId5"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId6"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("Custom::CustomResource"),
						PhysicalResourceId: aws.String("PhysicalResourceId6"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        6,
				unsupportedStackResourcesLength: 0,
			},
		},
		{
			name: "resource counts check 2 for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        2,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 3 for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 4 for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 0,
			},
		},
		{
			name: "resource counts check 1 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
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
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::IAM::Role"),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::ECR::Repository"),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId5"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::Backup::BackupVault"),
						PhysicalResourceId: aws.String("PhysicalResourceId5"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId6"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("Custom::CustomResource"),
						PhysicalResourceId: aws.String("PhysicalResourceId6"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        6,
				unsupportedStackResourcesLength: 3,
			},
		},
		{
			name: "resource counts check 2 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        2,
				unsupportedStackResourcesLength: 2,
			},
		},
		{
			name: "resource counts check 3 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 4 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 5 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        2,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 6 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 7 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::S3::Bucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 0,
			},
		},
		{
			name: "resource counts check 8 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("Custom::"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        2,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 9 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("Custom::"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 1,
			},
		},
		{
			name: "resource counts check 10 for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				stackResourceSummaries: []types.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("Custom::"),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:        1,
				unsupportedStackResourcesLength: 0,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			config := aws.Config{}
			operatorFactory := NewOperatorFactory(config)
			operatorCollection := NewOperatorCollection(config, operatorFactory, tt.args.targetResourceTypes)

			operatorCollection.SetOperatorCollection(tt.args.stackName, tt.args.stackResourceSummaries)

			got := want{
				logicalResourceIdsLength:        len(operatorCollection.logicalResourceIds),
				unsupportedStackResourcesLength: len(operatorCollection.unsupportedStackResources),
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestOperatorCollection_containsResourceType(t *testing.T) {
	logger.NewLogger(false)
	ctx := context.TODO()

	type args struct {
		ctx                 context.Context
		stackName           *string
		targetResourceTypes []string
		resource            string
	}

	cases := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "S3 Bucket for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::S3::Bucket",
			},
			want: true,
		},
		{
			name: "IAM Role for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::IAM::Role",
			},
			want: true,
		},
		{
			name: "ECR Repository for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::ECR::Repository",
			},
			want: true,
		},
		{
			name: "BACKUP VAULT for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::Backup::BackupVault",
			},
			want: true,
		},
		{
			name: "CloudFormation Stack for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::CloudFormation::Stack",
			},
			want: true,
		},
		{
			name: "custom resource for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "Custom::Abc",
			},
			want: true,
		},
		{
			name: "unsupported resource for all target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::DynamoDB::Table",
			},
			want: false,
		},
		{
			name: "exists in partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				resource:            "AWS::S3::Bucket",
			},
			want: true,
		},
		{
			name: "not exists in partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				resource:            "AWS::Backup::BackupVault",
			},
			want: false,
		},
		{
			name: "custom resource exists in for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				resource:            "Custom::Abc",
			},
			want: true,
		},
		{
			name: "unsupported resource for partial target resource types",
			args: args{
				ctx:                 ctx,
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForPartialServices,
				resource:            "AWS::DynamoDB::Table",
			},
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			config := aws.Config{}
			operatorFactory := NewOperatorFactory(config)
			operatorCollection := NewOperatorCollection(config, operatorFactory, tt.args.targetResourceTypes)

			got := operatorCollection.containsResourceType(tt.args.resource)

			if got != tt.want {
				t.Errorf("got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
