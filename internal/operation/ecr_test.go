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

var _ client.IEcr = (*MockEcr)(nil)
var _ client.IEcr = (*DeleteRepositoryErrorMockEcr)(nil)
var _ client.IEcr = (*CheckEcrExistsErrorMockEcr)(nil)
var _ client.IEcr = (*CheckEcrNotExistsMockEcr)(nil)

/*
	Mocks for client
*/
type MockEcr struct{}

func NewMockEcr() *MockEcr {
	return &MockEcr{}
}

func (m *MockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return nil
}

func (m *MockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type DeleteRepositoryErrorMockEcr struct{}

func NewDeleteRepositoryErrorMockEcr() *DeleteRepositoryErrorMockEcr {
	return &DeleteRepositoryErrorMockEcr{}
}

func (m *DeleteRepositoryErrorMockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return fmt.Errorf("DeleteRepositoryError")
}

func (m *DeleteRepositoryErrorMockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return true, nil
}

type CheckEcrExistsErrorMockEcr struct{}

func NewCheckEcrExistsErrorMockEcr() *CheckEcrExistsErrorMockEcr {
	return &CheckEcrExistsErrorMockEcr{}
}

func (m *CheckEcrExistsErrorMockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return nil
}

func (m *CheckEcrExistsErrorMockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, fmt.Errorf("DescribeRepositoriesError")
}

type CheckEcrNotExistsMockEcr struct{}

func NewCheckEcrNotExistsMockEcr() *CheckEcrNotExistsMockEcr {
	return &CheckEcrNotExistsMockEcr{}
}

func (m *CheckEcrNotExistsMockEcr) DeleteRepository(ctx context.Context, repositoryName *string) error {
	return nil
}

func (m *CheckEcrNotExistsMockEcr) CheckEcrExists(ctx context.Context, repositoryName *string) (bool, error) {
	return false, nil
}

/*
	Test Cases
*/
func TestEcrOperator_DeleteRepository(t *testing.T) {
	io.NewLogger(false) // this test cannot do in parallel because this is a global variable
	mock := NewMockEcr()
	DeleteRepositoryErrorMock := NewDeleteRepositoryErrorMockEcr()
	checkEcrExistsErrorMock := NewCheckEcrExistsErrorMockEcr()
	checkEcrNotExistsMock := NewCheckEcrNotExistsMockEcr()

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
			name: "delete ecr successfully",
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
			name: "delete bucket failure for check bucket exists errors",
			args: args{
				ctx:            context.Background(),
				repositoryName: aws.String("test"),
				client:         checkEcrExistsErrorMock,
			},
			want:    fmt.Errorf("DescribeRepositoriesError"),
			wantErr: true,
		},
		{
			name: "delete bucket successfully for bucket not exists",
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
			ecrOperator := NewEcrOperator(tt.args.client)

			err := ecrOperator.DeleteEcr(tt.args.ctx, tt.args.repositoryName)
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

func TestEcrOperator_DeleteResourcesForEcr(t *testing.T) {
	io.NewLogger(false) // this test cannot do in parallel because this is a global variable
	mock := NewMockEcr()
	errorMock := NewDeleteRepositoryErrorMockEcr()

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
			ecrOperator := NewEcrOperator(tt.args.client)
			ecrOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::ECR::Repository"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := ecrOperator.DeleteResources(tt.args.ctx)
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
