package operations

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func DeleteBuckets(config aws.Config, resources []types.StackResourceSummary) error {
	// TODO: Concurrency Delete
	return nil
}
