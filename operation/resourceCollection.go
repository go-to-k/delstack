package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type ResourceCollection struct {
	config             aws.Config
	StackName          string
	OperatorCollection *OperatorCollection
}

func NewResourceCollection(config aws.Config, stackName string, stackResourceSummaries []types.StackResourceSummary) *ResourceCollection {
	operatorCollection := NewOperatorCollection(config)

	return &ResourceCollection{
		config:             config,
		StackName:          stackName,
		OperatorCollection: operatorCollection,
	}
}

func (collection *ResourceCollection) CheckResourceCounts() error {
	return collection.OperatorCollection.CheckResourceCounts(collection.StackName)
}

func (collection *ResourceCollection) DeleteResourceCollection() error {
	operatorList := collection.OperatorCollection.GetOperatorList()

	// TODO: Concurrency deletion of failed resources
	for _, operator := range *operatorList {
		if err := operator.DeleteResources(); err != nil {
			return err
		}
	}

	return nil
}
