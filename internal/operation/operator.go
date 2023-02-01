//go:generate mockgen -source=./operator.go -destination=./operator_mock.go -package=operation
package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type IOperator interface {
	AddResource(resource *types.StackResourceSummary)
	GetResourcesLength() int
	DeleteResources(ctx context.Context) error
}
