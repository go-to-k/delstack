//go:generate mockgen -source=./operator_collection.go -destination=./operator_collection_mock.go -package=operation -write_package_comment=false
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
	iamRoleOperator := c.operatorFactory.CreateIamRoleOperator()
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
				case resourcetype.IamRole:
					iamRoleOperator.AddResource(&stackResource)
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
	c.operators = append(c.operators, iamRoleOperator)
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
		{resourcetype.IamRole, "IAM Roles, including roles with policies from outside the stack."},
		{resourcetype.EcrRepository, "ECR Repositories, including repositories containing images."},
		{resourcetype.BackupVault, "Backup Vaults, including vaults containing recovery points."},
		{resourcetype.CloudformationStack, "Nested Child Stacks that failed to delete."},
		{"Custom::Xxx", "Custom Resources, but they will be deleted on its own."},
	}
	supportedStackResources := "\nSupported resources for force deletion of DELETE_FAILED resources are followings.\n" + *io.ToStringAsTableFormat(supportedStackResourcesHeader, supportedStackResourcesData)

	unsupportedResourceError := title + unsupportedStackResources + supportedStackResources

	return fmt.Errorf("UnsupportedResourceError: %v", unsupportedResourceError)
}
