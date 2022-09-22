package operation

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
)

/*
	Test Cases
*/
func TestSetOperatorCollection(t *testing.T) {
	logger.NewLogger()
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

	targetResourceTypesForAllServices := resourcetype.GetResourceTypes()
	targetResourceTypesForPartialServices := []string{
		resourcetype.S3_BUCKET,
		resourcetype.IAM_ROLE,
		resourcetype.CUSTOM_RESOURCE,
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.S3_BUCKET),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.IAM_ROLE),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.ECR_REPOSITORY),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId5"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.BACKUP_VAULT),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
						PhysicalResourceId: aws.String("PhysicalResourceId1"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.S3_BUCKET),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.IAM_ROLE),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.ECR_REPOSITORY),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId5"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String(resourcetype.BACKUP_VAULT),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
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
						ResourceType:       aws.String(resourcetype.CLOUDFORMATION_STACK),
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
						ResourceType:       aws.String(resourcetype.S3_BUCKET),
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
						ResourceType:       aws.String(resourcetype.S3_BUCKET),
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
						ResourceType:       aws.String(resourcetype.S3_BUCKET),
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
						ResourceType:       aws.String(resourcetype.CUSTOM_RESOURCE),
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
						ResourceType:       aws.String(resourcetype.CUSTOM_RESOURCE),
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
						ResourceType:       aws.String(resourcetype.CUSTOM_RESOURCE),
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
