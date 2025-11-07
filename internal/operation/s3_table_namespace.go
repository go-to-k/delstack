package operation

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Too Many Requests error often occurs, so limit the value
const SemaphoreWeightForS3TableNamespaces = 4

var _ IOperator = (*S3TableNamespaceOperator)(nil)

type S3TableNamespaceOperator struct {
	client    client.IS3Tables
	resources []*cfntypes.StackResourceSummary
}

func NewS3TableNamespaceOperator(client client.IS3Tables) *S3TableNamespaceOperator {
	return &S3TableNamespaceOperator{
		client:    client,
		resources: []*cfntypes.StackResourceSummary{},
	}
}

func (o *S3TableNamespaceOperator) AddResource(resource *cfntypes.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *S3TableNamespaceOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *S3TableNamespaceOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, namespace := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteS3TableNamespace(ctx, namespace.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *S3TableNamespaceOperator) DeleteS3TableNamespace(ctx context.Context, namespaceArn *string) error {
	// PhysicalResourceId is ARN format: arn:aws:s3tables:region:account-id:bucket/table-bucket-name|namespace-name
	// Extract tableBucketARN and namespace from the ARN
	tableBucketARN, namespace, err := o.parseS3TableNamespaceArn(namespaceArn)
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
	sem := semaphore.NewWeighted(SemaphoreWeightForS3TableNamespaces)

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

// parseS3TableNamespaceArn parses S3 Tables Namespace ARN and returns tableBucketARN and namespace
// ARN format: arn:aws:s3tables:region:account-id:bucket/table-bucket-name|namespace-name
func (o *S3TableNamespaceOperator) parseS3TableNamespaceArn(namespaceArn *string) (*string, *string, error) {
	if namespaceArn == nil {
		return nil, nil, &client.ClientError{
			Err: fmt.Errorf("DeleteS3TableNamespaceError: namespace ARN is nil"),
		}
	}

	// Split by "|" to separate table bucket ARN and namespace name
	parts := strings.Split(*namespaceArn, "|")
	if len(parts) != 2 {
		return nil, nil, &client.ClientError{
			Err: fmt.Errorf("DeleteS3TableNamespaceError: invalid namespace ARN format: %s", *namespaceArn),
		}
	}

	// The table bucket ARN is already in the correct format: arn:aws:s3tables:region:account-id:bucket/table-bucket-name
	tableBucketARN := aws.String(parts[0])
	namespace := aws.String(parts[1])

	return tableBucketARN, namespace, nil
}
