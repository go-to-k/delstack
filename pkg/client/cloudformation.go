//go:generate mockgen -source=./cloudformation.go -destination=./cloudformation_mock.go -package=client
package client

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const CloudFormationWaitNanoSecTime = time.Duration(4500000000000)

type ICloudFormation interface {
	DeleteStack(ctx context.Context, stackName *string, retainResources []string) error
	DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error)
	ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error)
	ListStacks(ctx context.Context) ([]types.StackSummary, error)
}

var _ ICloudFormation = (*CloudFormation)(nil)

type CloudFormation struct {
	client *cloudformation.Client
	waiter *cloudformation.StackDeleteCompleteWaiter
}

func NewCloudFormation(client *cloudformation.Client, waiter *cloudformation.StackDeleteCompleteWaiter) *CloudFormation {
	return &CloudFormation{
		client,
		waiter,
	}
}

func (c *CloudFormation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
	input := &cloudformation.DeleteStackInput{
		StackName:       stackName,
		RetainResources: retainResources,
	}

	if _, err := c.client.DeleteStack(ctx, input); err != nil {
		return err
	}

	if err := c.waitDeleteStack(ctx, stackName); err != nil {
		return err
	}

	return nil
}

func (c *CloudFormation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	output, err := c.client.DescribeStacks(ctx, input)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return output, false, nil
	}

	return output, true, err
}

func (c *CloudFormation) waitDeleteStack(ctx context.Context, stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := c.waiter.Wait(ctx, input, CloudFormationWaitNanoSecTime)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		return err
	}

	return nil
}

func (c *CloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

		output, err := c.client.ListStackResources(ctx, input)
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

func (c *CloudFormation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
	var nextToken *string
	stackSummaries := []types.StackSummary{}

	for {
		select {
		case <-ctx.Done():
			return stackSummaries, ctx.Err()
		default:
		}

		input := &cloudformation.ListStacksInput{
			StackStatusFilter: []types.StackStatus{
				types.StackStatusCreateInProgress,
				types.StackStatusCreateFailed,
				types.StackStatusCreateComplete,
				types.StackStatusRollbackInProgress,
				types.StackStatusRollbackFailed,
				types.StackStatusRollbackComplete,
				types.StackStatusDeleteInProgress,
				types.StackStatusDeleteFailed,
				types.StackStatusUpdateInProgress,
				types.StackStatusUpdateCompleteCleanupInProgress,
				types.StackStatusUpdateComplete,
				types.StackStatusUpdateFailed,
				types.StackStatusUpdateRollbackInProgress,
				types.StackStatusUpdateRollbackFailed,
				types.StackStatusUpdateRollbackCompleteCleanupInProgress,
				types.StackStatusUpdateRollbackComplete,
				types.StackStatusReviewInProgress,
				types.StackStatusImportInProgress,
				types.StackStatusImportComplete,
				types.StackStatusImportRollbackInProgress,
				types.StackStatusImportRollbackFailed,
				types.StackStatusImportRollbackComplete,
			},
			NextToken: nextToken,
		}

		output, err := c.client.ListStacks(ctx, input)
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
