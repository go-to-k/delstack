package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/option"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type S3 struct {
	client *s3.Client
}

func NewS3(config aws.Config) *S3 {
	client := s3.NewFromConfig(config)
	return &S3{
		client,
	}
}

func (s3Client *S3) DeleteBucket(bucketName *string) error {
	input := &s3.DeleteBucketInput{
		Bucket: bucketName,
	}

	_, err := s3Client.client.DeleteBucket(context.TODO(), input)

	return err
}

func (s3Client *S3) DeleteObjects(bucketName *string, objects []types.ObjectIdentifier) ([]types.Error, error) {
	eg, ctx := errgroup.WithContext(context.Background())
	outputsCh := make(chan *s3.DeleteObjectsOutput)
	sem := semaphore.NewWeighted(int64(option.ConcurrencyNum))

	errors := []types.Error{}
	nextObjects := make([]types.ObjectIdentifier, len(objects))
	copy(nextObjects, objects)

	for {
		inputObjects := []types.ObjectIdentifier{}

		if len(nextObjects) > 1000 {
			inputObjects = append(inputObjects, nextObjects[:1000]...)
			nextObjects = nextObjects[1000:]
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

		eg.Go(func() error {
			sem.Acquire(context.Background(), 1)
			defer sem.Release(1)

			var (
				output     *s3.DeleteObjectsOutput
				err        error
				retryCount int
			)
			for {
				output, err = s3Client.client.DeleteObjects(context.TODO(), input)
				if err != nil && strings.Contains(err.Error(), "api error SlowDown") {
					retryCount++
					if retryCount > option.MaxRetryCount {
						logger.Logger.Warn().Msg(err.Error() + "\nRetried over " + strconv.Itoa(option.MaxRetryCount) + " but failed. ")
						return fmt.Errorf("RetryCountOverError: %v", bucketName)
					}

					logger.Logger.Warn().Msg(err.Error() + "\nRetrying...")
					time.Sleep(time.Second * 10)
					continue
				}
				if err != nil {
					return err
				}
				break
			}

			select {
			case outputsCh <- output:
			case <-ctx.Done():
				return ctx.Err()
			}

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

	for outputErrors := range outputsCh {
		if len(outputErrors.Errors) > 0 {
			errors = append(errors, outputErrors.Errors...)
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return errors, nil
}

func (s3Client *S3) ListObjectVersions(bucketName *string) ([]types.ObjectIdentifier, error) {
	var keyMarker *string
	var versionIdMarker *string
	objectIdentifiers := []types.ObjectIdentifier{}

	for {
		input := &s3.ListObjectVersionsInput{
			Bucket:          bucketName,
			KeyMarker:       keyMarker,
			VersionIdMarker: versionIdMarker,
		}

		output, err := s3Client.client.ListObjectVersions(context.TODO(), input)
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
