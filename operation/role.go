package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func DeleteRoles(config aws.Config, resources []types.StackResourceSummary) error {
	// TODO: Concurrency Delete
	return nil
}
