package operation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

func TestEC2VPCEndpointServiceOperator_DeleteEC2VPCEndpointService(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx       context.Context
		serviceId *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEC2)
		want          error
		wantErr       bool
	}{
		{
			name: "delete vpc endpoint service successfully when no connections",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-111"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-111")).Return([]ec2types.VpcEndpointConnection{}, nil)
				m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-111")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vpc endpoint service successfully when all connections are already harmless",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-222"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-222")).Return([]ec2types.VpcEndpointConnection{
					{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("rejected")},
					{VpcEndpointId: aws.String("vpce-2"), VpcEndpointState: ec2types.State("deleted")},
				}, nil)
				m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-222")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vpc endpoint service successfully after rejecting blocking connections",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-333"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				gomock.InOrder(
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-333")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("available")},
						{VpcEndpointId: aws.String("vpce-2"), VpcEndpointState: ec2types.State("pendingAcceptance")},
						{VpcEndpointId: aws.String("vpce-3"), VpcEndpointState: ec2types.State("rejected")},
					}, nil),
					m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-333"), []string{"vpce-1", "vpce-2"}).Return(nil),
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-333")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("rejected")},
						{VpcEndpointId: aws.String("vpce-2"), VpcEndpointState: ec2types.State("rejected")},
						{VpcEndpointId: aws.String("vpce-3"), VpcEndpointState: ec2types.State("rejected")},
					}, nil),
					m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-333")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vpc endpoint service successfully with connections split into multiple reject batches",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-444"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				connections := []ec2types.VpcEndpointConnection{}
				firstBatch := []string{}
				secondBatch := []string{}
				for i := 0; i < vpcEndpointConnectionRejectionBatchSize+1; i++ {
					vpcEndpointId := fmt.Sprintf("vpce-%d", i)
					connections = append(connections, ec2types.VpcEndpointConnection{
						VpcEndpointId:    aws.String(vpcEndpointId),
						VpcEndpointState: ec2types.State("available"),
					})
					if i < vpcEndpointConnectionRejectionBatchSize {
						firstBatch = append(firstBatch, vpcEndpointId)
					} else {
						secondBatch = append(secondBatch, vpcEndpointId)
					}
				}

				gomock.InOrder(
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-444")).Return(connections, nil),
					m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-444"), firstBatch).Return(nil),
					m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-444"), secondBatch).Return(nil),
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-444")).Return([]ec2types.VpcEndpointConnection{}, nil),
					m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-444")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete vpc endpoint service anyway when connections never leave the blocking states",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-555"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-555")).Return([]ec2types.VpcEndpointConnection{
					{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("available")},
				}, nil).Times(vpcEndpointConnectionRejectionMaxAttempts)
				// Rejected only once: a connection that stays `Available` in the describe
				// results must not be rejected again on every attempt.
				m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-555"), []string{"vpce-1"}).Return(nil).Times(1)
				m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-555")).Return(fmt.Errorf("ExistingVpcEndpointConnections"))
			},
			want:    fmt.Errorf("ExistingVpcEndpointConnections"),
			wantErr: true,
		},
		{
			name: "wait for a pending connection to settle before rejecting it",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-999"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				gomock.InOrder(
					// `Pending` cannot be rejected yet, but it becomes `Available` shortly,
					// so it must not be mistaken for "nothing blocks the deletion".
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-999")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("pending")},
					}, nil),
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-999")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("available")},
					}, nil),
					m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-999"), []string{"vpce-1"}).Return(nil),
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-999")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("rejected")},
					}, nil),
					m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-999")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "reject blocking connections reported with the SDK constant casing",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-aaa"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				// The API returns `available` / `pendingAcceptance`, but the states are
				// matched case-insensitively, so the types.State constants work too.
				gomock.InOrder(
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-aaa")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.StateAvailable},
						{VpcEndpointId: aws.String("vpce-2"), VpcEndpointState: ec2types.StatePendingAcceptance},
					}, nil),
					m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-aaa"), []string{"vpce-1", "vpce-2"}).Return(nil),
					m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-aaa")).Return([]ec2types.VpcEndpointConnection{
						{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.StateRejected},
						{VpcEndpointId: aws.String("vpce-2"), VpcEndpointState: ec2types.StateRejected},
					}, nil),
					m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-aaa")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "describe vpc endpoint connections failure aborts the deletion",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-666"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-666")).Return(nil, fmt.Errorf("DescribeVpcEndpointConnectionsError"))
			},
			want:    fmt.Errorf("DescribeVpcEndpointConnectionsError"),
			wantErr: true,
		},
		{
			name: "reject vpc endpoint connections failure aborts the deletion",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-777"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-777")).Return([]ec2types.VpcEndpointConnection{
					{VpcEndpointId: aws.String("vpce-1"), VpcEndpointState: ec2types.State("available")},
				}, nil)
				m.EXPECT().RejectVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-777"), []string{"vpce-1"}).Return(fmt.Errorf("RejectVpcEndpointConnectionsError"))
			},
			want:    fmt.Errorf("RejectVpcEndpointConnectionsError"),
			wantErr: true,
		},
		{
			name: "delete vpc endpoint service configuration failure",
			args: args{
				ctx:       context.Background(),
				serviceId: aws.String("vpce-svc-888"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), aws.String("vpce-svc-888")).Return([]ec2types.VpcEndpointConnection{}, nil)
				m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), aws.String("vpce-svc-888")).Return(fmt.Errorf("DeleteVpcEndpointServiceConfigurationError"))
			},
			want:    fmt.Errorf("DeleteVpcEndpointServiceConfigurationError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ec2Mock := client.NewMockIEC2(ctrl)
			tt.prepareMockFn(ec2Mock)

			ec2VPCEndpointServiceOperator := NewEC2VPCEndpointServiceOperator(ec2Mock)
			ec2VPCEndpointServiceOperator.rejectionInterval = time.Millisecond

			err := ec2VPCEndpointServiceOperator.DeleteEC2VPCEndpointService(tt.args.ctx, tt.args.serviceId)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
			}
		})
	}
}

