package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/client"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/resourcetype"
)

var _ client.IEcr = (*mockEcr)(nil)
var _ client.IEcr = (*errorMockEcr)(nil)

/*
	Mocks for client
*/
type mockEcr struct{}

func NewMockEcr() *mockEcr {
	return &mockEcr{}
}

func (m *mockEcr) DeleteRepository(repositoryName *string) error {
	return nil
}

type errorMockEcr struct{}

func NewErrorMockEcr() *errorMockEcr {
	return &errorMockEcr{}
}

func (m *errorMockEcr) DeleteRepository(repositoryName *string) error {
	return fmt.Errorf("DeleteRepositoryError")
}

/*
	Test Cases
*/
func TestDeleteRepository(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockEcr()
	errorMock := NewErrorMockEcr()

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
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete ecr failure",
			args: args{
				ctx:            ctx,
				repositoryName: aws.String("test"),
				client:         errorMock,
			},
			want:    fmt.Errorf("DeleteRepositoryError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ecrOperator := NewEcrOperator(tt.args.client)

			err := ecrOperator.DeleteEcr(tt.args.repositoryName)
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

func TestDeleteResourcesForEcrVault(t *testing.T) {
	logger.NewLogger()
	ctx := context.TODO()
	mock := NewMockEcr()
	errorMock := NewErrorMockEcr()

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
				ctx:    ctx,
				client: mock,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx:    ctx,
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
				ResourceType:       aws.String(resourcetype.ECR_REPOSITORY),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := ecrOperator.DeleteResources()
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
