package operation

import (
	"context"
	"runtime"

	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*S3TablesNamespaceOperator)(nil)

type S3TablesNamespaceOperator struct {
	client    client.IS3Tables
	resources []*cfntypes.StackResourceSummary
}

func NewS3TablesNamespaceOperator(client client.IS3Tables) *S3TablesNamespaceOperator {
	return &S3TablesNamespaceOperator{
		client:    client,
		resources: []*cfntypes.StackResourceSummary{},
	}
}

func (o *S3TablesNamespaceOperator) AddResource(resource *cfntypes.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *S3TablesNamespaceOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *S3TablesNamespaceOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, namespace := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteS3TablesNamespace(ctx, namespace.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *S3TablesNamespaceOperator) DeleteS3TablesNamespace(ctx context.Context, namespaceArn *string) error {
	// PhysicalResourceId is ARN format: arn:aws:s3tables:region:account-id:bucket/table-bucket-name|namespace-name
	// Extract tableBucketARN and namespace from the ARN
	tableBucketARN, namespace, err := client.ParseS3TablesNamespaceArn(namespaceArn)
	if err != nil {
		return err
	}

	exists, err := o.client.CheckNamespaceExists(ctx, tableBucketARN, namespace)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	eg := errgroup.Group{}
	sem := semaphore.NewWeighted(SemaphoreWeight)

	var continuationToken *string
	for {
		select {
		case <-ctx.Done():
			return &client.ClientError{
				ResourceName: namespaceArn,
				Err:          ctx.Err(),
			}
		default:
		}

		output, err := o.client.ListTablesByPage(ctx, tableBucketARN, namespace, continuationToken)
		if err != nil {
			return err
		}
		if len(output.Tables) == 0 {
			break
		}

		for _, table := range output.Tables {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			eg.Go(func() error {
				defer sem.Release(1)
				if err := o.client.DeleteTable(ctx, table.Name, namespace, tableBucketARN); err != nil {
					return err
				}
				return nil
			})
		}

		continuationToken = output.ContinuationToken
		if continuationToken == nil {
			break
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return o.client.DeleteNamespace(ctx, namespace, tableBucketARN)
}
