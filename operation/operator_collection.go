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
	operatorList       []Operator
}

func NewOperatorCollection(config aws.Config, stackName *string, stackResourceSummaries []types.StackResourceSummary) *OperatorCollection {
	logicalResourceIds := []string{}
	stackOperator := NewStackOperator(config)
	bucketOperator := NewBucketOperator(config)
	roleOperator := NewRoleOperator(config)
	ecrOperator := NewECROperator(config)
	backupVaultOperator := NewBackupVaultOperator(config)
	customOperator := NewCustomOperator() // Implicit instances that do not actually delete resources

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			stackResource := v // Copy for pointer used below
			logicalResourceIds = append(logicalResourceIds, aws.ToString(v.LogicalResourceId))

			switch *v.ResourceType {
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
		stackName:          aws.ToString(stackName),
		logicalResourceIds: logicalResourceIds,
		operatorList:       operatorList,
	}
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() []string {
	return operatorCollection.logicalResourceIds
}

func (operatorCollection *OperatorCollection) GetOperatorList() []Operator {
	return operatorCollection.operatorList
}

func (operatorCollection *OperatorCollection) RaiseNotSupportedServicesError() error {
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
	fmt.Println("<Custom Resources> will be deleted on its own.")
	fmt.Println("===========================================================")
	fmt.Println("")

	return fmt.Errorf("not supported services error")
}
