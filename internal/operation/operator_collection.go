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
}

func NewOperatorCollection(config aws.Config, operatorFactory *OperatorFactory) *OperatorCollection {
	return &OperatorCollection{
		operatorFactory: operatorFactory,
	}
}

func (c *OperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
	c.stackName = aws.ToString(stackName)

	// Reset for each cloudformation delete stack loop
	c.logicalResourceIds = []string{}
	c.unsupportedStackResources = []types.StackResourceSummary{}
	c.operators = []IOperator{}

	s3BucketOperator := c.operatorFactory.CreateS3BucketOperator()
	s3DirectoryBucketOperator := c.operatorFactory.CreateS3DirectoryBucketOperator()
	s3TableBucketOperator := c.operatorFactory.CreateS3TableBucketOperator()
	S3TableNamespaceOperator := c.operatorFactory.CreateS3TableNamespaceOperator()
	s3VectorBucketOperator := c.operatorFactory.CreateS3VectorBucketOperator()
	iamGroupOperator := c.operatorFactory.CreateIamGroupOperator()
	ecrRepositoryOperator := c.operatorFactory.CreateEcrRepositoryOperator()
	backupVaultOperator := c.operatorFactory.CreateBackupVaultOperator()
	cloudformationStackOperator := c.operatorFactory.CreateCloudFormationStackOperator()
	customOperator := c.operatorFactory.CreateCustomOperator()

	for _, resource := range stackResourceSummaries {
		if resource.ResourceStatus != "DELETE_FAILED" {
			continue
		}

		c.logicalResourceIds = append(c.logicalResourceIds, aws.ToString(resource.LogicalResourceId))

		if !c.containsResourceType(*resource.ResourceType) {
			c.unsupportedStackResources = append(c.unsupportedStackResources, resource)
		} else {
			switch *resource.ResourceType {
			case resourcetype.S3Bucket:
				s3BucketOperator.AddResource(&resource)
			case resourcetype.S3DirectoryBucket:
				s3DirectoryBucketOperator.AddResource(&resource)
			case resourcetype.S3TableBucket:
				s3TableBucketOperator.AddResource(&resource)
			case resourcetype.S3TableNamespace:
				S3TableNamespaceOperator.AddResource(&resource)
			case resourcetype.S3VectorBucket:
				s3VectorBucketOperator.AddResource(&resource)
			case resourcetype.IamGroup:
				iamGroupOperator.AddResource(&resource)
			case resourcetype.EcrRepository:
				ecrRepositoryOperator.AddResource(&resource)
			case resourcetype.BackupVault:
				backupVaultOperator.AddResource(&resource)
			case resourcetype.CloudformationStack:
				cloudformationStackOperator.AddResource(&resource)
			default:
				if strings.Contains(*resource.ResourceType, resourcetype.CustomResource) {
					customOperator.AddResource(&resource)
				}
			}
		}
	}

	c.operators = append(c.operators, s3BucketOperator)
	c.operators = append(c.operators, s3DirectoryBucketOperator)
	c.operators = append(c.operators, s3TableBucketOperator)
	c.operators = append(c.operators, S3TableNamespaceOperator)
	c.operators = append(c.operators, s3VectorBucketOperator)
	c.operators = append(c.operators, iamGroupOperator)
	c.operators = append(c.operators, ecrRepositoryOperator)
	c.operators = append(c.operators, backupVaultOperator)
	c.operators = append(c.operators, cloudformationStackOperator)
	c.operators = append(c.operators, customOperator)
}

func (c *OperatorCollection) containsResourceType(resource string) bool {
	for _, t := range resourcetype.ResourceTypes {
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

	unsupportedTable, err := io.ToStringAsTableFormat(unsupportedStackResourcesHeader, unsupportedStackResourcesData)
	if err != nil {
		return fmt.Errorf("UnsupportedResourceError: failed to create unsupported resources table, %w", err)
	}
	unsupportedStackResources := "\nThese are the resources unsupported, so failed delete:\n" + *unsupportedTable

	supportedStackResourcesHeader := []string{"ResourceType", "Description"}
	supportedStackResourcesData := [][]string{
		{resourcetype.S3Bucket, "S3 Buckets, including buckets with Non-empty or Versioning enabled and DeletionPolicy not Retain."},
		{resourcetype.S3DirectoryBucket, "S3 Directory Buckets for S3 Express One Zone, including buckets with Non-empty and DeletionPolicy not Retain."},
		{resourcetype.S3TableBucket, "S3 Table Buckets, including buckets with any namespaces or tables and DeletionPolicy not Retain."},
		{resourcetype.S3TableNamespace, "S3 Table Namespaces, including namespaces with any tables and DeletionPolicy not Retain."},
		{resourcetype.S3VectorBucket, "S3 Vector Buckets, including buckets with any indexes and DeletionPolicy not Retain."},
		{resourcetype.IamGroup, "IAM Groups, including groups with IAM users from outside the stack."},
		{resourcetype.EcrRepository, "ECR Repositories, including repositories that contain images and where the `EmptyOnDelete` is not true."},
		{resourcetype.BackupVault, "Backup Vaults, including vaults containing recovery points."},
		{resourcetype.CloudformationStack, "Nested Child Stacks that failed to delete."},
		{"Custom::Xxx", "Custom Resources, including resources that do not return a SUCCESS status."},
	}

	supportedTable, err := io.ToStringAsTableFormat(supportedStackResourcesHeader, supportedStackResourcesData)
	if err != nil {
		return fmt.Errorf("UnsupportedResourceError: failed to create supported resources table, %w", err)
	}
	supportedStackResources := "\nSupported resources for force deletion of DELETE_FAILED resources are followings.\n" + *supportedTable

	issueLink := "\nIf you want to delete the unsupported resources, please create an issue at GitHub(https://github.com/go-to-k/delstack/issues).\n"

	unsupportedResourceError := title + unsupportedStackResources + supportedStackResources + issueLink

	return fmt.Errorf("UnsupportedResourceError: %v", unsupportedResourceError)
}
