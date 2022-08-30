package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type ResourceCollection struct {
	config              aws.Config
	StackName           string
	LogicalResourceIds  []string
	StackOperator       *StackOperator
	BucketOperator      *BucketOperator
	RoleOperator        *RoleOperator
	ECROperator         *ECROperator
	BackupVaultOperator *BackupVaultOperator
	CustomOperator      *CustomOperator
}

func NewResourceCollection(config aws.Config, stackName string, stackResourceSummaries []types.StackResourceSummary) *ResourceCollection {
	var logicalResourceIds []string
	stackOperator := NewStackOperator(config)
	bucketOperator := NewBucketOperator(config)
	roleOperator := NewRoleOperator(config)
	ecrOperator := NewECROperator(config)
	backupVaultOperator := NewBackupVaultOperator(config)
	customOperator := NewCustomOperator(config)

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			// elseでremoveでも？
			// それかcountでintでも？
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

	return &ResourceCollection{
		config:              config,
		StackName:           stackName,
		LogicalResourceIds:  logicalResourceIds,
		StackOperator:       stackOperator,
		BucketOperator:      bucketOperator,
		RoleOperator:        roleOperator,
		ECROperator:         ecrOperator,
		BackupVaultOperator: backupVaultOperator,
		CustomOperator:      customOperator,
	}
}

func (collection *ResourceCollection) CheckResourceCounts() error {
	collectionLength := collection.StackOperator.GetResourcesLength() +
		collection.BucketOperator.GetResourcesLength() +
		collection.RoleOperator.GetResourcesLength() +
		collection.ECROperator.GetResourcesLength() +
		collection.BackupVaultOperator.GetResourcesLength() +
		collection.CustomOperator.GetResourcesLength()

	if len(collection.LogicalResourceIds) != collectionLength {
		fmt.Println("===========================================================")
		fmt.Printf("%v is FAILED !!!", collection.StackName)
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

func (collection *ResourceCollection) DeleteResourceCollection() error {
	// TODO: Concurrency deletion of failed resources
	if err := collection.StackOperator.DeleteResources(); err != nil {
		return err
	}
	if err := collection.BucketOperator.DeleteResources(); err != nil {
		return err
	}
	if err := collection.RoleOperator.DeleteResources(); err != nil {
		return err
	}
	if err := collection.ECROperator.DeleteResources(); err != nil {
		return err
	}
	if err := collection.BackupVaultOperator.DeleteResources(); err != nil {
		return err
	}
	if err := collection.CustomOperator.DeleteResources(); err != nil {
		return err
	}

	return nil
}
