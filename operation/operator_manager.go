package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"golang.org/x/sync/errgroup"
)

type OperatorManager struct {
	operatorCollection *OperatorCollection
}

func NewOperatorManager(config aws.Config, stackName *string, stackResourceSummaries []types.StackResourceSummary) *OperatorManager {
	operatorFactory := NewOperatorFactory(config)

	return &OperatorManager{
		operatorCollection: NewOperatorCollection(config, operatorFactory, stackName, stackResourceSummaries),
	}
}

func (operatorManager *OperatorManager) getOperatorResourcesLength() int {
	var length int
	for _, operator := range operatorManager.operatorCollection.GetOperatorList() {
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

func (operatorManager *OperatorManager) DeleteResourceCollection() error {
	var eg errgroup.Group

	for _, operator := range operatorManager.operatorCollection.GetOperatorList() {
		operator := operator
		eg.Go(func() error {
			return operator.DeleteResources()
		})
	}

	return eg.Wait()
}
