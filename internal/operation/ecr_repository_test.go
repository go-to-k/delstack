package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
)

/*
	Test Cases
*/

func TestEcrRepositoryOperator_DeleteEcrRepository(t *testing.T) {
	io.NewLogger(false)
	mock := client.NewMockEcr()
	DeleteRepositoryErrorMock := client.NewDeleteRepositoryErrorMockEcr()
	checkEcrExistsErrorMock := client.NewCheckEcrExistsErrorMockEcr()
	checkEcrNotExistsMock := client.NewCheckEcrNotExistsMockEcr()

	type args struct {
		ctx            context.Context
		repositoryName *string
		client         client.IEcr
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete ecr repository successfully",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				client:         mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete ecr repository failure",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				client:         DeleteRepositoryErrorMock,
			},
			want:    fmt.Errorf("DeleteRepositoryError"),
			wantErr: true,
		},
		{
			name: "delete ecr repository failure for check ecr repository exists errors",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				client:         checkEcrExistsErrorMock,
			},
			want:    fmt.Errorf("DescribeRepositoriesError"),
			wantErr: true,
		},
		{
			name: "delete ecr repository successfully for ecr repository not exists",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				client:         checkEcrNotExistsMock,
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ecrRepositoryOperator := NewEcrRepositoryOperator(tt.args.client)

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
	mock := client.NewMockEcr()
	errorMock := client.NewDeleteRepositoryErrorMockEcr()

	type args struct {
		ctx    context.Context
		client client.IEcr
	}

	cases := []struct {
		name    string
		args    args
		want    error
		wantErr bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx:    context.Background(),
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    context.Background(),
				client: errorMock,
			},
			want:    fmt.Errorf("DeleteRepositoryError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ecrRepositoryOperator := NewEcrRepositoryOperator(tt.args.client)
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
