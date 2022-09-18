package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/logger"
)

type OperatorCollection struct {
	stackName                 string
	logicalResourceIds        []string
	unsupportedStackResources []types.StackResourceSummary
	operatorList              []Operator
}

func NewOperatorCollection(config aws.Config, operatorFactory IOperatorFactory, stackName *string, stackResourceSummaries []types.StackResourceSummary) *OperatorCollection {
	logicalResourceIds := []string{}
	unsupportedStackResources := []types.StackResourceSummary{}

	stackOperator := operatorFactory.CreateStackOperator()
	bucketOperator := operatorFactory.CreateBucketOperator()
	roleOperator := operatorFactory.CreateRoleOperator()
	ecrOperator := operatorFactory.CreateEcrOperator()
	backupVaultOperator := operatorFactory.CreateBackupVaultOperator()
	customOperator := operatorFactory.CreateCustomOperator()

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			stackResource := v // Copy for pointer used below
			logicalResourceIds = append(logicalResourceIds, aws.ToString(stackResource.LogicalResourceId))

			switch *stackResource.ResourceType {
			case "AWS::CloudFormation::Stack":
				stackOperator.AddResources(&stackResource)
			case "AWS::S3::Bucket":
				bucketOperator.AddResources(&stackResource)
			case "AWS::IAM::Role":
				roleOperator.AddResources(&stackResource)
			case "AWS::ECR::Repository":
				ecrOperator.AddResources(&stackResource)
			case "AWS::Backup::BackupVault":
				backupVaultOperator.AddResources(&stackResource)
			default:
				if strings.Contains(*v.ResourceType, "Custom::") {
					customOperator.AddResources(&stackResource)
				} else {
					unsupportedStackResources = append(unsupportedStackResources, stackResource)
				}
			}
		}
	}

	var operatorList []Operator
	operatorList = append(operatorList, stackOperator)
	operatorList = append(operatorList, bucketOperator)
	operatorList = append(operatorList, roleOperator)
	operatorList = append(operatorList, ecrOperator)
	operatorList = append(operatorList, backupVaultOperator)
	operatorList = append(operatorList, customOperator)

	return &OperatorCollection{
		stackName:                 aws.ToString(stackName),
		logicalResourceIds:        logicalResourceIds,
		unsupportedStackResources: unsupportedStackResources,
		operatorList:              operatorList,
	}
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() []string {
	return operatorCollection.logicalResourceIds
}

func (operatorCollection *OperatorCollection) GetOperatorList() []Operator {
	return operatorCollection.operatorList
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
		{"AWS::S3::Bucket", "S3 Buckets, including buckets with Non-empty or Versioning enabled and DeletionPolicy not Retain."},
		{"AWS::IAM::Role", "IAM Roles, including roles with policies from outside the stack."},
		{"AWS::ECR::Repository", "ECR Repositories, including repositories containing images."},
		{"AWS::Backup::BackupVault", "Backup Vaults, including vaults containing recovery points."},
		{"AWS::CloudFormation::Stack", "Nested Child Stacks that failed to delete."},
		{"Custom::Xxx", "Custom Resources, but they will be deleted on its own."},
	}
	supportedStackResources := "\nSupported resources for force deletion of DELETE_FAILED resources are followings.\n" + *logger.ToStringAsTableFormat(supportedStackResourcesHeader, supportedStackResourcesData)

	unsupportedResourceError := title + unsupportedStackResources + supportedStackResources

	return fmt.Errorf("UnsupportedResourceError: %v", unsupportedResourceError)
}
