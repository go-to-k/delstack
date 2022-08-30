package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type OperatorCollection struct {
	config             aws.Config
	StackName          string
	LogicalResourceIds []string
	OperatorList       []IOperator
}

func NewOperatorCollection(config aws.Config, stackName string, stackResourceSummaries []types.StackResourceSummary) *OperatorCollection {
	logicalResourceIds := []string{}
	stackOperator := NewStackOperator(config)
	bucketOperator := NewBucketOperator(config)
	roleOperator := NewRoleOperator(config)
	ecrOperator := NewECROperator(config)
	backupVaultOperator := NewBackupVaultOperator(config)
	customOperator := NewCustomOperator(config)

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			logicalResourceIds = append(logicalResourceIds, *v.LogicalResourceId)

			switch *v.ResourceType {
			case "AWS::CloudFormation::Stack":
				stackOperator.AddResources(v)
			case "AWS::S3::Bucket":
				bucketOperator.AddResources(v)
			case "AWS::IAM::Role":
				roleOperator.AddResources(v)
			case "AWS::ECR::Repository":
				ecrOperator.AddResources(v)
			case "AWS::Backup::BackupVault":
				backupVaultOperator.AddResources(v)
			default:
				if strings.Contains(*v.ResourceType, "Custom::") {
					customOperator.AddResources(v)
				}
			}
		}
	}

	var operatorList []IOperator
	operatorList = append(operatorList, stackOperator)
	operatorList = append(operatorList, bucketOperator)
	operatorList = append(operatorList, roleOperator)
	operatorList = append(operatorList, ecrOperator)
	operatorList = append(operatorList, backupVaultOperator)
	operatorList = append(operatorList, customOperator)

	return &OperatorCollection{
		config:             config,
		StackName:          stackName,
		LogicalResourceIds: logicalResourceIds,
		OperatorList:       operatorList,
	}
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() *[]string {
	return &operatorCollection.LogicalResourceIds
}

func (operatorCollection *OperatorCollection) getResourcesLengthFromOperatorList() int {
	var length int
	for _, operator := range operatorCollection.OperatorList {
		length += operator.GetResourcesLength()
	}
	return length
}

func (operatorCollection *OperatorCollection) CheckResourceCounts() error {
	collectionLength := operatorCollection.getResourcesLengthFromOperatorList()

	if len(operatorCollection.LogicalResourceIds) != collectionLength {
		fmt.Println("===========================================================")
		fmt.Printf("%v is FAILED !!!", operatorCollection.StackName)
		fmt.Println("")
		fmt.Println("The deletion seems to be failing for some other reason.")
		fmt.Println("This function supports force deletion of ")
		fmt.Println("<S3 buckets> that are Non-empty or Versioning enabled")
		fmt.Println("and <IAM roles> with policies attached from outside the stack,")
		fmt.Println("and <ECR> still contains images,")
		fmt.Println("and <BackupVault> contains recovery points,")
		fmt.Println("and <Nested Child Stack>.")
		fmt.Println("<Custom Resources> was also forced to delete.")
		fmt.Println("===========================================================")
		fmt.Println("")

		return fmt.Errorf("not supported services error")
	}

	return nil
}

func (operatorCollection *OperatorCollection) DeleteResourceCollection() error {
	// TODO: Concurrency deletion of failed resources
	for _, operator := range operatorCollection.OperatorList {
		if err := operator.DeleteResources(); err != nil {
			return err
		}
	}

	return nil
}
