package operation

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
)

var targetResourceTypesForAllServices = resourcetype.GetResourceTypes()

/*
	Test Cases
*/

func TestOperatorCollection_SetOperatorCollection(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx                    context.Context
		stackName              *string
		targetResourceTypes    []string
		stackResourceSummaries []types.StackResourceSummary
	}

	type want struct {
		logicalResourceIdsLength                   int
		unsupportedStackResourcesLength            int
		s3BucketOperatorResourcesLength            int
		s3DirectoryBucketOperatorResourcesLength   int
		s3TableBucketOperatorResourcesLength       int
		S3TableNamespaceOperatorResourcesLength    int
		s3VectorBucketOperatorResourcesLength      int
		iamGroupOperatorResourcesLength            int
		ecrRepositoryOperatorResourcesLength       int
		backupVaultOperatorResourcesLength         int
		cloudformationStackOperatorResourcesLength int
		customOperatorResourcesLength              int
	}

	cases := []struct {
		name string
		args args
		want want
	}{
		{
			name: "resource counts check 1 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
						ResourceType:       aws.String("AWS::S3Express::DirectoryBucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::S3Tables::TableBucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId5"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::S3Tables::Namespace"),
						PhysicalResourceId: aws.String("PhysicalResourceId5"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId6"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::S3Vectors::VectorBucket"),
						PhysicalResourceId: aws.String("PhysicalResourceId6"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId7"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::IAM::Group"),
						PhysicalResourceId: aws.String("PhysicalResourceId7"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId8"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::ECR::Repository"),
						PhysicalResourceId: aws.String("PhysicalResourceId8"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId9"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::Backup::BackupVault"),
						PhysicalResourceId: aws.String("PhysicalResourceId9"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId10"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("Custom::CustomResource"),
						PhysicalResourceId: aws.String("PhysicalResourceId10"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:                   10,
				unsupportedStackResourcesLength:            0,
				s3BucketOperatorResourcesLength:            1,
				s3DirectoryBucketOperatorResourcesLength:   1,
				s3TableBucketOperatorResourcesLength:       1,
				S3TableNamespaceOperatorResourcesLength:    1,
				s3VectorBucketOperatorResourcesLength:      1,
				iamGroupOperatorResourcesLength:            1,
				ecrRepositoryOperatorResourcesLength:       1,
				backupVaultOperatorResourcesLength:         1,
				cloudformationStackOperatorResourcesLength: 1,
				customOperatorResourcesLength:              1,
			},
		},
		{
			name: "resource counts check 2 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
				logicalResourceIdsLength:                   2,
				unsupportedStackResourcesLength:            1,
				s3BucketOperatorResourcesLength:            0,
				s3DirectoryBucketOperatorResourcesLength:   0,
				s3TableBucketOperatorResourcesLength:       0,
				S3TableNamespaceOperatorResourcesLength:    0,
				s3VectorBucketOperatorResourcesLength:      0,
				iamGroupOperatorResourcesLength:            0,
				ecrRepositoryOperatorResourcesLength:       0,
				backupVaultOperatorResourcesLength:         0,
				cloudformationStackOperatorResourcesLength: 1,
				customOperatorResourcesLength:              0,
			},
		},
		{
			name: "resource counts check 3 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:                   4,
				unsupportedStackResourcesLength:            2,
				s3BucketOperatorResourcesLength:            0,
				s3DirectoryBucketOperatorResourcesLength:   0,
				s3TableBucketOperatorResourcesLength:       0,
				S3TableNamespaceOperatorResourcesLength:    0,
				s3VectorBucketOperatorResourcesLength:      0,
				iamGroupOperatorResourcesLength:            0,
				ecrRepositoryOperatorResourcesLength:       0,
				backupVaultOperatorResourcesLength:         0,
				cloudformationStackOperatorResourcesLength: 2,
				customOperatorResourcesLength:              0,
			},
		},
		{
			name: "resource counts check 4 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
				logicalResourceIdsLength:                   1,
				unsupportedStackResourcesLength:            1,
				s3BucketOperatorResourcesLength:            0,
				s3DirectoryBucketOperatorResourcesLength:   0,
				s3TableBucketOperatorResourcesLength:       0,
				S3TableNamespaceOperatorResourcesLength:    0,
				s3VectorBucketOperatorResourcesLength:      0,
				iamGroupOperatorResourcesLength:            0,
				ecrRepositoryOperatorResourcesLength:       0,
				backupVaultOperatorResourcesLength:         0,
				cloudformationStackOperatorResourcesLength: 0,
				customOperatorResourcesLength:              0,
			},
		},
		{
			name: "resource counts check 5 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:                   2,
				unsupportedStackResourcesLength:            2,
				s3BucketOperatorResourcesLength:            0,
				s3DirectoryBucketOperatorResourcesLength:   0,
				s3TableBucketOperatorResourcesLength:       0,
				S3TableNamespaceOperatorResourcesLength:    0,
				s3VectorBucketOperatorResourcesLength:      0,
				iamGroupOperatorResourcesLength:            0,
				ecrRepositoryOperatorResourcesLength:       0,
				backupVaultOperatorResourcesLength:         0,
				cloudformationStackOperatorResourcesLength: 0,
				customOperatorResourcesLength:              0,
			},
		},
		{
			name: "resource counts check 6 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
				logicalResourceIdsLength:                   1,
				unsupportedStackResourcesLength:            0,
				s3BucketOperatorResourcesLength:            0,
				s3DirectoryBucketOperatorResourcesLength:   0,
				s3TableBucketOperatorResourcesLength:       0,
				S3TableNamespaceOperatorResourcesLength:    0,
				s3VectorBucketOperatorResourcesLength:      0,
				iamGroupOperatorResourcesLength:            0,
				ecrRepositoryOperatorResourcesLength:       0,
				backupVaultOperatorResourcesLength:         0,
				cloudformationStackOperatorResourcesLength: 1,
				customOperatorResourcesLength:              0,
			},
		},
		{
			name: "resource counts check 7 for all target resource types",
			args: args{
				ctx:                 context.Background(),
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
						ResourceType:       aws.String("AWS::CloudFormation::Stack"),
						PhysicalResourceId: aws.String("PhysicalResourceId2"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId3"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId3"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId4"),
						ResourceStatus:     "DELETE_COMPLETE",
						ResourceType:       aws.String("AWS::DynamoDB::Table"),
						PhysicalResourceId: aws.String("PhysicalResourceId4"),
					},
				},
			},
			want: want{
				logicalResourceIdsLength:                   2,
				unsupportedStackResourcesLength:            0,
				s3BucketOperatorResourcesLength:            0,
				s3DirectoryBucketOperatorResourcesLength:   0,
				s3TableBucketOperatorResourcesLength:       0,
				S3TableNamespaceOperatorResourcesLength:    0,
				s3VectorBucketOperatorResourcesLength:      0,
				iamGroupOperatorResourcesLength:            0,
				ecrRepositoryOperatorResourcesLength:       0,
				backupVaultOperatorResourcesLength:         0,
				cloudformationStackOperatorResourcesLength: 2,
				customOperatorResourcesLength:              0,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			config := aws.Config{}
			operatorFactory := NewOperatorFactory(config)
			operatorCollection := NewOperatorCollection(config, operatorFactory, tt.args.targetResourceTypes)

			operatorCollection.SetOperatorCollection(tt.args.stackName, tt.args.stackResourceSummaries)

			s3BucketOperatorResourcesLength := 0
			s3DirectoryBucketOperatorResourcesLength := 0
			s3TableBucketOperatorResourcesLength := 0
			S3TableNamespaceOperatorResourcesLength := 0
			s3VectorBucketOperatorResourcesLength := 0
			iamGroupOperatorResourcesLength := 0
			ecrRepositoryOperatorResourcesLength := 0
			backupVaultOperatorResourcesLength := 0
			cloudformationStackOperatorResourcesLength := 0
			customOperatorResourcesLength := 0

			for _, operator := range operatorCollection.GetOperators() {
				switch operator := operator.(type) {
				case *S3BucketOperator:
					if operator.GetDirectoryBucketsFlag() {
						s3DirectoryBucketOperatorResourcesLength += operator.GetResourcesLength()
					} else {
						s3BucketOperatorResourcesLength += operator.GetResourcesLength()
					}
				case *S3TableBucketOperator:
					s3TableBucketOperatorResourcesLength += operator.GetResourcesLength()
				case *S3TableNamespaceOperator:
					S3TableNamespaceOperatorResourcesLength += operator.GetResourcesLength()
				case *S3VectorBucketOperator:
					s3VectorBucketOperatorResourcesLength += operator.GetResourcesLength()
				case *IamGroupOperator:
					iamGroupOperatorResourcesLength += operator.GetResourcesLength()
				case *EcrRepositoryOperator:
					ecrRepositoryOperatorResourcesLength += operator.GetResourcesLength()
				case *BackupVaultOperator:
					backupVaultOperatorResourcesLength += operator.GetResourcesLength()
				case *CloudFormationStackOperator:
					cloudformationStackOperatorResourcesLength += operator.GetResourcesLength()
				case *CustomOperator:
					customOperatorResourcesLength += operator.GetResourcesLength()
				default:
				}
			}

			got := want{
				logicalResourceIdsLength:                   len(operatorCollection.logicalResourceIds),
				unsupportedStackResourcesLength:            len(operatorCollection.unsupportedStackResources),
				s3BucketOperatorResourcesLength:            s3BucketOperatorResourcesLength,
				s3DirectoryBucketOperatorResourcesLength:   s3DirectoryBucketOperatorResourcesLength,
				s3TableBucketOperatorResourcesLength:       s3TableBucketOperatorResourcesLength,
				S3TableNamespaceOperatorResourcesLength:    S3TableNamespaceOperatorResourcesLength,
				s3VectorBucketOperatorResourcesLength:      s3VectorBucketOperatorResourcesLength,
				iamGroupOperatorResourcesLength:            iamGroupOperatorResourcesLength,
				ecrRepositoryOperatorResourcesLength:       ecrRepositoryOperatorResourcesLength,
				backupVaultOperatorResourcesLength:         backupVaultOperatorResourcesLength,
				cloudformationStackOperatorResourcesLength: cloudformationStackOperatorResourcesLength,
				customOperatorResourcesLength:              customOperatorResourcesLength,
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestOperatorCollection_SetOperatorCollection_MultipleCallsResetState(t *testing.T) {
	io.NewLogger(false)

	config := aws.Config{}
	operatorFactory := NewOperatorFactory(config)
	targetResourceTypes := []string{
		"AWS::IAM::Group",
		"AWS::S3::Bucket",
	}
	operatorCollection := NewOperatorCollection(config, operatorFactory, targetResourceTypes)

	stackName := aws.String("test-stack")

	// First call with 2 DELETE_FAILED resources
	operatorCollection.SetOperatorCollection(stackName, []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("IamGroup1"),
			PhysicalResourceId: aws.String("test-group-1"),
			ResourceType:       aws.String("AWS::IAM::Group"),
			ResourceStatus:     "DELETE_FAILED",
		},
		{
			LogicalResourceId:  aws.String("Bucket1"),
			PhysicalResourceId: aws.String("test-bucket-1"),
			ResourceType:       aws.String("AWS::S3::Bucket"),
			ResourceStatus:     "DELETE_FAILED",
		},
	})

	// Second call with 1 DELETE_FAILED resource (simulating loop iteration)
	operatorCollection.SetOperatorCollection(stackName, []types.StackResourceSummary{
		{
			LogicalResourceId:  aws.String("Bucket2"),
			PhysicalResourceId: aws.String("test-bucket-2"),
			ResourceType:       aws.String("AWS::S3::Bucket"),
			ResourceStatus:     "DELETE_FAILED",
		},
	})

	// Verify second call results - should be reset, not accumulated
	logicalIds := operatorCollection.GetLogicalResourceIds()
	if len(logicalIds) != 1 {
		t.Errorf("Second call: expected 1 logical resource ID, got %d (IDs: %v)", len(logicalIds), logicalIds)
	}
	if len(logicalIds) == 1 && logicalIds[0] != "Bucket2" {
		t.Errorf("Second call: expected logical resource ID 'Bucket2', got '%s'", logicalIds[0])
	}

	operators := operatorCollection.GetOperators()
	totalResourcesSecondCall := 0
	for _, op := range operators {
		totalResourcesSecondCall += op.GetResourcesLength()
	}
	if totalResourcesSecondCall != 1 {
		t.Errorf("Second call: expected 1 total resource in operators (reset, not accumulated), got %d", totalResourcesSecondCall)
	}
}

