package client

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

const CloudformationWaitNanoSecTime = time.Duration(4500000000000)

type ICloudformation interface {
	DeleteStack(ctx context.Context, stackName *string, retainResources []string) error
	DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error)
	ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error)
	ListStacks(ctx context.Context) ([]types.StackSummary, error)
}

var _ ICloudformation = (*Cloudformation)(nil)

type Cloudformation struct {
	client *cloudformation.Client
	waiter *cloudformation.StackDeleteCompleteWaiter
}

func NewCloudformation(client *cloudformation.Client, waiter *cloudformation.StackDeleteCompleteWaiter) *Cloudformation {
	return &Cloudformation{
		client,
		waiter,
	}
}

func (c *Cloudformation) DeleteStack(ctx context.Context, stackName *string, retainResources []string) error {
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

func (c *Cloudformation) DescribeStacks(ctx context.Context, stackName *string) (*cloudformation.DescribeStacksOutput, bool, error) {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	output, err := c.client.DescribeStacks(ctx, input)
	if err != nil && strings.Contains(err.Error(), "does not exist") {
		return output, false, nil
	}

	return output, true, err
}

func (c *Cloudformation) waitDeleteStack(ctx context.Context, stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := c.waiter.Wait(ctx, input, CloudformationWaitNanoSecTime)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		return err
	}

	return nil
}

func (c *Cloudformation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
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

func (c *Cloudformation) ListStacks(ctx context.Context) ([]types.StackSummary, error) {
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
