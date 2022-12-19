package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"golang.org/x/sync/errgroup"
)

type IOperatorManager interface {
	SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary)
	CheckResourceCounts() error
	GetLogicalResourceIds() []string
	DeleteResourceCollection(ctx context.Context) error
}

var _ IOperatorManager = (*OperatorManager)(nil)

type OperatorManager struct {
	operatorCollection IOperatorCollection
}

func NewOperatorManager(operatorCollection IOperatorCollection) *OperatorManager {
	return &OperatorManager{
		operatorCollection: operatorCollection,
	}
}

func (operatorManager *OperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
	operatorManager.operatorCollection.SetOperatorCollection(stackName, stackResourceSummaries)
}

func (operatorManager *OperatorManager) getOperatorResourcesLength() int {
	var length int
	for _, operator := range operatorManager.operatorCollection.GetOperators() {
		length += operator.GetResourcesLength()
	}
	return length
}

func (operatorManager *OperatorManager) CheckResourceCounts() error {
	logicalResourceIdsLength := len(operatorManager.operatorCollection.GetLogicalResourceIds())
	operatorResourcesLength := operatorManager.getOperatorResourcesLength()

	if logicalResourceIdsLength != operatorResourcesLength {
		return operatorManager.operatorCollection.RaiseUnsupportedResourceError()
	}

	return nil
}

func (operatorManager *OperatorManager) GetLogicalResourceIds() []string {
	return operatorManager.operatorCollection.GetLogicalResourceIds()
}

func (operatorManager *OperatorManager) DeleteResourceCollection(ctx context.Context) error {
	var eg errgroup.Group

	for _, operator := range operatorManager.operatorCollection.GetOperators() {
		operator := operator
		eg.Go(func() error {
			return operator.DeleteResources(ctx)
		})
	}

	return eg.Wait()
}
