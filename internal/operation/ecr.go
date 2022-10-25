package operation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/option"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*EcrOperator)(nil)

type EcrOperator struct {
	client    client.IEcr
	resources []*types.StackResourceSummary
}

func NewEcrOperator(client client.IEcr) *EcrOperator {
	return &EcrOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (operator *EcrOperator) AddResource(resource *types.StackResourceSummary) {
	operator.resources = append(operator.resources, resource)
}

func (operator *EcrOperator) GetResourcesLength() int {
	return len(operator.resources)
}

func (operator *EcrOperator) DeleteResources() error {
	var eg errgroup.Group
	sem := semaphore.NewWeighted(int64(option.ConcurrencyNum))

	for _, repository := range operator.resources {
		repository := repository
		eg.Go(func() (err error) {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			return operator.DeleteEcr(repository.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (operator *EcrOperator) DeleteEcr(repositoryName *string) error {
	exists, err := operator.client.CheckEcrExists(repositoryName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return operator.client.DeleteRepository(repositoryName)
}