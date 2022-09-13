package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ Operator = (*ECROperator)(nil)

type ECROperator struct {
	client    *client.ECR
	resources []*types.StackResourceSummary
}

func NewECROperator(config aws.Config) *ECROperator {
	client := client.NewECR(config)
	return &ECROperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *ECROperator) AddResources(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *ECROperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *ECROperator) DeleteResources() error {
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(option.CONCURRENCY_NUM))

	for _, repository := range operator.resources {
		repository := repository
		eg.Go(func() (err error) {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			return operator.DeleteECR(repository.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (operator *ECROperator) DeleteECR(repositoryName *string) error {
	return operator.client.DeleteRepository(repositoryName)
}
