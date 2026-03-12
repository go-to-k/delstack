package preprocessor

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type IPreprocessor interface {
	Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error
}

// FilterResourcesByType filters stack resources by resource type
func FilterResourcesByType(resources []types.StackResourceSummary, resourceType string) []types.StackResourceSummary {
	var filtered []types.StackResourceSummary
	for _, resource := range resources {
		if aws.ToString(resource.ResourceType) == resourceType && resource.ResourceStatus != types.ResourceStatusDeleteComplete {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}
