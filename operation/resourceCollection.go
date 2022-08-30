package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type ResourceCollection struct {
	config             aws.Config
	StackName          string
	LogicalResourceIds *[]string
	OperatorList       *[]IOperator
	OperatorCollection *OperatorCollection
}

func NewResourceCollection(config aws.Config, stackName string, stackResourceSummaries []types.StackResourceSummary) *ResourceCollection {
	operatorCollection := NewOperatorCollection(config)
	logicalResourceIds := operatorCollection.GetLogicalResourceIds()
	operatorList := operatorCollection.GetOperatorList()

	return &ResourceCollection{
		config:             config,
		StackName:          stackName,
		LogicalResourceIds: logicalResourceIds,
		OperatorList:       operatorList,
		OperatorCollection: operatorCollection,
	}
}

func (collection *ResourceCollection) CheckResourceCounts() error {
	return collection.OperatorCollection.CheckResourceCounts(collection.StackName)
}

func (collection *ResourceCollection) DeleteResourceCollection() error {
	// TODO: Concurrency deletion of failed resources
	for _, operator := range *collection.OperatorList {
		if err := operator.DeleteResources(); err != nil {
			return err
		}
	}

	return nil
}
