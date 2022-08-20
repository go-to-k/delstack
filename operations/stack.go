package operations

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
)

func DeleteStackResources(config aws.Config, stackName string) error {
	cfnClient := client.NewCloudFormation(config)

	stackOutputBeforeDelete, isExistBeforeDelete, err := cfnClient.DescribeStacks(&stackName)
	if err != nil {
		return err
	}
	if !isExistBeforeDelete {
		fmt.Println("The stack is not exists")
		return err
	}

	if *stackOutputBeforeDelete.Stacks[0].EnableTerminationProtection {
		fmt.Println("TerminationProtection is enabled")
		return nil
	}

	if err := cfnClient.DeleteStack(&stackName, []string{}); err != nil {
		return err
	}

	stackOutputAfterDelete, isExistAfterDelete, err := cfnClient.DescribeStacks(&stackName)
	if err != nil {
		return err
	}
	if !isExistAfterDelete {
		fmt.Println("Successfully deleted without failed resources")
		return nil
	}
	if stackOutputAfterDelete.Stacks[0].StackStatus != "DELETE_FAILED" {
		log.Fatalf("StackStatus is expected to be DELETE_FAILED, but %s", stackOutputAfterDelete.Stacks[0].StackStatus)
		return err
	}

	stackResources, err := cfnClient.ListStackResources(&stackName)
	if err != nil {
		return err
	}

	var logicalResourceIdsForRetainResources []string
	var (
		stackArray  []types.StackResourceSummary
		bucketArray []types.StackResourceSummary
		roleArray   []types.StackResourceSummary
		ecrArray    []types.StackResourceSummary
		backupArray []types.StackResourceSummary
		customArray []types.StackResourceSummary
	)

	for _, v := range stackResources.StackResourceSummaries {
		if v.ResourceStatus == "DELETE_FAILED" {
			logicalResourceIdsForRetainResources = append(logicalResourceIdsForRetainResources, *v.LogicalResourceId)

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

	if len(logicalResourceIdsForRetainResources) != len(stackArray)+len(bucketArray)+len(roleArray)+len(ecrArray)+len(ecrArray)+len(backupArray)+len(customArray) {
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
	}

	// TODO: remove
	fmt.Printf("%v", len(logicalResourceIdsForRetainResources))
	fmt.Printf("%v", len(stackArray))
	fmt.Printf("%v", len(bucketArray))
	fmt.Printf("%v", len(roleArray))
	fmt.Printf("%v", len(ecrArray))
	fmt.Printf("%v", len(backupArray))
	fmt.Printf("%v", len(customArray))

	if err := cfnClient.DeleteStack(&stackName, logicalResourceIdsForRetainResources); err != nil {
		return err
	}

	return nil
}
