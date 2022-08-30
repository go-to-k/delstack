package client

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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

func (cfnClient *CloudFormation) DeleteStack(stackName *string, retainResources []string) error {
	input := &cloudformation.DeleteStackInput{
		StackName:       stackName,
		RetainResources: retainResources,
	}

	_, err := cfnClient.client.DeleteStack(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed delete the cloudformation stack, %v", err)
		return err
	}

	if err := cfnClient.waitDeleteStack(stackName); err != nil {
		return err
	}

	return nil
}

func (cfnClient *CloudFormation) DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	output, err := cfnClient.client.DescribeStacks(context.TODO(), input)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return output, false, nil
	} else if err != nil {
		log.Fatalf("failed describe the cloudformation stack, %v", err)
		return output, true, err
	}

	return output, true, nil
}

func (cfnClient *CloudFormation) waitDeleteStack(stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := cfnClient.waiter.Wait(context.TODO(), input, 3600000000000)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		log.Fatalf("failed wait for stack deletion, %v", err)
		return err
	}

	return nil
}

func (cfnClient *CloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	var nextToken *string
	StackResourceSummaries := []types.StackResourceSummary{}

	for {
		input := &cloudformation.ListStackResourcesInput{
			StackName: stackName,
			NextToken: nextToken,
		}

		output, err := cfnClient.client.ListStackResources(context.TODO(), input)
		if err != nil {
			log.Fatalf("failed list the cloudformation stack resources, %v", err)
			return StackResourceSummaries, err
		}

		StackResourceSummaries = append(StackResourceSummaries, output.StackResourceSummaries...)
		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return StackResourceSummaries, nil
}