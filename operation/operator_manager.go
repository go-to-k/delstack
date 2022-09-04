package operation

import (
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"golang.org/x/sync/errgroup"
)

var CONCURRENCY_NUM = runtime.NumCPU()

type OperatorManager struct {
	operatorCollection *OperatorCollection
}

func NewOperatorManager(config aws.Config, stackName *string, stackResourceSummaries []types.StackResourceSummary) *OperatorManager {
	return &OperatorManager{
		operatorCollection: NewOperatorCollection(config, stackName, stackResourceSummaries),
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
		return operatorManager.operatorCollection.RaiseNotSupportedServicesError()
	}

	return nil
}

func (operatorManager *OperatorManager) GetLogicalResourceIds() []string {
	return operatorManager.operatorCollection.GetLogicalResourceIds()
}

func (operatorManager *OperatorManager) DeleteResourceCollection() error {
	var eg errgroup.Group
	semaphore := make(chan struct{}, CONCURRENCY_NUM)

	for _, operator := range operatorManager.operatorCollection.GetOperatorList() {
		operator := operator
		eg.Go(func() error {
			semaphore <- struct{}{}
			if err := operator.DeleteResources(); err != nil {
				return err
			}
			<-semaphore
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}
