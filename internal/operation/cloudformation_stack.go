package operation

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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

type S3UploadResult struct {
	TemplateURL *string
	BucketName  *string
	Key         *string
}

type CloudFormationStackOperator struct {
	config              aws.Config
	client              client.ICloudFormation
	s3Client            client.IS3
	resources           []*types.StackResourceSummary
	targetResourceTypes []string
}

func NewCloudFormationStackOperator(config aws.Config, client client.ICloudFormation, s3Client client.IS3, targetResourceTypes []string) *CloudFormationStackOperator {
	return &CloudFormationStackOperator{
		config:              config,
		client:              client,
		s3Client:            s3Client,
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

type exportKey struct {
	exportingStack string
	exportName     string
}

// BuildDependencyGraph analyzes Output/Import dependencies among the specified stacks
func (o *CloudFormationStackOperator) BuildDependencyGraph(
	ctx context.Context,
	stackNames []string,
) (*StackDependencyGraph, error) {
	graph := NewStackDependencyGraph(stackNames)

	stackNameSet := make(map[string]struct{})
	for _, name := range stackNames {
		stackNameSet[name] = struct{}{}
	}

	externalReferences := make(map[exportKey][]string)

	for _, stackName := range stackNames {
		stacks, err := o.client.DescribeStacks(ctx, aws.String(stackName))
		if err != nil {
			return nil, err
		}

		if len(stacks) == 0 {
			continue
		}

		stack := stacks[0]

		for _, output := range stack.Outputs {
			if output.ExportName == nil {
				continue
			}

			exportName := *output.ExportName

			importingStacks, err := o.client.ListImports(ctx, aws.String(exportName))
			if err != nil {
				// If the export is not imported by any stack, ListImports returns ValidationError
				// This is not an error condition for dependency graph building
				if strings.Contains(err.Error(), "is not imported by any stack") {
					continue
				}
				return nil, err
			}

			for _, importingStack := range importingStacks {
				if _, isTarget := stackNameSet[importingStack]; isTarget {
					graph.AddDependency(importingStack, stackName)
				} else {
					key := exportKey{
						exportingStack: stackName,
						exportName:     exportName,
					}
					externalReferences[key] = append(externalReferences[key], importingStack)
				}
			}
		}
	}

	if len(externalReferences) > 0 {
		return nil, o.buildExternalReferenceError(externalReferences)
	}

	return graph, nil
}

func (o *CloudFormationStackOperator) buildExternalReferenceError(externalReferences map[exportKey][]string) error {
	keys := slices.Collect(maps.Keys(externalReferences))
	slices.SortFunc(keys, func(a, b exportKey) int {
		if c := strings.Compare(a.exportingStack, b.exportingStack); c != 0 {
			return c
		}
		return strings.Compare(a.exportName, b.exportName)
	})

	var messages []string
	for _, key := range keys {
		stacks := slices.Sorted(slices.Values(externalReferences[key]))
		stackList := "'" + strings.Join(stacks, "', '") + "'"
		messages = append(messages, fmt.Sprintf("Stack '%s' exports '%s' which is imported by non-target stack(s) %s",
			key.exportingStack, key.exportName, stackList))
	}
	return fmt.Errorf("deletion would break dependencies for non-target stacks:\n%s", strings.Join(messages, "\n"))
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

	stack := &stacks[0]

	// If the stack is in the ROLLBACK_COMPLETE state, it is not possible to update the stack.
	if stack.StackStatus == types.StackStatusRollbackComplete {
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

	modifiedTemplate, changed, err := removeDeletionPolicyFromTemplate(template)
	if err != nil {
		return err
	}
	if changed {
		// Check if the template size exceeds the CloudFormation limit (51,200 bytes)
		const maxTemplateBodySize = 51200
		if len(modifiedTemplate) > maxTemplateBodySize {
			uploadResult, uploadErr := o.uploadTemplateToS3(ctx, stackName, &modifiedTemplate, stack)
			if uploadErr != nil {
				// no wrap because uploadTemplateToS3 already wraps the error
				return uploadErr
			}

			io.Logger.Info().Msgf("Created temporary S3 bucket for large template: %s", *uploadResult.BucketName)

			updateErr := o.client.UpdateStackWithTemplateURL(ctx, stackName, uploadResult.TemplateURL, stack.Parameters)

			// Ensure S3 cleanup happens even if UpdateStack fails (`updateErr != nil`)
			// Delete temporary S3 bucket and template immediately after UpdateStack completes (success or failure)
			if deleteErr := o.deleteTemplateFromS3(ctx, uploadResult.BucketName, uploadResult.Key); deleteErr != nil {
				// Log the error but don't fail the operation
				io.Logger.Warn().Msgf("Failed to delete temporary S3 bucket and template (bucket: %s, key: %s). You may need to delete it manually: %v", *uploadResult.BucketName, *uploadResult.Key, deleteErr)
			} else {
				io.Logger.Info().Msgf("Deleted temporary S3 bucket: %s", *uploadResult.BucketName)
			}

			if updateErr != nil {
				return fmt.Errorf("TemplateS3UpdateError: failed to update stack with large template via S3: %w", updateErr)
			}
		} else {
			if err = o.client.UpdateStack(ctx, stackName, &modifiedTemplate, stack.Parameters); err != nil {
				return err
			}
		}
	}
	if len(nestedStacks) == 0 {
		return nil
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

func (o *CloudFormationStackOperator) uploadTemplateToS3(ctx context.Context, stackName *string, template *string, stack *types.Stack) (*S3UploadResult, error) {
	accountID := ""
	if stack != nil && stack.StackId != nil {
		arnParts := strings.Split(*stack.StackId, ":")
		if len(arnParts) >= 5 {
			accountID = arnParts[4]
		}
	}

	if accountID == "" {
		return nil, fmt.Errorf("TemplateS3UploadError: failed to extract account ID from stack ARN")
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	bucketName := fmt.Sprintf("delstack-templates-%s-%s-%s", accountID, o.config.Region, timestamp)

	// Ensure bucket cleanup if upload fails (only after bucket is created)
	bucketCreated := false
	defer func() {
		if bucketCreated {
			// If we return early due to error, clean up the bucket
			if cleanupErr := o.s3Client.DeleteBucket(ctx, &bucketName); cleanupErr != nil {
				io.Logger.Warn().Msgf("Failed to cleanup temporary S3 bucket (bucket: %s) after upload error. You may need to delete it manually: %v", bucketName, cleanupErr)
			}
		}
	}()

	if err := o.s3Client.CreateBucket(ctx, &bucketName); err != nil {
		return nil, fmt.Errorf("TemplateS3UploadError: failed to create S3 bucket: %w", err)
	}
	bucketCreated = true

	key := fmt.Sprintf("%s.template", *stackName)

	if err := o.s3Client.PutObject(ctx, &bucketName, &key, template); err != nil {
		return nil, fmt.Errorf("TemplateS3UploadError: failed to upload template to S3: %w", err)
	}

	// Success - don't cleanup bucket (it will be cleaned up by main defer)
	bucketCreated = false

	templateURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, o.config.Region, key)
	return &S3UploadResult{
		TemplateURL: &templateURL,
		BucketName:  &bucketName,
		Key:         &key,
	}, nil
}

func (o *CloudFormationStackOperator) deleteTemplateFromS3(ctx context.Context, bucketName *string, key *string) error {
	objectIdentifier := []s3types.ObjectIdentifier{
		{
			Key: key,
		},
	}
	errors, err := o.s3Client.DeleteObjects(ctx, bucketName, objectIdentifier)
	if err != nil {
		return fmt.Errorf("TemplateS3DeleteError: failed to delete temporary template from S3: %w", err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("TemplateS3DeleteError: failed to delete temporary template from S3: %v", errors)
	}

	if err := o.s3Client.DeleteBucket(ctx, bucketName); err != nil {
		return fmt.Errorf("TemplateS3DeleteError: failed to delete temporary S3 bucket: %w", err)
	}

	return nil
}
