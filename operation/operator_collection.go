package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/logger"
)

type OperatorCollection struct {
	stackName                  string
	logicalResourceIds         []string
	notSupportedStackResources []types.StackResourceSummary
	operatorList               []Operator
}

func NewOperatorCollection(config aws.Config, stackName *string, stackResourceSummaries []types.StackResourceSummary) *OperatorCollection {
	logicalResourceIds := []string{}
	notSupportedStackResources := []types.StackResourceSummary{}
	stackOperator := NewStackOperator(config)
	bucketOperator := NewBucketOperator(config)
	roleOperator := NewRoleOperator(config)
	ecrOperator := NewECROperator(config)
	backupVaultOperator := NewBackupVaultOperator(config)
	customOperator := NewCustomOperator() // Implicit instances that do not actually delete resources

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
					notSupportedStackResources = append(notSupportedStackResources, stackResource)
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
		stackName:                  aws.ToString(stackName),
		logicalResourceIds:         logicalResourceIds,
		notSupportedStackResources: notSupportedStackResources,
		operatorList:               operatorList,
	}
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() []string {
	return operatorCollection.logicalResourceIds
}

func (operatorCollection *OperatorCollection) GetOperatorList() []Operator {
	return operatorCollection.operatorList
}

func (operatorCollection *OperatorCollection) RaiseNotSupportedResourceError() error {
	title := fmt.Sprintf("%v deletion is FAILED !!!\n", operatorCollection.stackName)

	notSupportedStackResourcesHeader := []string{"ResourceType", "Resource"}
	notSupportedStackResourcesData := [][]string{}

	for _, resource := range operatorCollection.notSupportedStackResources {
		notSupportedStackResourcesData = append(notSupportedStackResourcesData, []string{*resource.ResourceType, *resource.LogicalResourceId})
	}
	notSupportedStackResources := "\nThese are not supported resources so failed delete:\n" + *logger.ToStringAsTableFormat(notSupportedStackResourcesHeader, notSupportedStackResourcesData)

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

	notSupportedResourceError := title + notSupportedStackResources + supportedStackResources

	return fmt.Errorf("NotSupportedResourceError: %v", notSupportedResourceError)
}
