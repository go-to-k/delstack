package operation

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var _ IOperator = (*S3BucketOperator)(nil)

type S3BucketOperator struct {
	client               client.IS3
	resources            []*cfntypes.StackResourceSummary
	directoryBucketsFlag bool
}

func NewS3BucketOperator(client client.IS3) *S3BucketOperator {
	return &S3BucketOperator{
		client:    client,
		resources: []*cfntypes.StackResourceSummary{},
	}
}

func (o *S3BucketOperator) AddResource(resource *cfntypes.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *S3BucketOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *S3BucketOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, bucket := range o.resources {
		bucket := bucket
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			return o.DeleteS3Bucket(ctx, bucket.PhysicalResourceId)
		})
	}

	return eg.Wait()
}

func (o *S3BucketOperator) DeleteS3Bucket(ctx context.Context, bucketName *string) error {
	exists, err := o.client.CheckBucketExists(ctx, bucketName, o.directoryBucketsFlag)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	eg := errgroup.Group{}
	errorStr := ""
	errorsCount := 0
	errorsMtx := sync.Mutex{}
	var keyMarker *string
	var versionIdMarker *string
	for {
		var objects []s3types.ObjectIdentifier

		// ListObjectVersions/ListObjectsV2 API can only retrieve up to 1000 items, so it is good to pass it
		// directly to DeleteObjects, which can only delete up to 1000 items.
		objects, keyMarker, versionIdMarker, err = o.client.ListObjectsOrVersionsByPage(
			ctx,
			bucketName,
			keyMarker,
			versionIdMarker,
			o.directoryBucketsFlag,
		)
		if err != nil {
			return err
		}

		if len(objects) == 0 {
			break
		}

		eg.Go(func() error {
			// One DeleteObjects is executed for each loop of the List, and it usually ends during
			// the next loop. Therefore, there seems to be no throttling concern, so the number of
			// parallels is not limited by semaphore. (Throttling occurs at about 3500 deletions
			// per second.)
			gotErrors, err := o.client.DeleteObjects(ctx, bucketName, objects)
			if err != nil {
				return err
			}

			if len(gotErrors) > 0 {
				errorsMtx.Lock()
				errorsCount += len(gotErrors)
				for _, error := range gotErrors {
					errorStr += fmt.Sprintf("\nBucketName: %v\n", *bucketName)
					errorStr += fmt.Sprintf("Code: %v\n", *error.Code)
					errorStr += fmt.Sprintf("Key: %v\n", *error.Key)
					errorStr += fmt.Sprintf("VersionId: %v\n", *error.VersionId)
					errorStr += fmt.Sprintf("Message: %v\n", *error.Message)
				}
				errorsMtx.Unlock()
			}

			return nil
		})

		if keyMarker == nil && versionIdMarker == nil {
			break
		}
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if errorsCount > 0 {
		// The error is from `DeleteObjectsOutput.Errors`, not `err`.
		// However, we want to treat it as an error, so we use `client.ClientError`.
		return &client.ClientError{
			ResourceName: bucketName,
			Err:          fmt.Errorf("DeleteObjectsError: %v objects with errors were found. %v", errorsCount, errorStr),
		}
	}

	if err := o.client.DeleteBucket(ctx, bucketName); err != nil {
		return err
	}

	return nil
}

func (o *S3BucketOperator) SetDirectoryBucketsFlag() {
	o.directoryBucketsFlag = true
}
