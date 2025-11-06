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

func TestS3TablesNamespaceOperator_DeleteS3TablesNamespace(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx          context.Context
		namespaceArn *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIS3Tables)
		want          error
		wantErr       bool
	}{
		{
			name: "delete namespace successfully",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace")).Return(true, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("table1"),
							},
							{
								Name: aws.String("table2"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("table1"), aws.String("test-namespace"), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket")).Return(nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("table2"), aws.String("test-namespace"), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket")).Return(nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("test-namespace"), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete namespace failure for parse arn errors",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("invalid-arn"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {},
			want:          fmt.Errorf("invalid namespace ARN format: invalid-arn"),
			wantErr:       true,
		},
		{
			name: "delete namespace failure for check namespace exists errors",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace")).Return(false, fmt.Errorf("CheckNamespaceExistsError"))
			},
			want:    fmt.Errorf("CheckNamespaceExistsError"),
			wantErr: true,
		},
		{
			name: "delete namespace successfully for namespace not exists",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete namespace failure for list tables errors",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace")).Return(true, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace"), nil).Return(nil, fmt.Errorf("ListTablesByPageError"))
			},
			want:    fmt.Errorf("ListTablesByPageError"),
			wantErr: true,
		},
		{
			name: "delete namespace failure for delete table errors",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace")).Return(true, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables: []types.TableSummary{
							{
								Name: aws.String("table1"),
							},
						},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteTable(gomock.Any(), aws.String("table1"), aws.String("test-namespace"), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket")).Return(fmt.Errorf("DeleteTableError"))
			},
			want:    fmt.Errorf("DeleteTableError"),
			wantErr: true,
		},
		{
			name: "delete namespace failure for delete namespace errors",
			args: args{
				ctx:          context.Background(),
				namespaceArn: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace"),
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace")).Return(true, nil)
				m.EXPECT().ListTablesByPage(gomock.Any(), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket"), aws.String("test-namespace"), nil).Return(
					&client.ListTablesByPageOutput{
						Tables:            []types.TableSummary{},
						ContinuationToken: nil,
					}, nil)
				m.EXPECT().DeleteNamespace(gomock.Any(), aws.String("test-namespace"), aws.String("arn:aws:s3tables:us-east-1:111111111111:test-table-bucket")).Return(fmt.Errorf("DeleteNamespaceError"))
			},
			want:    fmt.Errorf("DeleteNamespaceError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3TablesClient := client.NewMockIS3Tables(ctrl)
			tt.prepareMockFn(mockS3TablesClient)

			operator := NewS3TablesNamespaceOperator(mockS3TablesClient)

			err := operator.DeleteS3TablesNamespace(tt.args.ctx, tt.args.namespaceArn)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
			}
		})
	}
}

func TestS3TablesNamespaceOperator_DeleteResources(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx       context.Context
		resources []*cfnTypes.StackResourceSummary
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
				resources: []*cfnTypes.StackResourceSummary{
					{
						PhysicalResourceId: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace1"),
					},
					{
						PhysicalResourceId: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace2"),
					},
				},
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil).Times(2)
				m.EXPECT().ListTablesByPage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(
					&client.ListTablesByPageOutput{
						Tables:            []types.TableSummary{},
						ContinuationToken: nil,
					}, nil).Times(2)
				m.EXPECT().DeleteNamespace(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
				resources: []*cfnTypes.StackResourceSummary{
					{
						PhysicalResourceId: aws.String("arn:aws:s3tables:us-east-1:111111111111:bucket/test-table-bucket/namespace/test-namespace1"),
					},
				},
			},
			prepareMockFn: func(m *client.MockIS3Tables) {
				m.EXPECT().CheckNamespaceExists(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, fmt.Errorf("CheckNamespaceExistsError"))
			},
			want:    fmt.Errorf("CheckNamespaceExistsError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockS3TablesClient := client.NewMockIS3Tables(ctrl)
			tt.prepareMockFn(mockS3TablesClient)

			operator := NewS3TablesNamespaceOperator(mockS3TablesClient)
			for _, resource := range tt.args.resources {
				operator.AddResource(resource)
			}

			err := operator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
			}
		})
	}
}
