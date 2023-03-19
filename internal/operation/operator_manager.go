//go:generate mockgen -source=$GOFILE -destination=./operator_manager_mock.go -package=$GOPACKAGE -write_package_comment=false
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

func (m *OperatorManager) SetOperatorCollection(stackName *string, stackResourceSummaries []types.StackResourceSummary) {
	m.operatorCollection.SetOperatorCollection(stackName, stackResourceSummaries)
}

func (m *OperatorManager) getOperatorResourcesLength() int {
	var length int
	for _, operator := range m.operatorCollection.GetOperators() {
		length += operator.GetResourcesLength()
	}
	return length
}

func (m *OperatorManager) CheckResourceCounts() error {
	logicalResourceIdsLength := len(m.operatorCollection.GetLogicalResourceIds())
	operatorResourcesLength := m.getOperatorResourcesLength()

	if logicalResourceIdsLength != operatorResourcesLength {
		return m.operatorCollection.RaiseUnsupportedResourceError()
	}

	return nil
}

func (m *OperatorManager) GetLogicalResourceIds() []string {
	return m.operatorCollection.GetLogicalResourceIds()
}

func (m *OperatorManager) DeleteResourceCollection(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, operator := range m.operatorCollection.GetOperators() {
		operator := operator
		eg.Go(func() error {
			return operator.DeleteResources(ctx)
		})
	}

	return eg.Wait()
}
