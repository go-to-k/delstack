package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type OperatorCollection struct {
	LogicalResourceIds  []string
	StackOperator       *StackOperator
	BucketOperator      *BucketOperator
	RoleOperator        *RoleOperator
	ECROperator         *ECROperator
	BackupVaultOperator *BackupVaultOperator
	CustomOperator      *CustomOperator
}

func NewOperatorCollection(config aws.Config) *OperatorCollection {
	logicalResourceIds := []string{}
	stackOperator := NewStackOperator(config)
	bucketOperator := NewBucketOperator(config)
	roleOperator := NewRoleOperator(config)
	ecrOperator := NewECROperator(config)
	backupVaultOperator := NewBackupVaultOperator(config)
	customOperator := NewCustomOperator(config)

	return &OperatorCollection{
		LogicalResourceIds:  logicalResourceIds,
		StackOperator:       stackOperator,
		BucketOperator:      bucketOperator,
		RoleOperator:        roleOperator,
		ECROperator:         ecrOperator,
		BackupVaultOperator: backupVaultOperator,
		CustomOperator:      customOperator,
	}
}

func (operatorCollection *OperatorCollection) AddStackResources(stackResourceSummaries []types.StackResourceSummary) {
	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			// elseでremoveでも？
			// それかcountでintでも？
			operatorCollection.LogicalResourceIds = append(operatorCollection.LogicalResourceIds, *v.LogicalResourceId)

			switch *v.ResourceType {
			case "AWS::CloudFormation::Stack":
				operatorCollection.StackOperator.AddResources(v)
			case "AWS::S3::Bucket":
				operatorCollection.BucketOperator.AddResources(v)
			case "AWS::IAM::Role":
				operatorCollection.RoleOperator.AddResources(v)
			case "AWS::ECR::Repository":
				operatorCollection.ECROperator.AddResources(v)
			case "AWS::Backup::BackupVault":
				operatorCollection.BackupVaultOperator.AddResources(v)
			default:
				if strings.Contains(*v.ResourceType, "Custom::") {
					operatorCollection.CustomOperator.AddResources(v)
				}
			}
		}
	}
}

func (operatorCollection *OperatorCollection) GetLogicalResourceIds() *[]string {
	return &operatorCollection.LogicalResourceIds
}

func (operatorCollection *OperatorCollection) CheckResourceCounts(stackName string) error {
	collectionLength := operatorCollection.StackOperator.GetResourcesLength() +
		operatorCollection.BucketOperator.GetResourcesLength() +
		operatorCollection.RoleOperator.GetResourcesLength() +
		operatorCollection.ECROperator.GetResourcesLength() +
		operatorCollection.BackupVaultOperator.GetResourcesLength() +
		operatorCollection.CustomOperator.GetResourcesLength()

	if len(operatorCollection.LogicalResourceIds) != collectionLength {
		fmt.Println("===========================================================")
		fmt.Printf("%v is FAILED !!!", stackName)
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

func (operatorCollection *OperatorCollection) GetOperatorList() *[]IOperator {
	var operatorList []IOperator

	operatorList = append(operatorList, operatorCollection.StackOperator)
	operatorList = append(operatorList, operatorCollection.BucketOperator)
	operatorList = append(operatorList, operatorCollection.RoleOperator)
	operatorList = append(operatorList, operatorCollection.ECROperator)
	operatorList = append(operatorList, operatorCollection.BackupVaultOperator)
	operatorList = append(operatorList, operatorCollection.CustomOperator)

	return &operatorList
}
