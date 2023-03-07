package operation

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*CloudFormationStackOperator)(nil)

const StackNameRule = `^arn:aws:cloudformation:[^:]*:[0-9]*:stack/([^/]*)/.*$`

var StackNameRuleRegExp = regexp.MustCompile(StackNameRule)

type CloudFormationStackOperator struct {
	config              aws.Config
	client              client.ICloudFormation
	resources           []*types.StackResourceSummary
	targetResourceTypes []string
}

func NewCloudFormationStackOperator(config aws.Config, client client.ICloudFormation, targetResourceTypes []string) *CloudFormationStackOperator {
	return &CloudFormationStackOperator{
		config:              config,
		client:              client,
		resources:           []*types.StackResourceSummary{},
		targetResourceTypes: targetResourceTypes,
	}
}

func (o *CloudFormationStackOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *CloudFormationStackOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *CloudFormationStackOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, stack := range o.resources {
		stack := stack
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)
			stackName := StackNameRuleRegExp.ReplaceAllString(aws.ToString(stack.PhysicalResourceId), `$1`)

			isRootStack := false
			operatorFactory := NewOperatorFactory(o.config)
			operatorCollection := NewOperatorCollection(o.config, operatorFactory, o.targetResourceTypes)
			operatorManager := NewOperatorManager(operatorCollection)

			return o.DeleteCloudFormationStack(ctx, aws.String(stackName), isRootStack, operatorManager)
		})
	}

	return eg.Wait()
}

func (o *CloudFormationStackOperator) DeleteCloudFormationStack(ctx context.Context, stackName *string, isRootStack bool, operatorManager IOperatorManager) error {
	isSuccess, err := o.deleteStackNormally(ctx, stackName, isRootStack)
	if err != nil {
		return err
	}
	if isSuccess {
		return nil
	}

	stackResourceSummaries, err := o.client.ListStackResources(ctx, stackName)
	if err != nil {
		return err
	}

	operatorManager.SetOperatorCollection(stackName, stackResourceSummaries)

	if err := operatorManager.CheckResourceCounts(); err != nil {
		return err
	}

	if err := operatorManager.DeleteResourceCollection(ctx); err != nil {
		return err
	}

	if err := o.client.DeleteStack(ctx, stackName, operatorManager.GetLogicalResourceIds()); err != nil {
		return err
	}

	return nil
}

func (o *CloudFormationStackOperator) deleteStackNormally(ctx context.Context, stackName *string, isRootStack bool) (bool, error) {
	stackOutputBeforeDelete, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return false, err
	}
	if len(stackOutputBeforeDelete) == 0 && isRootStack {
		errMsg := fmt.Sprintf("%s stack not found.", *stackName)
		return false, fmt.Errorf("NotExistsError: %v", errMsg)
	}
	if len(stackOutputBeforeDelete) == 0 {
		return true, nil
	}

	if *stackOutputBeforeDelete[0].EnableTerminationProtection {
		return false, fmt.Errorf("TerminationProtectionIsEnabled: %v", *stackName)
	}

	if err := o.client.DeleteStack(ctx, stackName, []string{}); err != nil {
		return false, err
	}

	stackOutputAfterDelete, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return false, err
	}
	if len(stackOutputAfterDelete) == 0 {
		io.Logger.Info().Msg("No resources were DELETE_FAILED.")
		return true, nil
	}
	if stackOutputAfterDelete[0].StackStatus != "DELETE_FAILED" {
		return false, fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but %v: %v", stackOutputAfterDelete[0].StackStatus, *stackName)
	}

	return false, nil
}

func (o *CloudFormationStackOperator) ListStacksFilteredByKeyword(ctx context.Context, keyword *string) ([]string, error) {
	filteredStacks := []string{}

	stackSummaries, err := o.client.ListStacks(ctx)
	if err != nil {
		return filteredStacks, err
	}

	for _, stackSummary := range stackSummaries {
		// pass the nested child stacks
		if stackSummary.RootId != nil {
			continue
		}

		// for case-insensitive
		lowerStackName := strings.ToLower(*stackSummary.StackName)
		lowerKeyword := strings.ToLower(*keyword)
		if strings.Contains(lowerStackName, lowerKeyword) {
			filteredStacks = append(filteredStacks, *stackSummary.StackName)
		}
	}

	return filteredStacks, nil
}
