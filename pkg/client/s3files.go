//go:generate go tool mockgen -source=$GOFILE -destination=s3files_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/aws/aws-sdk-go-v2/service/s3files/types"
)

type IS3Files interface {
	DeleteFileSystem(ctx context.Context, fileSystemId *string) error
	CheckFileSystemExists(ctx context.Context, fileSystemId *string) (bool, error)
}

var _ IS3Files = (*S3Files)(nil)

type S3Files struct {
	client *s3files.Client
}

func NewS3Files(client *s3files.Client) *S3Files {
	return &S3Files{
		client,
	}
}

// DeleteFileSystem deletes the file system even if it has data pending export to S3.
func (s *S3Files) DeleteFileSystem(ctx context.Context, fileSystemId *string) error {
	input := &s3files.DeleteFileSystemInput{
		FileSystemId: fileSystemId,
		ForceDelete:  aws.Bool(true),
	}

	_, err := s.client.DeleteFileSystem(ctx, input)
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: fileSystemId,
			Err:          err,
		}
	}
	return nil
}

func (s *S3Files) CheckFileSystemExists(ctx context.Context, fileSystemId *string) (bool, error) {
	input := &s3files.GetFileSystemInput{
		FileSystemId: fileSystemId,
	}

	_, err := s.client.GetFileSystem(ctx, input)
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, &ClientError{
			ResourceName: fileSystemId,
			Err:          err,
		}
	}
	return true, nil
}
