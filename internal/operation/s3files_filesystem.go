package operation

import (
	"context"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*S3FilesFileSystemOperator)(nil)

type S3FilesFileSystemOperator struct {
	client    client.IS3Files
	resources []*types.StackResourceSummary
}

func NewS3FilesFileSystemOperator(client client.IS3Files) *S3FilesFileSystemOperator {
	return &S3FilesFileSystemOperator{
		client:    client,
		resources: []*types.StackResourceSummary{},
	}
}

func (o *S3FilesFileSystemOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *S3FilesFileSystemOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *S3FilesFileSystemOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteS3FilesFileSystem(ctx, resource.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (o *S3FilesFileSystemOperator) DeleteS3FilesFileSystem(ctx context.Context, fileSystemId *string) error {
	exists, err := o.client.CheckFileSystemExists(ctx, fileSystemId)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return o.client.DeleteFileSystem(ctx, fileSystemId)
}
