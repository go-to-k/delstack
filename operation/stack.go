package operation

import (
	"context"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ Operator = (*StackOperator)(nil)

const STACK_NAME_RULE = `^arn:aws:cloudformation:[^:]*:[0-9]*:stack/([^/]*)/.*$`

type StackOperator struct {
	config    aws.Config
	client    *client.CloudFormation
	resources []*types.StackResourceSummary
}

func NewStackOperator(config aws.Config) *StackOperator {
	client := client.NewCloudFormation(config)
	return &StackOperator{
		config:    config,
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *StackOperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *StackOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *StackOperator) DeleteResources() error {
	var eg errgroup.Group
	re := regexp.MustCompile(STACK_NAME_RULE)
	sem := semaphore.NewWeighted(int64(option.CONCURRENCY_NUM))

	for _, stack := range operator.resources {
		stack := stack
		eg.Go(func() error {
			stackName := re.ReplaceAllString(aws.ToString(stack.PhysicalResourceId), `$1`)
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			isRootStack := false
			if err := operator.DeleteStackResources(aws.String(stackName), isRootStack); err != nil {
				return err
			}

			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (operator *StackOperator) DeleteStackResources(stackName *string, isRootStack bool) error {
	if isRootStack {
		err := operator.deleteRootStack(stackName)
		if err != nil {
			return err
		}
	}

	stackResourceSummaries, err := operator.client.ListStackResources(stackName)
	if err != nil {
		return err
	}

	operatorManager := NewOperatorManager(operator.config, stackName, stackResourceSummaries)
	if err := operatorManager.CheckResourceCounts(); err != nil {
		return err
	}

	if err := operator.client.DeleteStack(stackName, operatorManager.GetLogicalResourceIds()); err != nil {
		return err
	}

	if err := operatorManager.DeleteResourceCollection(); err != nil {
		return err
	}

	return nil
}

func (operator *StackOperator) deleteRootStack(stackName *string) error {
	stackOutputBeforeDelete, isExistBeforeDelete, err := operator.client.DescribeStacks(stackName)
	if err != nil {
		return err
	}
	if !isExistBeforeDelete {
		logger.Logger.Info().Msgf("The stack is not exists: %v\n", *stackName)
		return err
	}

	if *stackOutputBeforeDelete.Stacks[0].EnableTerminationProtection {
		logger.Logger.Info().Msgf("TerminationProtection is enabled: %v\n", *stackName)
		return nil
	}

	if err := operator.client.DeleteStack(stackName, []string{}); err != nil {
		return err
	}

	stackOutputAfterDelete, isExistAfterDelete, err := operator.client.DescribeStacks(stackName)
	if err != nil {
		return err
	}
	if !isExistAfterDelete {
		logger.Logger.Info().Msgf("Successfully deleted without failed resources: %v\n", *stackName)
		return nil
	}
	if stackOutputAfterDelete.Stacks[0].StackStatus != "DELETE_FAILED" {
		logger.Logger.Fatal().Msgf("Error: StackStatus is expected to be DELETE_FAILED, but %v: %v", stackOutputAfterDelete.Stacks[0].StackStatus, *stackName)
		return err
	}

	return nil
}
