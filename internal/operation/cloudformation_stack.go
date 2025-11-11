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
	"github.com/go-to-k/delstack/internal/resourcetype"
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

	for {
		stackResourceSummaries, err := o.client.ListStackResources(ctx, stackName)
		if err != nil {
			return err
		}

		operatorManager.SetOperatorCollection(stackName, stackResourceSummaries)

		if err = operatorManager.CheckResourceCounts(); err != nil {
			return err
		}

		if err = operatorManager.DeleteResourceCollection(ctx); err != nil {
			return err
		}

		if err = o.client.DeleteStack(ctx, stackName, operatorManager.GetLogicalResourceIds()); err != nil {
			return err
		}

		stacksAfterDelete, err := o.client.DescribeStacks(ctx, stackName)
		if err != nil {
			return err
		}
		if len(stacksAfterDelete) == 0 {
			break
		}
		if stacksAfterDelete[0].StackStatus != types.StackStatusDeleteFailed {
			return fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but %v: %v", stacksAfterDelete[0].StackStatus, *stackName)
		}
	}

	return nil
}

func (o *CloudFormationStackOperator) deleteStackNormally(ctx context.Context, stackName *string, isRootStack bool) (bool, error) {
	stacksBeforeDelete, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return false, err
	}
	if len(stacksBeforeDelete) == 0 && isRootStack {
		errMsg := fmt.Sprintf("%s not found", *stackName)
		return false, fmt.Errorf("NotExistsError: %v", errMsg)
	}
	if len(stacksBeforeDelete) == 0 {
		return true, nil
	}

	if stacksBeforeDelete[0].EnableTerminationProtection != nil && *stacksBeforeDelete[0].EnableTerminationProtection {
		return false, fmt.Errorf("TerminationProtectionError: %v", *stackName)
	}
	if o.isExceptedByStackStatus(stacksBeforeDelete[0].StackStatus) {
		return false, fmt.Errorf("OperationInProgressError: Stacks with XxxInProgress cannot be deleted, but %v: %v", stacksBeforeDelete[0].StackStatus, *stackName)
	}

	if deleteErr := o.client.DeleteStack(ctx, stackName, []string{}); deleteErr != nil {
		return false, deleteErr
	}

	stacksAfterDelete, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return false, err
	}
	if len(stacksAfterDelete) == 0 {
		io.Logger.Info().Msgf("%v: No resources were DELETE_FAILED.", *stackName)
		return true, nil
	}
	if stacksAfterDelete[0].StackStatus != types.StackStatusDeleteFailed {
		return false, fmt.Errorf("StackStatusError: StackStatus is expected to be DELETE_FAILED, but %v: %v", stacksAfterDelete[0].StackStatus, *stackName)
	}

	return false, nil
}

