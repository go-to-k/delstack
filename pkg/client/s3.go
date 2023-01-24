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

var SleepTimeSecForS3 = 10

type IS3 interface {
	DeleteBucket(ctx context.Context, bucketName *string) error
	DeleteObjects(ctx context.Context, bucketName *string, objects []types.ObjectIdentifier) ([]types.Error, error)
	ListObjectVersions(ctx context.Context, bucketName *string) ([]types.ObjectIdentifier, error)
	CheckBucketExists(ctx context.Context, bucketName *string) (bool, error)
}

var _ IS3 = (*S3)(nil)

type S3 struct {
	client *s3.Client
}

func NewS3(client *s3.Client) *S3 {
	return &S3{
		client,
	}
}

func (s *S3) DeleteBucket(ctx context.Context, bucketName *string) error {
	input := &s3.DeleteBucketInput{
		Bucket: bucketName,
	}

	_, err := s.client.DeleteBucket(ctx, input)

	return err
}

func (s *S3) DeleteObjects(ctx context.Context, bucketName *string, objects []types.ObjectIdentifier) ([]types.Error, error) {
	errors := []types.Error{}
	if len(objects) == 0 {
		return errors, nil
	}

	eg, ctx := errgroup.WithContext(ctx)
	outputsCh := make(chan *s3.DeleteObjectsOutput, int64(runtime.NumCPU()))
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	wg := sync.WaitGroup{}

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

		if err := sem.Acquire(ctx, 1); err != nil {
			return errors, err
		}
		eg.Go(func() error {
			defer sem.Release(1)

			retryable := func(err error) bool {
				return strings.Contains(err.Error(), "api error SlowDown")
			}

			output, err := Retry(
				&RetryInput[s3.DeleteObjectsInput, s3.DeleteObjectsOutput, s3.Options]{
					Ctx:              ctx,
					SleepTimeSec:     SleepTimeSecForS3,
					TargetResource:   bucketName,
					Input:            input,
					ApiCaller:        s.client.DeleteObjects,
					RetryableChecker: retryable,
				},
			)
			if err != nil {
				return err
			}

			outputsCh <- output
			return nil
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

func (s *S3) ListObjectVersions(ctx context.Context, bucketName *string) ([]types.ObjectIdentifier, error) {
	var keyMarker *string
	var versionIdMarker *string
	objectIdentifiers := []types.ObjectIdentifier{}

	for {
		select {
		case <-ctx.Done():
			return objectIdentifiers, ctx.Err()
		default:
		}

		input := &s3.ListObjectVersionsInput{
			Bucket:          bucketName,
			KeyMarker:       keyMarker,
			VersionIdMarker: versionIdMarker,
		}

		output, err := s.client.ListObjectVersions(ctx, input)
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

func (s *S3) CheckBucketExists(ctx context.Context, bucketName *string) (bool, error) {
	input := &s3.ListBucketsInput{}

	output, err := s.client.ListBuckets(ctx, input)
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
