package operation

import "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

type IOperator interface {
	AddResource(resource *types.StackResourceSummary)
	GetResourcesLength() int
	DeleteResources() error
}
