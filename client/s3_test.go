package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-to-k/delstack/logger"
)

var _ IS3SDKClient = (*mockS3SDKClient)(nil)
var _ IS3SDKClient = (*errorMockS3SDKClient)(nil)
var _ IS3SDKClient = (*apiErrorMockS3SDKClient)(nil)
var _ IS3SDKClient = (*outputErrorForDeleteObjectsMockS3SDKClient)(nil)
var _ IS3SDKClient = (*emptyMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*versionsMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*deleteMarkersMockForListObjectVersionsS3SDKClient)(nil)

var sleepTimeSecForS3 = 1

/*
	Mocks for SDK Client
*/
type mockS3SDKClient struct{}

func NewMockS3SDKClient() *mockS3SDKClient {
	return &mockS3SDKClient{}
}

func (m *mockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *mockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, nil
}

func (m *mockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{
				Key:       aws.String("KeyForVersions"),
				VersionId: aws.String("VersionIdForVersions"),
			},
		},
		DeleteMarkers: []types.DeleteMarkerEntry{
			{
				Key:       aws.String("KeyForDeleteMarkers"),
				VersionId: aws.String("VersionIdForDeleteMarkers"),
			},
		},
	}
	return output, nil
}

type errorMockS3SDKClient struct{}

func NewErrorMockS3SDKClient() *errorMockS3SDKClient {
	return &errorMockS3SDKClient{}
}

func (m *errorMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, fmt.Errorf("DeleteBucketError")
}

func (m *errorMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, fmt.Errorf("DeleteObjectsError")
}

func (m *errorMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, fmt.Errorf("ListObjectVersionsError")
}

type apiErrorMockS3SDKClient struct{}

func NewApiErrorMockS3SDKClient() *apiErrorMockS3SDKClient {
	return &apiErrorMockS3SDKClient{}
}

func (m *apiErrorMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, fmt.Errorf("api error SlowDown")
}

func (m *apiErrorMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors:  []types.Error{},
	}
	return output, fmt.Errorf("api error SlowDown")
}

func (m *apiErrorMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, fmt.Errorf("api error SlowDown")
}

type outputErrorForDeleteObjectsMockS3SDKClient struct{}

func NewOutputErrorForDeleteObjectsMockS3SDKClient() *outputErrorForDeleteObjectsMockS3SDKClient {
	return &outputErrorForDeleteObjectsMockS3SDKClient{}
}

func (m *outputErrorForDeleteObjectsMockS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *outputErrorForDeleteObjectsMockS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	output := &s3.DeleteObjectsOutput{
		Deleted: []types.DeletedObject{},
		Errors: []types.Error{
			{
				Key:       aws.String("Key"),
				Code:      aws.String("Code"),
				Message:   aws.String("Message"),
				VersionId: aws.String("VersionId"),
			},
		},
	}
	return output, nil
}

func (m *outputErrorForDeleteObjectsMockS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	return nil, nil
}

type emptyMockForListObjectVersionsS3SDKClient struct{}

func NewEmptyMockForListObjectVersionsS3SDKClient() *emptyMockForListObjectVersionsS3SDKClient {
	return &emptyMockForListObjectVersionsS3SDKClient{}
}

func (m *emptyMockForListObjectVersionsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *emptyMockForListObjectVersionsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *emptyMockForListObjectVersionsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions:      []types.ObjectVersion{},
		DeleteMarkers: []types.DeleteMarkerEntry{},
	}
	return output, nil
}

type versionsMockForListObjectVersionsS3SDKClient struct{}

func NewVersionsMockForListObjectVersionsS3SDKClient() *versionsMockForListObjectVersionsS3SDKClient {
	return &versionsMockForListObjectVersionsS3SDKClient{}
}

func (m *versionsMockForListObjectVersionsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *versionsMockForListObjectVersionsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *versionsMockForListObjectVersionsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{
			{
				Key:       aws.String("KeyForVersions"),
				VersionId: aws.String("VersionIdForVersions"),
			},
		},
		DeleteMarkers: []types.DeleteMarkerEntry{},
	}
	return output, nil
}

type deleteMarkersMockForListObjectVersionsS3SDKClient struct{}

func NewDeleteMarkersMockForListObjectVersionsS3SDKClient() *deleteMarkersMockForListObjectVersionsS3SDKClient {
	return &deleteMarkersMockForListObjectVersionsS3SDKClient{}
}

func (m *deleteMarkersMockForListObjectVersionsS3SDKClient) DeleteBucket(ctx context.Context, params *s3.DeleteBucketInput, optFns ...func(*s3.Options)) (*s3.DeleteBucketOutput, error) {
	return nil, nil
}

