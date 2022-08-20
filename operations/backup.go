package operations

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func DeleteBackups(config aws.Config, resources []types.StackResourceSummary) error {
	return nil
}
