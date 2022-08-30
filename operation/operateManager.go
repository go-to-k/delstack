package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type OperatorManager struct {
	operatorCollection *OperatorCollection
}

func NewOperatorManager(config aws.Config, stackName string, stackResourceSummaries []types.StackResourceSummary) *OperatorManager {
	return &OperatorManager{
		operatorCollection: NewOperatorCollection(config, stackName, stackResourceSummaries),
	}
}

func (operatorManager *OperatorManager) getResourcesLengthFromOperatorList() int {
	var length int
	for _, operator := range *operatorManager.operatorCollection.GetOperatorList() {
		length += operator.GetResourcesLength()
	}
	return length
}

func (operatorManager *OperatorManager) CheckResourceCounts() error {
	collectionLength := operatorManager.getResourcesLengthFromOperatorList()

	if len(*operatorManager.operatorCollection.GetLogicalResourceIds()) != collectionLength {
		return operatorManager.operatorCollection.GetNotSupportedServicesError()
	}

	return nil
}

func (operatorManager *OperatorManager) GetLogicalResourceIds() *[]string {
	return operatorManager.operatorCollection.GetLogicalResourceIds()
}

func (operatorManager *OperatorManager) DeleteResourceCollection() error {
	// TODO: Concurrency deletion of failed resources
	for _, operator := range *operatorManager.operatorCollection.GetOperatorList() {
		if err := operator.DeleteResources(); err != nil {
			return err
		}
	}

	return nil
}
