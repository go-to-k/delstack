package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/option"
	"github.com/go-to-k/delstack/shared"
	flags "github.com/jessevdk/go-flags"
)

var opts option.Option

// TODO: EXITまわり統一（異常と正常でどうするか）
func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		// os.Exit(1)
		return
	}

	cfg, err := shared.LoadAwsConfig(opts.Profile)
	if err != nil {
		os.Exit(1)
	}

	cfnClient := shared.NewCloudFormation(cfg)

	stackOutputBeforeDelete, isExistBeforeDelete, err := cfnClient.DescribeStacks(&opts.StackName)
	if err != nil {
		os.Exit(1)
	}
	if !isExistBeforeDelete {
		fmt.Println("The stack is not exists")
		os.Exit(1)
	}

	if *stackOutputBeforeDelete.Stacks[0].EnableTerminationProtection {
		fmt.Println("TerminationProtection is enabled")
		return
	}

	if err := cfnClient.DeleteStack(&opts.StackName); err != nil {
		os.Exit(1)
	}

	stackOutputAfterDelete, isExistAfterDelete, err := cfnClient.DescribeStacks(&opts.StackName)
	if err != nil {
		os.Exit(1)
	}
	if !isExistAfterDelete {
		fmt.Println("Successfully deleted without failed resources")
		return
	}
	if stackOutputAfterDelete.Stacks[0].StackStatus != "DELETE_FAILED" {
		log.Fatalf("StackStatus is expected to be DELETE_FAILED, but %s", stackOutputAfterDelete.Stacks[0].StackStatus)
		os.Exit(1)
	}

	stackResources, err := cfnClient.ListStackResources(&opts.StackName)
	if err != nil {
		os.Exit(1)
	}

	resourcesLength := 0
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
			resourcesLength++

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

	if resourcesLength != len(stackArray)+len(bucketArray)+len(roleArray)+len(ecrArray)+len(ecrArray)+len(backupArray)+len(customArray) {
		fmt.Println("===========================================================")
		fmt.Printf("%v is FAILED !!!", opts.StackName)
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

	fmt.Printf("%v", resourcesLength)
	fmt.Printf("%v", len(stackArray))
	fmt.Printf("%v", len(bucketArray))
	fmt.Printf("%v", len(roleArray))
	fmt.Printf("%v", len(ecrArray))
	fmt.Printf("%v", len(backupArray))
	fmt.Printf("%v", len(customArray))
}
