package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestAthenaWorkGroupOperator_DeleteAthenaWorkGroup(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx           context.Context
		workGroupName *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIAthena)
		want          error
		wantErr       bool
	}{
		{
			name: "delete athena work group successfully",
			args: args{
				ctx:           context.Background(),
				workGroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIAthena) {
				m.EXPECT().CheckAthenaWorkGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DeleteWorkGroup(gomock.Any(), aws.String("test")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete athena work group failure",
			args: args{
				ctx:           context.Background(),
				workGroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIAthena) {
				m.EXPECT().CheckAthenaWorkGroupExists(gomock.Any(), aws.String("test")).Return(true, nil)
				m.EXPECT().DeleteWorkGroup(gomock.Any(), aws.String("test")).Return(fmt.Errorf("DeleteWorkGroupError"))
			},
			want:    fmt.Errorf("DeleteWorkGroupError"),
			wantErr: true,
		},
		{
			name: "delete athena work group failure for check athena work group exists errors",
			args: args{
				ctx:           context.Background(),
				workGroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIAthena) {
				m.EXPECT().CheckAthenaWorkGroupExists(gomock.Any(), aws.String("test")).Return(false, fmt.Errorf("GetWorkGroupError"))
			},
			want:    fmt.Errorf("GetWorkGroupError"),
			wantErr: true,
		},
		{
			name: "delete athena work group successfully for athena work group not exists",
			args: args{
				ctx:           context.Background(),
				workGroupName: aws.String("test"),
			},
			prepareMockFn: func(m *client.MockIAthena) {
				m.EXPECT().CheckAthenaWorkGroupExists(gomock.Any(), aws.String("test")).Return(false, nil)
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			athenaMock := client.NewMockIAthena(ctrl)
			tt.prepareMockFn(athenaMock)

			athenaWorkGroupOperator := NewAthenaWorkGroupOperator(athenaMock)

			err := athenaWorkGroupOperator.DeleteAthenaWorkGroup(tt.args.ctx, tt.args.workGroupName)
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

func TestAthenaWorkGroupOperator_DeleteResourcesForAthenaWorkGroup(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIAthena)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIAthena) {
				m.EXPECT().CheckAthenaWorkGroupExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(true, nil)
				m.EXPECT().DeleteWorkGroup(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIAthena) {
				m.EXPECT().CheckAthenaWorkGroupExists(gomock.Any(), aws.String("PhysicalResourceId1")).Return(false, fmt.Errorf("GetWorkGroupError"))
			},
			want:    fmt.Errorf("GetWorkGroupError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			athenaMock := client.NewMockIAthena(ctrl)
			tt.prepareMockFn(athenaMock)

			athenaWorkGroupOperator := NewAthenaWorkGroupOperator(athenaMock)
			athenaWorkGroupOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::Athena::WorkGroup"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := athenaWorkGroupOperator.DeleteResources(tt.args.ctx)
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
