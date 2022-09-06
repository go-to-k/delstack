package operation

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"golang.org/x/sync/errgroup"
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
	var semaphore = make(chan struct{}, CONCURRENCY_NUM)

	for _, repository := range operator.resources {
		repository := repository
		eg.Go(func() (err error) {
			semaphore <- struct{}{}

			if err := operator.DeleteECR(repository.PhysicalResourceId); err != nil {
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

func (operator *ECROperator) DeleteECR(repositoryName *string) error {
	if err := operator.client.DeleteRepository(repositoryName); err != nil {
		return err
	}

	return nil
}