func TestOperatorCollection_containsResourceType(t *testing.T) {
	io.NewLogger(false)

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
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::S3::Bucket",
			},
			want: true,
		},
		{
			name: "S3 Directory Bucket for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::S3Express::DirectoryBucket",
			},
			want: true,
		},
		{
			name: "S3 Table Bucket for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::S3Tables::TableBucket",
			},
			want: true,
		},
		{
			name: "S3 Table Namespace for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::S3Tables::Namespace",
			},
			want: true,
		},
		{
			name: "S3 Vector Bucket for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::S3Vectors::VectorBucket",
			},
			want: true,
		},
		{
			name: "IAM Group for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::IAM::Group",
			},
			want: true,
		},
		{
			name: "ECR Repository for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::ECR::Repository",
			},
			want: true,
		},
		{
			name: "BACKUP VAULT for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::Backup::BackupVault",
			},
			want: true,
		},
		{
			name: "CloudFormation Stack for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "AWS::CloudFormation::Stack",
			},
			want: true,
		},
		{
			name: "custom resource for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
				resource:            "Custom::Abc",
			},
			want: true,
		},
		{
			name: "unsupported resource for all target resource types",
			args: args{
				ctx:                 context.Background(),
				stackName:           aws.String("test"),
				targetResourceTypes: targetResourceTypesForAllServices,
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
