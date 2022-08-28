package operation

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type ResourceCollection struct {
	config             aws.Config
	StackName          string
	LogicalResourceIds []string
	StackArray         []types.StackResourceSummary
	BucketArray        []types.StackResourceSummary
	RoleArray          []types.StackResourceSummary
	ECRArray           []types.StackResourceSummary
	BackupArray        []types.StackResourceSummary
	CustomArray        []types.StackResourceSummary
}

func NewResourceCollection(config aws.Config, stackName string, stackResourceSummaries []types.StackResourceSummary) *ResourceCollection {
	var logicalResourceIds []string
	var (
		stackArray  []types.StackResourceSummary
		bucketArray []types.StackResourceSummary
		roleArray   []types.StackResourceSummary
		ecrArray    []types.StackResourceSummary
		backupArray []types.StackResourceSummary
		customArray []types.StackResourceSummary
	)

	for _, v := range stackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			logicalResourceIds = append(logicalResourceIds, *v.LogicalResourceId)

			switch *v.ResourceType {
			case "AWS::CloudFormation::Stack":
				stackArray = append(stackArray, v)
			case "AWS::S3::Bucket":
				bucketArray = append(bucketArray, v)
			case "AWS::IAM::Role":
				roleArray = append(roleArray, v)
			case "AWS::ECR::Repository":
				ecrArray = append(ecrArray, v)
			case "AWS::Backup::BackupVault":
				backupArray = append(backupArray, v)
			default:
				if strings.Contains(*v.ResourceType, "Custom::") {
					customArray = append(customArray, v)
				}
			}
		}
	}

	return &ResourceCollection{
		config:             config,
		StackName:          stackName,
		LogicalResourceIds: logicalResourceIds,
		StackArray:         stackArray,
		BucketArray:        bucketArray,
		RoleArray:          roleArray,
		ECRArray:           ecrArray,
		BackupArray:        backupArray,
		CustomArray:        customArray,
	}
}

func (collection *ResourceCollection) CheckResourceCounts() error {
	collectionLength := len(collection.StackArray) +
		len(collection.BucketArray) +
		len(collection.RoleArray) +
		len(collection.ECRArray) +
		len(collection.BackupArray) +
		len(collection.CustomArray)

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

	if err := DeleteStacks(collection.config, collection.StackArray); err != nil {
		return err
	}
	if err := DeleteBuckets(collection.config, collection.BucketArray); err != nil {
		return err
	}
	if err := DeleteRoles(collection.config, collection.RoleArray); err != nil {
		return err
	}
	if err := DeleteECRs(collection.config, collection.ECRArray); err != nil {
		return err
	}
	if err := DeleteBackupVaults(collection.config, collection.ECRArray); err != nil {
		return err
	}
	if err := DeleteCustoms(collection.config, collection.CustomArray); err != nil {
		return err
	}

	return nil
}
