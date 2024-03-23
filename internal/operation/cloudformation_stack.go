package operation

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"sort"
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

// Except xxInProgress
// StackStatusDeleteComplete is not included because DescribeStacks does not return a DELETE_COMPLETE stack.
var StackStatusExceptionsForDescribeStacks = []types.StackStatus{
	types.StackStatusCreateInProgress,
	types.StackStatusRollbackInProgress,
	types.StackStatusDeleteInProgress,
	types.StackStatusUpdateInProgress,
	types.StackStatusUpdateCompleteCleanupInProgress,
	types.StackStatusUpdateRollbackInProgress,
	types.StackStatusUpdateRollbackCompleteCleanupInProgress,
	types.StackStatusReviewInProgress,
	types.StackStatusImportInProgress,
	types.StackStatusImportRollbackInProgress,
}

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
	stacksBeforeDelete, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return false, err
	}
	if len(stacksBeforeDelete) == 0 && isRootStack {
		errMsg := fmt.Sprintf("%s stack not found.", *stackName)
		return false, fmt.Errorf("NotExistsError: %v", errMsg)
	}
	if len(stacksBeforeDelete) == 0 {
		return true, nil
	}

	if stacksBeforeDelete[0].EnableTerminationProtection != nil && *stacksBeforeDelete[0].EnableTerminationProtection {
		return false, fmt.Errorf("TerminationProtectionIsEnabled: %v", *stackName)
	}
	if o.isExceptedByStackStatus(stacksBeforeDelete[0].StackStatus) {
		return false, fmt.Errorf("OperationInProgressError: Stacks with XxxInProgress cannot be deleted, but %v: %v", stacksBeforeDelete[0].StackStatus, *stackName)
	}

	if err := o.client.DeleteStack(ctx, stackName, []string{}); err != nil {
		return false, err
	}

	stacksAfterDelete, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return false, err
	}
	if len(stacksAfterDelete) == 0 {
		io.Logger.Info().Msgf("%v: No resources were DELETE_FAILED.", *stackName)
		return true, nil
	}
	if stacksAfterDelete[0].StackStatus != "DELETE_FAILED" {
		return false, fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but %v: %v", stacksAfterDelete[0].StackStatus, *stackName)
	}

	return false, nil
}

func (o *CloudFormationStackOperator) GetSortedStackNames(ctx context.Context, stackNames []string) ([]string, error) {
	sortedStackNames := []string{}
	gotStacks := []types.Stack{}
	notFoundStackNames := []string{}
	stackNamesInProgress := []struct {
		stackName   string
		stackStatus types.StackStatus
	}{}

	for _, stackName := range stackNames {
		stack, err := o.client.DescribeStacks(ctx, aws.String(stackName))
		if err != nil {
			return sortedStackNames, err
		}

		if len(stack) == 0 {
			notFoundStackNames = append(notFoundStackNames, stackName)
			continue
		}
		if o.isExceptedByStackStatus(stack[0].StackStatus) {
			stackNamesInProgress = append(stackNamesInProgress, struct {
				stackName   string
				stackStatus types.StackStatus
			}{
				stackName:   stackName,
				stackStatus: stack[0].StackStatus,
			})
			continue
		}
		gotStacks = append(gotStacks, stack[0]) // DescribeStacks returns a stack with a single element
	}

	if len(notFoundStackNames) > 0 {
		errMsg := fmt.Sprintf("%s stack not found.", strings.Join(notFoundStackNames, ", "))
		return sortedStackNames, fmt.Errorf("NotExistsError: %v", errMsg)
	}
	if len(stackNamesInProgress) > 0 {
		var stackNamesWithStatus []string
		for _, stack := range stackNamesInProgress {
			stackNamesWithStatus = append(stackNamesWithStatus, fmt.Sprintf("%s: %s", stack.stackStatus, stack.stackName))
		}
		errMsg := fmt.Sprintf("Stacks with XxxInProgress cannot be deleted, but %s", strings.Join(stackNamesWithStatus, ", "))
		return sortedStackNames, fmt.Errorf("OperationInProgressError: %v", errMsg)
	}

	// Sort gotStacks in descending order by stack.CreationTime
	sort.Slice(gotStacks, func(i, j int) bool {
		return gotStacks[i].CreationTime.After(*gotStacks[j].CreationTime)
	})
	for _, stack := range gotStacks {
		sortedStackNames = append(sortedStackNames, *stack.StackName)
	}
	return sortedStackNames, nil
}

func (o *CloudFormationStackOperator) ListStacksFilteredByKeyword(ctx context.Context, keyword *string) ([]string, error) {
	filteredStacks := []string{}

	// Use DescribeStacks instead of ListStacks to take EnableTerminationProtection
	stacks, err := o.client.DescribeStacks(ctx, nil)
	if err != nil {
		return filteredStacks, err
	}

	lowerKeyword := strings.ToLower(*keyword)

	for _, stack := range stacks {
		// except the nested child stacks
		if stack.RootId != nil {
			continue
		}

		// except the stacks with EnableTerminationProtection
		if stack.EnableTerminationProtection != nil && *stack.EnableTerminationProtection {
			continue
		}

		// except the stacks that are in the exception list
		if o.isExceptedByStackStatus(stack.StackStatus) {
			continue
		}

		// for case-insensitive
		lowerStackName := strings.ToLower(*stack.StackName)
		if strings.Contains(lowerStackName, lowerKeyword) {
			filteredStacks = append(filteredStacks, *stack.StackName)
		}
	}

	if len(filteredStacks) == 0 {
		errMsg := fmt.Sprintf("No stacks matching the keyword (%s).", *keyword)
		return filteredStacks, fmt.Errorf("NotExistsError: %v", errMsg)
	}

	return filteredStacks, nil
}

func (o *CloudFormationStackOperator) isExceptedByStackStatus(stackStatus types.StackStatus) bool {
	for _, status := range StackStatusExceptionsForDescribeStacks {
		if stackStatus == status {
			return true
		}
	}
	return false
}
