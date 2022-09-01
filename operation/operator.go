package operation

import "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

type Operator interface {
	AddResources(resource *types.StackResourceSummary)
	GetResourcesLength() int
	DeleteResources() error
}
