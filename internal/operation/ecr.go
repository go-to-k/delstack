package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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

func (operator *EcrOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, repository := range operator.resources {
		repository := repository
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return operator.DeleteEcr(ctx, repository.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (operator *EcrOperator) DeleteEcr(ctx context.Context, repositoryName *string) error {
	exists, err := operator.client.CheckEcrExists(ctx, repositoryName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return operator.client.DeleteRepository(ctx, repositoryName)
}
