//go:generate mockgen -source=$GOFILE -destination=operator_collection_mock.go -package=$GOPACKAGE -write_package_comment=false
package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
)

type IOperatorCollection interface {
	SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary)
	GetLogicalResourceIds() []string
	GetOperators() []IOperator
	RaiseUnsupportedResourceError() error
}

var _ IOperatorCollection = (*OperatorCollection)(nil)

type OperatorCollection struct {
	stackName                 string
	operatorFactory           *OperatorFactory
	logicalResourceIds        []string
	unsupportedStackResources []types.StackResourceSummary
	operators                 []IOperator
	targetResourceTypes       []string
}

func NewOperatorCollection(config aws.Config, operatorFactory *OperatorFactory, targetResourceTypes []string) *OperatorCollection {
	return &OperatorCollection{
		operatorFactory:     operatorFactory,
		targetResourceTypes: targetResourceTypes,
	}
}

func (c *OperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
	c.stackName = aws.ToString(stackName)

	s3BucketOperator := c.operatorFactory.CreateS3BucketOperator()
	s3DirectoryBucketOperator := c.operatorFactory.CreateS3DirectoryBucketOperator()
	s3TableBucketOperator := c.operatorFactory.CreateS3TableBucketOperator()
	iamGroupOperator := c.operatorFactory.CreateIamGroupOperator()
	ecrRepositoryOperator := c.operatorFactory.CreateEcrRepositoryOperator()
	backupVaultOperator := c.operatorFactory.CreateBackupVaultOperator()
	cloudformationStackOperator := c.operatorFactory.CreateCloudFormationStackOperator(c.targetResourceTypes)
	customOperator := c.operatorFactory.CreateCustomOperator()

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			stackResource := v // Copy for pointer used below
			c.logicalResourceIds = append(c.logicalResourceIds, aws.ToString(stackResource.LogicalResourceId))

			if !c.containsResourceType(*stackResource.ResourceType) {
				c.unsupportedStackResources = append(c.unsupportedStackResources, stackResource)
			} else {
				switch *stackResource.ResourceType {
				case resourcetype.S3Bucket:
					s3BucketOperator.AddResource(&stackResource)
				case resourcetype.S3DirectoryBucket:
					s3DirectoryBucketOperator.AddResource(&stackResource)
				case resourcetype.S3TableBucket:
					s3TableBucketOperator.AddResource(&stackResource)
				case resourcetype.IamGroup:
					iamGroupOperator.AddResource(&stackResource)
				case resourcetype.EcrRepository:
					ecrRepositoryOperator.AddResource(&stackResource)
				case resourcetype.BackupVault:
					backupVaultOperator.AddResource(&stackResource)
				case resourcetype.CloudformationStack:
					cloudformationStackOperator.AddResource(&stackResource)
				default:
					if strings.Contains(*stackResource.ResourceType, resourcetype.CustomResource) {
						customOperator.AddResource(&stackResource)
					}
				}
			}
		}
	}

	c.operators = append(c.operators, s3BucketOperator)
	c.operators = append(c.operators, s3DirectoryBucketOperator)
	c.operators = append(c.operators, s3TableBucketOperator)
	c.operators = append(c.operators, iamGroupOperator)
	c.operators = append(c.operators, ecrRepositoryOperator)
	c.operators = append(c.operators, backupVaultOperator)
	c.operators = append(c.operators, cloudformationStackOperator)
	c.operators = append(c.operators, customOperator)
}

func (c *OperatorCollection) containsResourceType(resource string) bool {
	for _, t := range c.targetResourceTypes {
		if t == resource || (t == resourcetype.CustomResource && strings.Contains(resource, resourcetype.CustomResource)) {
			return true
		}
	}
	return false
}

func (c *OperatorCollection) GetLogicalResourceIds() []string {
	return c.logicalResourceIds
}

func (c *OperatorCollection) GetOperators() []IOperator {
	return c.operators
}

func (c *OperatorCollection) RaiseUnsupportedResourceError() error {
	title := fmt.Sprintf("%v deletion is FAILED !!!\n", c.stackName)

	unsupportedStackResourcesHeader := []string{"ResourceType", "Resource"}
	unsupportedStackResourcesData := [][]string{}

	for _, resource := range c.unsupportedStackResources {
		unsupportedStackResourcesData = append(unsupportedStackResourcesData, []string{*resource.ResourceType, *resource.LogicalResourceId})
	}
	unsupportedStackResources := "\nThese are the resources unsupported (or you did not selected in the interactive prompt), so failed delete:\n" + *io.ToStringAsTableFormat(unsupportedStackResourcesHeader, unsupportedStackResourcesData)

	supportedStackResourcesHeader := []string{"ResourceType", "Description"}
	supportedStackResourcesData := [][]string{
		{resourcetype.S3Bucket, "S3 Buckets, including buckets with Non-empty or Versioning enabled and DeletionPolicy not Retain."},
		{resourcetype.S3DirectoryBucket, "S3 Directory Buckets for S3 Express One Zone, including buckets with Non-empty and DeletionPolicy not Retain."},
		{resourcetype.S3TableBucket, "S3 Table Buckets, including buckets with any namespaces or tables and DeletionPolicy not Retain."},
		{resourcetype.IamGroup, "IAM Groups, including groups with IAM users from outside the stack."},
		{resourcetype.EcrRepository, "ECR Repositories, including repositories that contain images and where the `EmptyOnDelete` is not true."},
		{resourcetype.BackupVault, "Backup Vaults, including vaults containing recovery points."},
		{resourcetype.CloudformationStack, "Nested Child Stacks that failed to delete."},
		{"Custom::Xxx", "Custom Resources, including resources that do not return a SUCCESS status."},
	}
	supportedStackResources := "\nSupported resources for force deletion of DELETE_FAILED resources are followings.\n" + *io.ToStringAsTableFormat(supportedStackResourcesHeader, supportedStackResourcesData)

	issueLink := "\nIf you want to delete the unsupported resources, please create an issue at GitHub(https://github.com/go-to-k/delstack/issues).\n"

	unsupportedResourceError := title + unsupportedStackResources + supportedStackResources + issueLink

	return fmt.Errorf("UnsupportedResourceError: %v", unsupportedResourceError)
}
