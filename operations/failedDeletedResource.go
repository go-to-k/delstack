package operations

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type FailedDeletedResource struct {
	StackArray  []types.StackResourceSummary
	BucketArray []types.StackResourceSummary
	RoleArray   []types.StackResourceSummary
	ECRArray    []types.StackResourceSummary
	BackupArray []types.StackResourceSummary
	CustomArray []types.StackResourceSummary
}

func DeleteFailedDeletedResource(config aws.Config, failedDeletedResource FailedDeletedResource) error {
	// TODO: Concurrency deletion of failed resources

	if err := DeleteStacks(config, failedDeletedResource.StackArray); err != nil {
		return err
	}
	if err := DeleteBuckets(config, failedDeletedResource.BucketArray); err != nil {
		return err
	}
	if err := DeleteRoles(config, failedDeletedResource.RoleArray); err != nil {
		return err
	}
	if err := DeleteECRs(config, failedDeletedResource.ECRArray); err != nil {
		return err
	}
	if err := DeleteBackups(config, failedDeletedResource.ECRArray); err != nil {
		return err
	}
	if err := DeleteCustoms(config, failedDeletedResource.CustomArray); err != nil {
		return err
	}

	return nil
}
