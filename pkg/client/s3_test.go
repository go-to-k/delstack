package client

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var _ IS3SDKClient = (*MockS3SDKClient)(nil)
var _ IS3SDKClient = (*ErrorMockS3SDKClient)(nil)
var _ IS3SDKClient = (*ApiErrorMockS3SDKClient)(nil)
var _ IS3SDKClient = (*OutputErrorForDeleteObjectsMockS3SDKClient)(nil)
var _ IS3SDKClient = (*EmptyMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*VersionsMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*DeleteMarkersMockForListObjectVersionsS3SDKClient)(nil)
var _ IS3SDKClient = (*NotExistsMockForListBucketsS3SDKClient)(nil)

const sleepTimeSecForS3 = 1

/*
	Test Cases
*/

func TestS3_DeleteBucket(t *testing.T) {
	ctx := context.Background()
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

			err := s3Client.DeleteBucket(tt.args.ctx, tt.args.bucketName)
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

func TestS3_DeleteObjects(t *testing.T) {
	ctx := context.Background()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()
	apiErrorMock := NewApiErrorMockS3SDKClient()
	outputErrorMock := NewOutputErrorForDeleteObjectsMockS3SDKClient()

	objectsOverLimit := []types.ObjectIdentifier{}
	s3DeleteObjectsSizeOverLimit := s3DeleteObjectsSizeLimit*int(runtime.NumCPU())*2 + 1 // loop over cpu core size for channel waiting when next loop
	for i := 0; i < s3DeleteObjectsSizeOverLimit; i++ {
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
			name: "delete objects successfully if zero objects",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				objects:    []types.ObjectIdentifier{},
				client:     mock,
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
				err:    fmt.Errorf("RetryCountOverError: test, api error SlowDown\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
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

			output, err := s3Client.DeleteObjects(tt.args.ctx, tt.args.bucketName, tt.args.objects, sleepTimeSecForS3)
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

func TestS3_deleteObjectsWithRetry(t *testing.T) {
	ctx := context.Background()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()
	apiErrorMock := NewApiErrorMockS3SDKClient()
	outputErrorMock := NewOutputErrorForDeleteObjectsMockS3SDKClient()

	objectsOverLimit := []types.ObjectIdentifier{}
	s3DeleteObjectsSizeOverLimit := s3DeleteObjectsSizeLimit*int(runtime.NumCPU())*2 + 1 // loop over cpu core size for channel waiting when next loop
	for i := 0; i < s3DeleteObjectsSizeOverLimit; i++ {
		objectsOverLimit = append(objectsOverLimit, types.ObjectIdentifier{
			Key:       aws.String("Key"),
			VersionId: aws.String("VersionId"),
		})
	}

	type args struct {
		ctx        context.Context
		input      *s3.DeleteObjectsInput
		bucketName *string
		client     IS3SDKClient
	}

	type want struct {
		output *s3.DeleteObjectsOutput
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
				ctx: ctx,
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
				client:     mock,
			},
			want: want{
				output: &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{},
					Errors:  []types.Error{},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects over limit successfully",
			args: args{
				ctx: ctx,
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: objectsOverLimit,
						Quiet:   *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
				client:     mock,
			},
			want: want{
				output: &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{},
					Errors:  []types.Error{},
				},
				err: nil,
			},
			wantErr: false,
		},
		{
			name: "delete objects failure",
			args: args{
				ctx: ctx,
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
				client:     errorMock,
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
				ctx: ctx,
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
				client:     apiErrorMock,
			},
			want: want{
				output: nil,
				err:    fmt.Errorf("RetryCountOverError: test, api error SlowDown\nRetryCount(" + strconv.Itoa(maxRetryCount) + ") over, but failed to delete. "),
			},
			wantErr: true,
		},
		{
			name: "delete objects failure for output errors",
			args: args{
				ctx: ctx,
				input: &s3.DeleteObjectsInput{
					Bucket: aws.String("test"),
					Delete: &types.Delete{
						Objects: []types.ObjectIdentifier{{
							Key:       aws.String("Key"),
							VersionId: aws.String("VersionId")},
						},
						Quiet: *aws.Bool(true),
					},
				},
				bucketName: aws.String("test"),
				client:     outputErrorMock,
			},
			want: want{
				output: &s3.DeleteObjectsOutput{
					Deleted: []types.DeletedObject{},
					Errors: []types.Error{
						{
							Key:       aws.String("Key"),
							Code:      aws.String("Code"),
							Message:   aws.String("Message"),
							VersionId: aws.String("VersionId"),
						},
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

			output, err := s3Client.deleteObjectsWithRetry(tt.args.ctx, tt.args.input, tt.args.bucketName, sleepTimeSecForS3)
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

func TestS3_ListObjectVersions(t *testing.T) {
	ctx := context.Background()
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

			output, err := s3Client.ListObjectVersions(tt.args.ctx, tt.args.bucketName)
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

func TestS3_CheckBucketExists(t *testing.T) {
	ctx := context.Background()
	mock := NewMockS3SDKClient()
	errorMock := NewErrorMockS3SDKClient()
	notExitsMock := NewNotExistsMockForListBucketsS3SDKClient()

	type args struct {
		ctx        context.Context
		bucketName *string
		client     IS3SDKClient
	}

	type want struct {
		exists bool
		err    error
	}

	cases := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "check bucket for bucket exists",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     mock,
			},
			want: want{
				exists: true,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check bucket for bucket do not exist",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     notExitsMock,
			},
			want: want{
				exists: false,
				err:    nil,
			},
			wantErr: false,
		},
		{
			name: "check bucket exists failure",
			args: args{
				ctx:        ctx,
				bucketName: aws.String("test"),
				client:     errorMock,
			},
			want: want{
				exists: false,
				err:    fmt.Errorf("ListBucketsError"),
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s3Client := NewS3(tt.args.client)

			output, err := s3Client.CheckBucketExists(tt.args.ctx, tt.args.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.err.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.err.Error())
				return
			}
			if !reflect.DeepEqual(output, tt.want.exists) {
				t.Errorf("output = %#v, want %#v", output, tt.want.exists)
			}
		})
	}
}
