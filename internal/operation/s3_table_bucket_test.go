package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/s3tables/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestS3TableBucketOperator_DeleteS3TableBucket(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx            context.Context
		tableBucketArn *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Tables)
		want          error
		wantErr       bool
	}{
		{
			name: "delete table bucket successfully",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
							{
								Namespace: []string{"Namespace2"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
							{
								Name: aws.String("Table2"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table2"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace2"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
							{
								Name: aws.String("Table2"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace2"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table2"), aws.String("Namespace2"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace2"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTableBucket(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete table bucket failure for check table bucket exists errors",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(false, fmt.Errorf("CheckTableBucketExistsError"))
			},
			want:    fmt.Errorf("CheckTableBucketExistsError"),
			wantErr: true,
		},
		{
			name: "delete table bucket successfully for table bucket not exists",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete table bucket failure for list namespaces errors",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(nil, fmt.Errorf("ListNamespacesByPageError"))
			},
			want:    fmt.Errorf("ListNamespacesByPageError"),
			wantErr: true,
		},
		{
			name: "delete table bucket failure for delete namespace errors",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
							{
								Namespace: []string{"Namespace2"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
							{
								Name: aws.String("Table2"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteTableError"))
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table2"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteTableError"))
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteNamespaceError"))
			},
			want:    fmt.Errorf("DeleteNamespaceError"),
			wantErr: true,
		},
		{
			name: "delete table bucket failure for delete table errors",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteTableError"))
			},
			want:    fmt.Errorf("DeleteTableError"),
			wantErr: true,
		},
		{
			name: "delete table bucket failure for delete table bucket errors",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTableBucket(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteTableBucketError"))
			},
			want:    fmt.Errorf("DeleteTableBucketError"),
			wantErr: true,
		},
		{
			name: "delete table bucket successfully for ListTablesByPage with zero length",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables:            []types.TableSummary{},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTableBucket(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete table bucket failure for ListTablesByPage with zero length",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables:            []types.TableSummary{},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteNamespaceError"))
				m.EXPECT().DeleteTableBucket(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(fmt.Errorf("DeleteTableBucketError"))
			},
			want:    fmt.Errorf("DeleteTableBucketError"),
			wantErr: true,
		},
		{
			name: "delete table bucket successfully if several loops are executed for ListTablesByPage",
			args: args{
				ctx:            context.Background(),
				tableBucketArn: aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
						},
						ContinuationToken: aws.String("ContinuationToken"),
					}, nil)
				m.EXPECT().ListTablesByPage(
					gomock.Any(),
					aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"),
					aws.String("Namespace"),
					aws.String("ContinuationToken"),
				).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table2"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table2"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTableBucket(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3Mock := client.NewMockIS3Tables(ctrl)
			tt.prepareMockFn(s3Mock)

			s3TableBucketOperator := NewS3TableBucketOperator(s3Mock)

			err := s3TableBucketOperator.DeleteS3TableBucket(tt.args.ctx, tt.args.tableBucketArn)
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

func TestS3TableBucketOperator_DeleteResourcesForS3TableBucket(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Tables)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
				m.EXPECT().ListNamespacesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), nil).Return(
					&client.ListNamespacesByPageOutput{
						Namespaces: []types.NamespaceSummary{
							{
								Namespace: []string{"Namespace"},
							},
							{
								Namespace: []string{"Namespace2"},
							},
						},
					}, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test"), aws.String("Namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("Table"),
							},
							{
								Name: aws.String("Table2"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("Table2"), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("Namespace2"), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
				m.EXPECT().DeleteTableBucket(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckTableBucketExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:123456789012:bucket/test")).Return(true, nil)
			},
			want:    fmt.Errorf("ListTablesByPageError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s3Mock := client.NewMockIS3Tables(ctrl)
			tt.prepareMockFn(s3Mock)

			s3TableBucketOperator := NewS3TableBucketOperator(s3Mock)

			s3TableBucketOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::S3::Bucket"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := s3TableBucketOperator.DeleteResources(tt.args.ctx)
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
