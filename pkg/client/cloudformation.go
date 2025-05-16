//go:generate mockgen -source=$GOFILE -destination=cloudformation_mock.go -package=$GOPACKAGE -write_package_comment=false
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
	DescribeStacks(ctx context.Context, stackName *string) ([]types.Stack, error)
	ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error)
	GetTemplate(ctx context.Context, stackName *string) (*string, error)
	UpdateStack(ctx context.Context, stackName *string, templateBody *string) error
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
		return &ClientError{
			ResourceName: stackName,
			Err:          err,
		}
	}

	if err := c.waitDeleteStack(ctx, stackName); err != nil {
		return &ClientError{
			ResourceName: stackName,
			Err:          err,
		}
	}

	return nil
}

func (c *CloudFormation) DescribeStacks(ctx context.Context, stackName *string) ([]types.Stack, error) {
	var nextToken *string
	stacks := []types.Stack{}

	for {
		select {
		case <-ctx.Done():
			return stacks, &ClientError{
				ResourceName: stackName,
				Err:          ctx.Err(),
			}
		default:
		}

		// If a stackName is nil, then return all stacks
		input := &cloudformation.DescribeStacksInput{
			NextToken: nextToken,
			StackName: stackName,
		}

		output, err := c.client.DescribeStacks(ctx, input)
		if err != nil && strings.Contains(err.Error(), "does not exist") {
			return stacks, nil
		}
		if err != nil {
			return stacks, &ClientError{
				ResourceName: stackName,
				Err:          err,
			}
		}

		if len(stacks) == 0 && len(output.Stacks) == 0 {
			return stacks, nil
		}
		stacks = append(stacks, output.Stacks...)

		nextToken = output.NextToken
		if nextToken == nil {
			break
		}
	}
	return stacks, nil
}

func (c *CloudFormation) waitDeleteStack(ctx context.Context, stackName *string) error {
	input := &cloudformation.DescribeStacksInput{
		StackName: stackName,
	}

	err := c.waiter.Wait(ctx, input, CloudFormationWaitNanoSecTime)
	if err != nil && !strings.Contains(err.Error(), "waiter state transitioned to Failure") {
		return err // return non wrapping error because wrap in public callers
	}

	return nil
}

func (c *CloudFormation) ListStackResources(ctx context.Context, stackName *string) ([]types.StackResourceSummary, error) {
	var nextToken *string
	stackResourceSummaries := []types.StackResourceSummary{}

	for {
		select {
		case <-ctx.Done():
			return stackResourceSummaries, &ClientError{
				ResourceName: stackName,
				Err:          ctx.Err(),
			}
		default:
		}

		input := &cloudformation.ListStackResourcesInput{
			StackName: stackName,
			NextToken: nextToken,
		}

		output, err := c.client.ListStackResources(ctx, input)
		if err != nil {
			return stackResourceSummaries, &ClientError{
				ResourceName: stackName,
				Err:          err,
			}
		}

		stackResourceSummaries = append(stackResourceSummaries, output.StackResourceSummaries...)
		nextToken = output.NextToken

		if nextToken == nil {
			break
		}
	}

	return stackResourceSummaries, nil
}

func (c *CloudFormation) GetTemplate(ctx context.Context, stackName *string) (*string, error) {
	input := &cloudformation.GetTemplateInput{
		StackName: stackName,
	}

	output, err := c.client.GetTemplate(ctx, input)
	if err != nil {
		return nil, &ClientError{
			ResourceName: stackName,
			Err:          err,
		}
	}

	return output.TemplateBody, nil
}

func (c *CloudFormation) UpdateStack(ctx context.Context, stackName *string, templateBody *string) error {
	input := &cloudformation.UpdateStackInput{
		StackName:    stackName,
		TemplateBody: templateBody,
		Capabilities: []types.Capability{
			types.CapabilityCapabilityIam,
		},
	}

	_, err := c.client.UpdateStack(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: stackName,
			Err:          err,
		}
	}

	return nil
}
