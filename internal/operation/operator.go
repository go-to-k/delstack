//go:generate mockgen -source=$GOFILE -destination=operator_mock.go -package=$GOPACKAGE -write_package_comment=false
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
