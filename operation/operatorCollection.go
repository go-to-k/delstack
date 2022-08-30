package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type OperatorCollection struct {
	stackName          string
	logicalResourceIds []string
	operatorList       []IOperator
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
		stackName:          stackName,
		logicalResourceIds: logicalResourceIds,
		operatorList:       operatorList,
	}
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() *[]string {
	return &operatorCollection.logicalResourceIds
}

func (operatorCollection *OperatorCollection) GetOperatorList() *[]IOperator {
	return &operatorCollection.operatorList
}

func (operatorCollection *OperatorCollection) GetNotSupportedServicesError() error {
	fmt.Println("===========================================================")
	fmt.Printf("%v is FAILED !!!", operatorCollection.stackName)
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
