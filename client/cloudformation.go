package client

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type CloudFormation struct {
	client *cloudformation.Client
	waiter *cloudformation.StackDeleteCompleteWaiter
}

func NewCloudFormation(config aws.Config) *CloudFormation {
	client := cloudformation.NewFromConfig(config)
	waiter := cloudformation.NewStackDeleteCompleteWaiter(client)
	return &CloudFormation{
		client,
		waiter,
	}
}

func (cfn *CloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	input := &cloudformation.DeleteStackInput{
		StackName:       stackName,
		RetainResources: retainResources,
	}

	_, err := cfn.client.DeleteStack(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed delete the cloudformation stack, %v", err)
		return err
	}

	if err := cfn.waitDeleteStack(stackName); err != nil {
		return err
	}

	return nil
}

func (cfn *CloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	output, err := cfn.client.DescribeStacks(context.TODO(), input)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return output, false, nil
	} else if err != nil {
		log.Fatalf("failed describe the cloudformation stack, %v", err)
		return output, true, err
	}

	return output, true, nil
}

func (cfn *CloudFormation) waitDeleteStack(stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := cfn.waiter.Wait(context.TODO(), input, 3600000000000)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		log.Fatalf("failed wait for stack deletion, %v", err)
		return err
	}

	return nil
}

func (cfn *CloudFormation) ListStackResources(stackName *string) (*cloudformation.ListStackResourcesOutput, error) {
	input := &cloudformation.ListStackResourcesInput{
		StackName: stackName,
	}

	output, err := cfn.client.ListStackResources(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed list the cloudformation stack resources, %v", err)
		return output, err
	}

	return output, nil
}
