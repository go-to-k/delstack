package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
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
	operatorFactory           IOperatorFactory
	logicalResourceIds        []string
	unsupportedStackResources []types.StackResourceSummary
	operators                 []IOperator
}

func NewOperatorCollection(config aws.Config, operatorFactory IOperatorFactory) *OperatorCollection {
	return &OperatorCollection{
		operatorFactory: operatorFactory,
	}
}

func (operatorCollection *OperatorCollection) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
	operatorCollection.stackName = aws.ToString(stackName)

	stackOperator := operatorCollection.operatorFactory.CreateStackOperator()
	bucketOperator := operatorCollection.operatorFactory.CreateBucketOperator()
	roleOperator := operatorCollection.operatorFactory.CreateRoleOperator()
	ecrOperator := operatorCollection.operatorFactory.CreateEcrOperator()
	backupVaultOperator := operatorCollection.operatorFactory.CreateBackupVaultOperator()
	customOperator := operatorCollection.operatorFactory.CreateCustomOperator()

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			stackResource := v // Copy for pointer used below
			operatorCollection.logicalResourceIds = append(operatorCollection.logicalResourceIds, aws.ToString(stackResource.LogicalResourceId))

			switch *stackResource.ResourceType {
			case resourcetype.CLOUDFORMATION_STACK:
				stackOperator.AddResource(&stackResource)
			case resourcetype.S3_STACK:
				bucketOperator.AddResource(&stackResource)
			case resourcetype.IAM_ROLE:
				roleOperator.AddResource(&stackResource)
			case resourcetype.ECR_REPOSITORY:
				ecrOperator.AddResource(&stackResource)
			case resourcetype.BACKUP_VAULT:
				backupVaultOperator.AddResource(&stackResource)
			default:
				if strings.Contains(*stackResource.ResourceType, resourcetype.CUSTOM_RESOURCE) {
					customOperator.AddResource(&stackResource)
				} else {
					operatorCollection.unsupportedStackResources = append(operatorCollection.unsupportedStackResources, stackResource)
				}
			}
		}
	}

	operatorCollection.operators = append(operatorCollection.operators, stackOperator)
	operatorCollection.operators = append(operatorCollection.operators, bucketOperator)
	operatorCollection.operators = append(operatorCollection.operators, roleOperator)
	operatorCollection.operators = append(operatorCollection.operators, ecrOperator)
	operatorCollection.operators = append(operatorCollection.operators, backupVaultOperator)
	operatorCollection.operators = append(operatorCollection.operators, customOperator)
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() []string {
	return operatorCollection.logicalResourceIds
}

func (operatorCollection *OperatorCollection) GetOperators() []IOperator {
	return operatorCollection.operators
}

func (operatorCollection *OperatorCollection) RaiseUnsupportedResourceError() error {
	title := fmt.Sprintf("%v deletion is FAILED !!!\n", operatorCollection.stackName)

	unsupportedStackResourcesHeader := []string{"ResourceType", "Resource"}
	unsupportedStackResourcesData := [][]string{}

	for _, resource := range operatorCollection.unsupportedStackResources {
		unsupportedStackResourcesData = append(unsupportedStackResourcesData, []string{*resource.ResourceType, *resource.LogicalResourceId})
	}
	unsupportedStackResources := "\nThese are unsupported resources so failed delete:\n" + *logger.ToStringAsTableFormat(unsupportedStackResourcesHeader, unsupportedStackResourcesData)

	supportedStackResourcesHeader := []string{"ResourceType", "Description"}
	supportedStackResourcesData := [][]string{
		{resourcetype.S3_STACK, "S3 Buckets, including buckets with Non-empty or Versioning enabled and DeletionPolicy not Retain."},
		{resourcetype.IAM_ROLE, "IAM Roles, including roles with policies from outside the stack."},
		{resourcetype.ECR_REPOSITORY, "ECR Repositories, including repositories containing images."},
		{resourcetype.BACKUP_VAULT, "Backup Vaults, including vaults containing recovery points."},
		{resourcetype.CLOUDFORMATION_STACK, "Nested Child Stacks that failed to delete."},
		{"Custom::Xxx", "Custom Resources, but they will be deleted on its own."},
	}
	supportedStackResources := "\nSupported resources for force deletion of DELETE_FAILED resources are followings.\n" + *logger.ToStringAsTableFormat(supportedStackResourcesHeader, supportedStackResourcesData)

	unsupportedResourceError := title + unsupportedStackResources + supportedStackResources

	return fmt.Errorf("UnsupportedResourceError: %v", unsupportedResourceError)
}