func (m *deleteMarkersMockForListObjectVersionsS3SDKClient) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return nil, nil
}

func (m *deleteMarkersMockForListObjectVersionsS3SDKClient) ListObjectVersions(ctx context.Context, params *s3.ListObjectVersionsInput, optFns ...func(*s3.Options)) (*s3.ListObjectVersionsOutput, error) {
	output := &s3.ListObjectVersionsOutput{
		Versions: []types.ObjectVersion{},
		DeleteMarkers: []types.DeleteMarkerEntry{
			{
				Key:       aws.String("KeyForDeleteMarkers"),
				VersionId: aws.String("VersionIdForDeleteMarkers"),
			},
		},
	}
	return output, nil
}

/*
	Test Cases
*/
func TestDeleteBucket(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()

	type args struct {
		ctx        context.Context
		bucketName *string
		client     IS3SDKClient
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete bucket successfully",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete bucket failure",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     errorMock,
			},
			want:    fmt.Errorf("DeleteBucketError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s3Client := NewS3(tt.args.client)

			err := s3Client.DeleteBucket(tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err, tt.want)
			}
		})
	}
}

func TestDeleteObjects(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()
	apiErrorMock := NewApiErrorMockS3SDKClient()
	outputErrorMock := NewOutputErrorForDeleteObjectsMockS3SDKClient()

	objectsOverLimit := []types.ObjectIdentifier{}
	for i := 0; i <= 1000; i++ {
		objectsOverLimit = append(objectsOverLimit, types.ObjectIdentifier{
			Key:       aws.String("Key"),
			VersionId: aws.String("VersionId"),
		})
	}

	type args struct {
		ctx        context.Context
		bucketName *string
		objects    []types.ObjectIdentifier
		client     IS3SDKClient
	}

	type want struct {
		output []types.Error
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "delete objects successfully",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				client: mock,
			},
			want: want{
				output: []types.Error{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects over limit successfully",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				objects:    objectsOverLimit,
				client:     mock,
			},
			want: want{
				output: []types.Error{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects failure",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				client: errorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("DeleteObjectsError"),
			},
			wantErr: true,
		},
		{
			name: "delete objects failure for api error",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				client: apiErrorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("RetryCountOverError: test"),
			},
			wantErr: true,
		},
		{
			name: "delete objects failure for output errors",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				objects: []types.ObjectIdentifier{
					{
						Key:       aws.String("Key"),
						VersionId: aws.String("VersionId"),
					},
				},
				client: outputErrorMock,
			},
			want: want{
				output: []types.Error{
					{
						Key:       aws.String("Key"),
						Code:      aws.String("Code"),
						Message:   aws.String("Message"),
						VersionId: aws.String("VersionId"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s3Client := NewS3(tt.args.client)

			output, err := s3Client.DeleteObjects(tt.args.bucketName, tt.args.objects, sleepTimeSecForS3)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}

func TestListObjectVersions(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()
	emptyMock := NewEmptyMockForListObjectVersionsS3SDKClient()
	versionsMock := NewVersionsMockForListObjectVersionsS3SDKClient()
	deleteMarkersMock := NewDeleteMarkersMockForListObjectVersionsS3SDKClient()

	type args struct {
		ctx        context.Context
		bucketName *string
		client     IS3SDKClient
	}

	type want struct {
		output []types.ObjectIdentifier
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "list objects versions successfully",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     mock,
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForVersions"),
						VersionId: aws.String("VersionIdForVersions"),
					},
					{
						Key:       aws.String("KeyForDeleteMarkers"),
						VersionId: aws.String("VersionIdForDeleteMarkers"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions failure",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     errorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("ListObjectVersionsError"),
			},
			wantErr: true,
		},
		{
			name: "list objects versions successfully(empty)",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     emptyMock,
			},
			want: want{
				output: []types.ObjectIdentifier{},
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions successfully(versions only)",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     versionsMock,
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForVersions"),
						VersionId: aws.String("VersionIdForVersions"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "list objects versions successfully(delete markers only)",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     deleteMarkersMock,
			},
			want: want{
				output: []types.ObjectIdentifier{
					{
						Key:       aws.String("KeyForDeleteMarkers"),
						VersionId: aws.String("VersionIdForDeleteMarkers"),
					},
				},
				err: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s3Client := NewS3(tt.args.client)

			output, err := s3Client.ListObjectVersions(tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.output) {
				t.Errorf("output = %#v, want %#v", output, tt.want.output)
			}
		})
	}
}
