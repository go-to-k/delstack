package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestS3VectorBucketOperator_DeleteS3VectorBucket(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx             context.Context
		vectorBucketArn *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Vectors)
		want          error
		wantErr       bool
	}{
		{
			name: "delete vector bucket successfully",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes: []types.IndexSummary{
							{
								IndexName: aws.String("index1"),
							},
							{
								IndexName: aws.String("index2"),
							},
						},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index1"), aws.String("test-vector-bucket")).Return(nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index2"), aws.String("test-vector-bucket")).Return(nil)
				m.EXPECT().DeleteVectorBucket(gomock.Any(), aws.String("test-vector-bucket")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vector bucket failure for check vector bucket exists errors",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(false, fmt.Errorf("CheckVectorBucketExistsError"))
			},
			want:    fmt.Errorf("CheckVectorBucketExistsError"),
			wantErr: true,
		},
		{
			name: "delete vector bucket successfully for vector bucket not exists",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vector bucket failure for list indexes errors",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(nil, fmt.Errorf("ListIndexesByPageError"))
			},
			want:    fmt.Errorf("ListIndexesByPageError"),
			wantErr: true,
		},
		{
			name: "delete vector bucket failure for delete index errors",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes: []types.IndexSummary{
							{
								IndexName: aws.String("index1"),
							},
						},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index1"), aws.String("test-vector-bucket")).Return(fmt.Errorf("DeleteIndexError"))
			},
			want:    fmt.Errorf("DeleteIndexError"),
			wantErr: true,
		},
		{
			name: "delete vector bucket failure for delete vector bucket errors",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes: []types.IndexSummary{
							{
								IndexName: aws.String("index1"),
							},
						},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index1"), aws.String("test-vector-bucket")).Return(nil)
				m.EXPECT().DeleteVectorBucket(gomock.Any(), aws.String("test-vector-bucket")).Return(fmt.Errorf("DeleteVectorBucketError"))
			},
			want:    fmt.Errorf("DeleteVectorBucketError"),
			wantErr: true,
		},
		{
			name: "delete vector bucket successfully for ListIndexesByPage with zero length",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes:   []types.IndexSummary{},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteVectorBucket(gomock.Any(), aws.String("test-vector-bucket")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vector bucket failure for ListIndexesByPage with zero length",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes:   []types.IndexSummary{},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteVectorBucket(gomock.Any(), aws.String("test-vector-bucket")).Return(fmt.Errorf("DeleteVectorBucketError"))
			},
			want:    fmt.Errorf("DeleteVectorBucketError"),
			wantErr: true,
		},
		{
			name: "delete vector bucket successfully if several loops are executed for ListIndexesByPage",
			args: args{
				ctx:             context.Background(),
				vectorBucketArn: aws.String("arn:aws:s3vectors:us-east-1:111111111111:bucket/test-vector-bucket"),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("test-vector-bucket")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes: []types.IndexSummary{
							{
								IndexName: aws.String("index1"),
							},
						},
						NextToken: aws.String("NextToken"),
					}, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("test-vector-bucket"), aws.String("NextToken"), nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes: []types.IndexSummary{
							{
								IndexName: aws.String("index2"),
							},
						},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index1"), aws.String("test-vector-bucket")).Return(nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index2"), aws.String("test-vector-bucket")).Return(nil)
				m.EXPECT().DeleteVectorBucket(gomock.Any(), aws.String("test-vector-bucket")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3Mock := client.NewMockIS3Vectors(ctrl)
			tt.prepareMockFn(s3Mock)

			s3VectorBucketOperator := NewS3VectorBucketOperator(s3Mock)

			err := s3VectorBucketOperator.DeleteS3VectorBucket(tt.args.ctx, tt.args.vectorBucketArn)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestS3VectorBucketOperator_DeleteResourcesForS3VectorBucket(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Vectors)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().ListIndexesByPage(gomock.Any(), aws.String("PhysicalResourceId1"), nil, nil).Return(
					&client.ListIndexesByPageOutput{
						Indexes: []types.IndexSummary{
							{
								IndexName: aws.String("index1"),
							},
							{
								IndexName: aws.String("index2"),
							},
						},
						NextToken: nil,
					}, nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index1"), aws.String("PhysicalResourceId1")).Return(nil)
				m.EXPECT().DeleteIndex(gomock.Any(), aws.String("index2"), aws.String("PhysicalResourceId1")).Return(nil)
				m.EXPECT().DeleteVectorBucket(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3Vectors) {
				m.EXPECT().CheckVectorBucketExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("CheckVectorBucketExistsError"))
			},
			want:    fmt.Errorf("CheckVectorBucketExistsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3Mock := client.NewMockIS3Vectors(ctrl)
			tt.prepareMockFn(s3Mock)

			s3VectorBucketOperator := NewS3VectorBucketOperator(s3Mock)

			s3VectorBucketOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::S3Vectors::VectorBucket"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := s3VectorBucketOperator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err.Error(), tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
