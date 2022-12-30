package client

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const s3DeleteObjectsSizeLimit = 1000

type IS3 interface {
	DeleteBucket(ctx context.Context, bucketName *string) error
	DeleteObjects(ctx context.Context, bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error)
	ListObjectVersions(ctx context.Context, bucketName *string) ([]types.ObjectIdentifier, error)
	CheckBucketExists(ctx context.Context, bucketName *string) (bool, error)
}

var _ IS3 = (*S3)(nil)

type IS3SDKClient interface {
	DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error)
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
}

type S3 struct {
	client IS3SDKClient
}

func NewS3(client IS3SDKClient) *S3 {
	return &S3{
		client,
	}
}

func (s3Client *S3) DeleteBucket(ctx context.Context, bucketName *string) error {
	input := &s3.DeleteBucketInput{
		Bucket: bucketName,
	}

	_, err := s3Client.client.DeleteBucket(ctx, input)

	return err
}

func (s3Client *S3) DeleteObjects(ctx context.Context, bucketName *string, objects []types.ObjectIdentifier, sleepTimeSec int) ([]types.Error, error) {
	eg, ctx := errgroup.WithContext(ctx)
	outputsCh := make(chan *s3.DeleteObjectsOutput, int64(runtime.NumCPU()))
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	wg := sync.WaitGroup{}

	errors := []types.Error{}
	nextObjects := make([]types.ObjectIdentifier, len(objects))
	copy(nextObjects, objects)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for outputErrors := range outputsCh {
			outputErrors := outputErrors
			if len(outputErrors.Errors) > 0 {
				errors = append(errors, outputErrors.Errors...)
			}
		}
	}()

	for {
		inputObjects := []types.ObjectIdentifier{}

		if len(nextObjects) > s3DeleteObjectsSizeLimit {
			inputObjects = append(inputObjects, nextObjects[:s3DeleteObjectsSizeLimit]...)
			nextObjects = nextObjects[s3DeleteObjectsSizeLimit:]
		} else {
			inputObjects = append(inputObjects, nextObjects...)
			nextObjects = nil
		}

		input := &s3.DeleteObjectsInput{
			Bucket: bucketName,
			Delete: &types.Delete{
				Objects: inputObjects,
				Quiet:   *aws.Bool(true),
			},
		}

		sem.Acquire(ctx, 1)
		eg.Go(func() error {
			defer sem.Release(1)

			return s3Client.deleteObjectsWithRetry(ctx, input, outputsCh, sleepTimeSec, bucketName)
		})

		if len(nextObjects) == 0 {
			break
		}
	}

	go func() {
		eg.Wait()
		close(outputsCh)
	}()

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// wait errors set before access an errors var at below return (for race)
	wg.Wait()

	return errors, nil
}

func (s3Client *S3) deleteObjectsWithRetry(
	ctx context.Context,
	input *s3.DeleteObjectsInput,
	outputsCh chan *s3.DeleteObjectsOutput,
	sleepTimeSec int,
	bucketName *string,
) error {
	var (
		output     *s3.DeleteObjectsOutput
		err        error
		retryCount int
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		output, err = s3Client.client.DeleteObjects(ctx, input)
		if err != nil && strings.Contains(err.Error(), "api error SlowDown") {
			retryCount++
			if err := WaitForRetry(retryCount, sleepTimeSec, bucketName, err); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}

		outputsCh <- output
		break
	}

	return nil
}

func (s3Client *S3) ListObjectVersions(ctx context.Context, bucketName *string) ([]types.ObjectIdentifier, error) {
	var keyMarker *string
	var versionIdMarker *string
	objectIdentifiers := []types.ObjectIdentifier{}

	for {
		input := &s3.ListObjectVersionsInput{
			Bucket:          bucketName,
			KeyMarker:       keyMarker,
			VersionIdMarker: versionIdMarker,
		}

		output, err := s3Client.client.ListObjectVersions(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, version := range output.Versions {
			objectIdentifier := types.ObjectIdentifier{
				Key:       version.Key,
				VersionId: version.VersionId,
			}
			objectIdentifiers = append(objectIdentifiers, objectIdentifier)
		}

		for _, deleteMarker := range output.DeleteMarkers {
			objectIdentifier := types.ObjectIdentifier{
				Key:       deleteMarker.Key,
				VersionId: deleteMarker.VersionId,
			}
			objectIdentifiers = append(objectIdentifiers, objectIdentifier)
		}

		keyMarker = output.NextKeyMarker
		versionIdMarker = output.NextVersionIdMarker

		if keyMarker == nil && versionIdMarker == nil {
			break
		}
	}

	return objectIdentifiers, nil
}

func (s3Client *S3) CheckBucketExists(ctx context.Context, bucketName *string) (bool, error) {
	input := &s3.ListBucketsInput{}

	output, err := s3Client.client.ListBuckets(ctx, input)
	if err != nil {
		return false, err
	}

	for _, bucket := range output.Buckets {
		if *bucket.Name == *bucketName {
			return true, nil
		}
	}

	return false, nil
}
