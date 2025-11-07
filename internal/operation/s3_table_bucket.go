package operation

import (
	"context"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Too Many Requests error often occurs, so limit the value
const SemaphoreWeightForS3Tables = 4

var _ IOperator = (*S3TableBucketOperator)(nil)

type S3TableBucketOperator struct {
	client    client.IS3Tables
	resources []*cfntypes.StackResourceSummary
}

func NewS3TableBucketOperator(client client.IS3Tables) *S3TableBucketOperator {
	return &S3TableBucketOperator{
		client:    client,
		resources: []*cfntypes.StackResourceSummary{},
	}
}

func (o *S3TableBucketOperator) AddResource(resource *cfntypes.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *S3TableBucketOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *S3TableBucketOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, bucket := range o.resources {
		bucket := bucket
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteS3TableBucket(ctx, bucket.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *S3TableBucketOperator) DeleteS3TableBucket(ctx context.Context, tableBucketArn *string) error {
	// PhysicalResourceId is ARN format: arn:aws:s3tables:region:account-id:bucket/table-bucket-name
	// Extract the bucket name from the ARN
	tableBucketName := tableBucketArn
	if tableBucketArn != nil {
		parts := strings.Split(*tableBucketArn, "/")
		if len(parts) > 0 {
			tableBucketName = aws.String(parts[len(parts)-1])
		}
	}

	exists, err := o.client.CheckTableBucketExists(ctx, tableBucketName)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	eg := errgroup.Group{}
	sem := semaphore.NewWeighted(SemaphoreWeightForS3Tables)
	var continuationToken *string
	for {
		select {
		case <-ctx.Done():
			return &client.ClientError{
				ResourceName: tableBucketArn,
				Err:          ctx.Err(),
			}
		default:
		}

		output, err := o.client.ListNamespacesByPage(
			ctx,
			tableBucketArn,
			continuationToken,
		)
		if err != nil {
			return err
		}
		if len(output.Namespaces) == 0 {
			break
		}

		for _, summary := range output.Namespaces {
			for _, namespace := range summary.Namespace {
				if err := sem.Acquire(ctx, 1); err != nil {
					return err
				}
				eg.Go(func() error {
					defer sem.Release(1)
					return o.deleteNamespace(ctx, tableBucketArn, aws.String(namespace))
				})
			}
		}

		continuationToken = output.ContinuationToken
		if continuationToken == nil {
			break
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if err := o.client.DeleteTableBucket(ctx, tableBucketArn); err != nil {
		return err
	}

	return nil
}

func (o *S3TableBucketOperator) deleteNamespace(
	ctx context.Context,
	tableBucketArn *string,
	namespace *string,
) error {
	eg := errgroup.Group{}
	sem := semaphore.NewWeighted(SemaphoreWeightForS3Tables)

	var continuationToken *string
	for {
		select {
		case <-ctx.Done():
			return &client.ClientError{
				ResourceName: tableBucketArn,
				Err:          ctx.Err(),
			}
		default:
		}

		output, err := o.client.ListTablesByPage(ctx, tableBucketArn, namespace, continuationToken)
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
				if err := o.client.DeleteTable(ctx, table.Name, namespace, tableBucketArn); err != nil {
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

	return o.client.DeleteNamespace(ctx, namespace, tableBucketArn)
}
