package operation

import (
	"context"
	"fmt"
	"regexp"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/logger"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*StackOperator)(nil)

const STACK_NAME_RULE = `^arn:aws:cloudformation:[^:]*:[0-9]*:stack/([^/]*)/.*$`

type StackOperator struct {
	config              aws.Config
	client              client.ICloudFormation
	resources           []*types.StackResourceSummary
	targetResourceTypes []string
}

func NewStackOperator(config aws.Config, client client.ICloudFormation, targetResourceTypes []string) *StackOperator {
	return &StackOperator{
		config:              config,
		client:              client,
		resources:           []*types.StackResourceSummary{},
		targetResourceTypes: targetResourceTypes,
	}
}

func (operator *StackOperator) AddResource(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *StackOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *StackOperator) DeleteResources() error {
	var eg errgroup.Group
	re := regexp.MustCompile(STACK_NAME_RULE)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, stack := range operator.resources {
		stack := stack
		eg.Go(func() error {
			stackName := re.ReplaceAllString(aws.ToString(stack.PhysicalResourceId), `$1`)
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			isRootStack := false
			operatorFactory := NewOperatorFactory(operator.config)
			operatorCollection := NewOperatorCollection(operator.config, operatorFactory, operator.targetResourceTypes)
			operatorManager := NewOperatorManager(operatorCollection)

			return operator.DeleteStackResources(aws.String(stackName), isRootStack, operatorManager)
		})
	}

	return eg.Wait()
}

func (operator *StackOperator) DeleteStackResources(stackName *string, isRootStack bool, operatorManager IOperatorManager) error {
	isSuccess, err := operator.deleteStackNormally(stackName, isRootStack)
	if err != nil {
		return err
	}
	if isSuccess {
		return nil
	}

	stackResourceSummaries, err := operator.client.ListStackResources(stackName)
	if err != nil {
		return err
	}

	operatorManager.SetOperatorCollection(stackName, stackResourceSummaries)

	if err := operatorManager.CheckResourceCounts(); err != nil {
		return err
	}

	if err := operatorManager.DeleteResourceCollection(); err != nil {
		return err
	}

	if err := operator.client.DeleteStack(stackName, operatorManager.GetLogicalResourceIds()); err != nil {
		return err
	}

	return nil
}

func (operator *StackOperator) deleteStackNormally(stackName *string, isRootStack bool) (bool, error) {
	stackOutputBeforeDelete, isExistBeforeDelete, err := operator.client.DescribeStacks(stackName)
	if err != nil {
		return false, err
	}
	if !isExistBeforeDelete && isRootStack {
		return false, fmt.Errorf("NotExistsError: %v", *stackName)
	}
	if !isExistBeforeDelete {
		return true, nil
	}

	if *stackOutputBeforeDelete.Stacks[0].EnableTerminationProtection {
		return false, fmt.Errorf("TerminationProtectionIsEnabled: %v", *stackName)
	}

	if stackOutputBeforeDelete.Stacks[0].StackStatus == "DELETE_FAILED" {
		return false, nil
	}

	if err := operator.client.DeleteStack(stackName, []string{}); err != nil {
		return false, err
	}

	stackOutputAfterDelete, isExistAfterDelete, err := operator.client.DescribeStacks(stackName)
	if err != nil {
		return false, err
	}
	if !isExistAfterDelete {
		logger.Logger.Info().Msg("No resources were DELETE_FAILED.")
		return true, nil
	}
	if stackOutputAfterDelete.Stacks[0].StackStatus != "DELETE_FAILED" {
		return false, fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but %v: %v", stackOutputAfterDelete.Stacks[0].StackStatus, *stackName)
	}

	return false, nil
}