func TestEC2VPCEndpointServiceOperator_DeleteResourcesForEC2VPCEndpointService(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx       context.Context
		resources []*cfnTypes.StackResourceSummary
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEC2)
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
				resources: []*cfnTypes.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::EC2::VPCEndpointService"),
						PhysicalResourceId: aws.String("vpce-svc-111"),
					},
					{
						LogicalResourceId:  aws.String("LogicalResourceId2"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::EC2::VPCEndpointService"),
						PhysicalResourceId: aws.String("vpce-svc-222"),
					},
				},
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), gomock.Any()).Return([]ec2types.VpcEndpointConnection{}, nil).Times(2)
				m.EXPECT().DeleteVpcEndpointServiceConfiguration(gomock.Any(), gomock.Any()).Return(nil).Times(2)
			},
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
				resources: []*cfnTypes.StackResourceSummary{
					{
						LogicalResourceId:  aws.String("LogicalResourceId1"),
						ResourceStatus:     "DELETE_FAILED",
						ResourceType:       aws.String("AWS::EC2::VPCEndpointService"),
						PhysicalResourceId: aws.String("vpce-svc-111"),
					},
				},
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeVpcEndpointConnections(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("DescribeVpcEndpointConnectionsError"))
			},
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ec2Mock := client.NewMockIEC2(ctrl)
			tt.prepareMockFn(ec2Mock)

			ec2VPCEndpointServiceOperator := NewEC2VPCEndpointServiceOperator(ec2Mock)
			ec2VPCEndpointServiceOperator.rejectionInterval = time.Millisecond
			for _, resource := range tt.args.resources {
				ec2VPCEndpointServiceOperator.AddResource(resource)
			}

			if got := ec2VPCEndpointServiceOperator.GetResourcesLength(); got != len(tt.args.resources) {
				t.Errorf("GetResourcesLength() = %v, want %v", got, len(tt.args.resources))
			}

			err := ec2VPCEndpointServiceOperator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %#v, wantErr %#v", err, tt.wantErr)
			}
		})
	}
}