func (o *CloudFormationStackOperator) GetSortedStackNames(ctx context.Context, stackNames []string) ([]string, error) {
	sortedStackNames := []string{}
	gotStacks := []types.Stack{}
	notFoundStackNames := []string{}
	terminationProtectionStackNames := []string{}

	type stackNameInProgress struct {
		stackName   string
		stackStatus types.StackStatus
	}
	stackNamesInProgress := []stackNameInProgress{}

	for _, stackName := range stackNames {
		stack, err := o.client.DescribeStacks(ctx, aws.String(stackName))
		if err != nil {
			return sortedStackNames, err
		}

		if len(stack) == 0 {
			notFoundStackNames = append(notFoundStackNames, stackName)
			continue
		}

		// except the stacks with EnableTerminationProtection
		if stack[0].EnableTerminationProtection != nil && *stack[0].EnableTerminationProtection {
			terminationProtectionStackNames = append(terminationProtectionStackNames, stackName)
			continue
		}

		// except the stacks that are in the exception list
		if o.isExceptedByStackStatus(stack[0].StackStatus) {
			stackNamesInProgress = append(stackNamesInProgress, stackNameInProgress{
				stackName:   stackName,
				stackStatus: stack[0].StackStatus,
			})
			continue
		}

		gotStacks = append(gotStacks, stack[0]) // DescribeStacks returns a stack with a single element
	}

	if len(notFoundStackNames) > 0 {
		errMsg := fmt.Sprintf("%s not found", strings.Join(notFoundStackNames, ", "))
		return sortedStackNames, fmt.Errorf("NotExistsError: %v", errMsg)
	}
	if len(terminationProtectionStackNames) > 0 {
		return sortedStackNames, fmt.Errorf("TerminationProtectionError: %v", strings.Join(terminationProtectionStackNames, ", "))
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
		errMsg := fmt.Sprintf("No stacks matching the keyword (%s)", *keyword)
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

func (o *CloudFormationStackOperator) RemoveDeletionPolicy(ctx context.Context, stackName *string) error {
	stacks, err := o.client.DescribeStacks(ctx, stackName)
	if err != nil {
		return err
	}
	if len(stacks) == 0 {
		return fmt.Errorf("NotExistsError: %v", *stackName)
	}
	// If the stack is in the ROLLBACK_COMPLETE state, it is not possible to update the stack.
	if stacks[0].StackStatus == types.StackStatusRollbackComplete {
		return nil
	}

	stackResourceSummaries, err := o.client.ListStackResources(ctx, stackName)
	if err != nil {
		return err
	}

	nestedStacks := []string{}
	for _, stackResourceSummary := range stackResourceSummaries {
		if aws.ToString(stackResourceSummary.ResourceType) == resourcetype.CloudformationStack {
			nestedStacks = append(nestedStacks, *stackResourceSummary.PhysicalResourceId)
		}
	}

	template, err := o.client.GetTemplate(ctx, stackName)
	if err != nil {
		return err
	}

	modifiedTemplate := o.removeDeletionPolicyFromTemplate(template)
	if len(nestedStacks) == 0 && modifiedTemplate == *template {
		return nil
	}
	if modifiedTemplate != *template {
		if err = o.client.UpdateStack(ctx, stackName, &modifiedTemplate, stacks[0].Parameters); err != nil {
			return err
		}
	}

	// If we update the child stack first, after the child stack is updated, the parent stack will be updated
	// and get the old child stack's TemplateURL, causing the child stack update to revert.
	// Therefore, we should update the parent stack instead of updating the child stack first.
	// Also, we don't control the number of threads with semaphore because the number of nested stacks is usually small.
	eg, ctx := errgroup.WithContext(ctx)
	for _, stackName := range nestedStacks {
		eg.Go(func() error {
			return o.RemoveDeletionPolicy(ctx, aws.String(stackName))
		})
	}

	return eg.Wait()
}

// removeDeletionPolicyFromTemplate removes DeletionPolicy properties with Retain or RetainExceptOnCreate values
// from CloudFormation templates while preserving the original formatting.
//
// This function uses a line-based string processing approach instead of YAML/JSON parsers to ensure that:
// - Original indentation (spaces/tabs) is completely preserved
// - Property order remains unchanged
// - Line breaks and whitespace are maintained exactly as in the input
//
// Supported formats:
// - YAML inline: "DeletionPolicy: Retain"
// - YAML block: "DeletionPolicy:\n  Retain"
// - JSON formatted: "\"DeletionPolicy\": \"Retain\""
// - JSON minified: single-line JSON without newlines
//
// Note: This does NOT remove DeletionPolicy with "Delete" or "Snapshot" values.
func (o *CloudFormationStackOperator) removeDeletionPolicyFromTemplate(template *string) string {
	// Handle minified JSON (single line)
	if !strings.Contains(*template, "\n") {
		return o.removeFromMinifiedJSON(*template)
	}

	// Handle multi-line templates (YAML or formatted JSON)
	return o.removeFromMultiLine(*template)
}

// removeFromMinifiedJSON removes DeletionPolicy from single-line (minified) JSON templates.
// It handles comma placement to maintain valid JSON syntax after removal.
func (o *CloudFormationStackOperator) removeFromMinifiedJSON(template string) string {
	// For minified JSON, use a simpler approach: match the entire key-value with surrounding commas
	// Match: "DeletionPolicy":"Retain", or ,"DeletionPolicy":"Retain" or "DeletionPolicy":"Retain"
	result := regexp.MustCompile(`["']?DeletionPolicy["']?\s*:\s*["']?(?:Retain|RetainExceptOnCreate)["']?\s*,\s*`).ReplaceAllString(template, "")
	result = regexp.MustCompile(`,\s*["']?DeletionPolicy["']?\s*:\s*["']?(?:Retain|RetainExceptOnCreate)["']?\s*`).ReplaceAllString(result, "")
	return result
}

// removeFromMultiLine removes DeletionPolicy from multi-line templates (formatted JSON or YAML).
// It preserves the original indentation, line breaks, and property order by processing line by line.
// Supports both YAML inline format ("DeletionPolicy: Retain") and block format ("DeletionPolicy:\n  Retain").
func (o *CloudFormationStackOperator) removeFromMultiLine(template string) string {
	lines := strings.Split(template, "\n")
	result := make([]string, 0, len(lines))

	// Pattern to match DeletionPolicy lines with Retain or RetainExceptOnCreate (inline format)
	inlinePattern := regexp.MustCompile(`^\s*["']?DeletionPolicy["']?\s*:\s*["']?(?:Retain|RetainExceptOnCreate)["']?\s*,?\s*$`)
	// Pattern for YAML block format: DeletionPolicy key without value on same line
	keyOnlyPattern := regexp.MustCompile(`^\s*["']?DeletionPolicy["']?\s*:\s*$`)
	// Pattern for the value line (indented Retain or RetainExceptOnCreate)
	valueOnlyPattern := regexp.MustCompile(`^\s+["']?(?:Retain|RetainExceptOnCreate)["']?\s*$`)

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check for YAML block format (key on one line, value on next)
		if keyOnlyPattern.MatchString(line) {
			if i+1 < len(lines) && valueOnlyPattern.MatchString(lines[i+1]) {
				// Skip both the key and value lines
				i += 2
				continue
			}
		}

		// Check for inline format (key and value on same line)
		if inlinePattern.MatchString(line) {
			// Remove trailing comma from previous line if next line is closing bracket
			if i > 0 && len(result) > 0 && i+1 < len(lines) {
				if regexp.MustCompile(`^\s*[}\]]`).MatchString(lines[i+1]) && regexp.MustCompile(`,\s*$`).MatchString(result[len(result)-1]) {
					result[len(result)-1] = regexp.MustCompile(`,(\s*)$`).ReplaceAllString(result[len(result)-1], "$1")
				}
			}
			i++
			continue
		}

		result = append(result, line)
		i++
	}

	return strings.Join(result, "\n")
}
