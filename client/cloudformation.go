package client

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/option"
)

type ICloudFormation interface {
	DeleteStack(stackName *string, retainResources []string) error
	DescribeStacks(stackName *string) (*cloudformation.DescribeStacksOutput, bool, error)
	ListStackResources(stackName *string) ([]types.StackResourceSummary, error)
}

var _ ICloudFormation = (*CloudFormation)(nil)

type ICloudFormationSDKClient interface {
	DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error)
}

type ICloudFormationSDKWaiter interface {
	Wait(ctx context.Context, params *cloudformation.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cloudformation.StackDeleteCompleteWaiterOptions)) error
}

type CloudFormation struct {
	client ICloudFormationSDKClient
	waiter ICloudFormationSDKWaiter
}

func NewCloudFormation(client ICloudFormationSDKClient, waiter ICloudFormationSDKWaiter) *CloudFormation {
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

	if _, err := cfnClient.client.DeleteStack(context.TODO(), input); err != nil {
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
	}

	return output, true, err
}

func (cfnClient *CloudFormation) waitDeleteStack(stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := cfnClient.waiter.Wait(context.TODO(), input, option.CloudFormationWaitNanoSecTime)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		return err
	}

	return nil
}

func (cfnClient *CloudFormation) ListStackResources(stackName *string) ([]types.StackResourceSummary, error) {
	var nextToken *string
	stackResourceSummaries := []types.StackResourceSummary{}

	for {
		input := &cloudformation.ListStackResourcesInput{
			StackName: stackName,
			NextToken: nextToken,
		}

		output, err := cfnClient.client.ListStackResources(context.TODO(), input)
		if err != nil {
			return stackResourceSummaries, err
		}

		stackResourceSummaries = append(stackResourceSummaries, output.StackResourceSummaries...)
		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return stackResourceSummaries, nil
}
