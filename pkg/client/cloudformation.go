package client

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const cloudFormationWaitNanoSecTime = time.Duration(4500000000000)

type ICloudFormation interface {
	DeleteStack(ctx context.Context, stackName *string, retainResources []string) error
	DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error)
	ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error)
	ListStacks(ctx context.Context) ([]types.StackSummary, error)
}

var _ ICloudFormation = (*CloudFormation)(nil)

type ICloudFormationSDKClient interface {
	DeleteStack(ctx context.Context, params *cloudformation.DeleteStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DeleteStackOutput, error)
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	ListStackResources(ctx context.Context, params *cloudformation.ListStackResourcesInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStackResourcesOutput, error)
	ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
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

func (cfnClient *CloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	input := &cloudformation.DeleteStackInput{
		StackName:       stackName,
		RetainResources: retainResources,
	}

	if _, err := cfnClient.client.DeleteStack(ctx, input); err != nil {
		return err
	}

	if err := cfnClient.waitDeleteStack(ctx, stackName); err != nil {
		return err
	}

	return nil
}

func (cfnClient *CloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	output, err := cfnClient.client.DescribeStacks(ctx, input)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return output, false, nil
	}

	return output, true, err
}

func (cfnClient *CloudFormation) waitDeleteStack(ctx context.Context, stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := cfnClient.waiter.Wait(ctx, input, cloudFormationWaitNanoSecTime)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		return err
	}

	return nil
}

func (cfnClient *CloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	var nextToken *string
	stackResourceSummaries := []types.StackResourceSummary{}

	for {
		select {
		case <-ctx.Done():
			return stackResourceSummaries, ctx.Err()
		default:
		}

		input := &cloudformation.ListStackResourcesInput{
			StackName: stackName,
			NextToken: nextToken,
		}

		output, err := cfnClient.client.ListStackResources(ctx, input)
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

func (cfnClient *CloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	var nextToken *string
	stackSummaries := []types.StackSummary{}

	for {
		select {
		case <-ctx.Done():
			return stackSummaries, ctx.Err()
		default:
		}

		input := &cloudformation.ListStacksInput{
			NextToken: nextToken,
		}

		output, err := cfnClient.client.ListStacks(ctx, input)
		if err != nil {
			return stackSummaries, err
		}

		stackSummaries = append(stackSummaries, output.StackSummaries...)
		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return stackSummaries, nil
}
