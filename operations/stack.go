package operations

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
)

func DeleteStacks(config aws.Config, resources []types.StackResourceSummary) error {
	// TODO: Concurrency DeleteStack
	re := regexp.MustCompile(`^arn:aws:cloudformation:[^:]*:[0-9]*:stack/([^/]*)/.*$`)
	for _, stack := range resources {
		stackName := re.ReplaceAllString(*stack.PhysicalResourceId, `$1`)
		if err := DeleteStackResources(config, stackName); err != nil {
			return err
		}
	}
	return nil
}

func DeleteStackResources(config aws.Config, stackName string) error {
	cfnClient := client.NewCloudFormation(config)

	stackOutputBeforeDelete, isExistBeforeDelete, err := cfnClient.DescribeStacks(stackName)
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

	if err := cfnClient.DeleteStack(stackName, []string{}); err != nil {
		return err
	}

	stackOutputAfterDelete, isExistAfterDelete, err := cfnClient.DescribeStacks(stackName)
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

	stackResourceSummaries, err := cfnClient.ListStackResources(stackName)
	if err != nil {
		return err
	}

	collection := NewResourceCollection(config, stackName, stackResourceSummaries)
	if err := collection.CheckResourceCounts(); err != nil {
		return err
	}

	if err := cfnClient.DeleteStack(stackName, collection.LogicalResourceIds); err != nil {
		return err
	}

	if err := collection.DeleteResourceCollection(); err != nil {
		return err
	}

	return nil
}
