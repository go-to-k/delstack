package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "github.com/golang/mock/gomock"
)

/*
	Test Cases
*/

func TestEcrRepositoryOperator_DeleteEcrRepository(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx            context.Context
		repositoryName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEcr)
		want          error
		wantErr       bool
	}{
		{
			name: "delete ecr repository successfully",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIEcr) {
				m.EXPECT().CheckEcrExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DeleteRepository(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete ecr repository failure",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIEcr) {
				m.EXPECT().CheckEcrExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DeleteRepository(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteRepositoryError"))
			},
			want:    fmt.Errorf("DeleteRepositoryError"),
			wantErr: true,
		},
		{
			name: "delete ecr repository failure for check ecr repository exists errors",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIEcr) {
				m.EXPECT().CheckEcrExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("DescribeRepositoriesError"))
			},
			want:    fmt.Errorf("DescribeRepositoriesError"),
			wantErr: true,
		},
		{
			name: "delete ecr repository successfully for ecr repository not exists",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIEcr) {
				m.EXPECT().CheckEcrExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ecrMock := client.NewMockIEcr(ctrl)
			tt.prepareMockFn(ecrMock)

			ecrRepositoryOperator := NewEcrRepositoryOperator(ecrMock)

			err := ecrRepositoryOperator.DeleteEcrRepository(tt.args.ctx, tt.args.repositoryName)
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

func TestEcrRepositoryOperator_DeleteResourcesForEcrRepository(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEcr)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIEcr) {
				m.EXPECT().CheckEcrExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().DeleteRepository(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIEcr) {
				m.EXPECT().CheckEcrExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("DescribeRepositoriesError"))
			},
			want:    fmt.Errorf("DescribeRepositoriesError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ecrMock := client.NewMockIEcr(ctrl)
			tt.prepareMockFn(ecrMock)

			ecrRepositoryOperator := NewEcrRepositoryOperator(ecrMock)
			ecrRepositoryOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::ECR::Repository"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := ecrRepositoryOperator.DeleteResources(tt.args.ctx)
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
