package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

func TestEC2SubnetOperator_DeleteEC2Subnet(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx      context.Context
		subnetId *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEC2)
		want          error
		wantErr       bool
	}{
		{
			name: "delete subnet successfully when no orphan ENIs",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-111"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return([]ec2types.NetworkInterface{}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), aws.String("subnet-111")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete subnet successfully after cleaning multiple orphan ENIs",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-222"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return([]ec2types.NetworkInterface{
					{NetworkInterfaceId: aws.String("eni-1")},
					{NetworkInterfaceId: aws.String("eni-2")},
				}, nil)
				m.EXPECT().DeleteNetworkInterface(gomock.Any(), aws.String("eni-1")).Return(nil)
				m.EXPECT().DeleteNetworkInterface(gomock.Any(), aws.String("eni-2")).Return(nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), aws.String("subnet-222")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "describe network interfaces failure aborts subnet deletion",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-333"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("DescribeNetworkInterfacesError"))
			},
			want:    fmt.Errorf("DescribeNetworkInterfacesError"),
			wantErr: true,
		},
		{
			name: "ENI deletion failure aborts subnet deletion",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-444"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return([]ec2types.NetworkInterface{
					{NetworkInterfaceId: aws.String("eni-9")},
				}, nil)
				m.EXPECT().DeleteNetworkInterface(gomock.Any(), aws.String("eni-9")).Return(fmt.Errorf("DeleteNetworkInterfaceError"))
			},
			want:    fmt.Errorf("DeleteNetworkInterfaceError"),
			wantErr: true,
		},
		{
			name: "subnet deletion failure (e.g. non-Lambda dependency) propagates",
			args: args{
				ctx:      context.Background(),
				subnetId: aws.String("subnet-555"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return([]ec2types.NetworkInterface{}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), aws.String("subnet-555")).Return(fmt.Errorf("DependencyViolation"))
			},
			want:    fmt.Errorf("DependencyViolation"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ec2Mock := client.NewMockIEC2(ctrl)
			tt.prepareMockFn(ec2Mock)

			operator := NewEC2SubnetOperator(ec2Mock)

			err := operator.DeleteEC2Subnet(tt.args.ctx, tt.args.subnetId)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}

func TestEC2SubnetOperator_DeleteResourcesForEC2Subnet(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEC2)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return([]ec2types.NetworkInterface{}, nil)
				m.EXPECT().DeleteSubnet(gomock.Any(), aws.String("subnet-111")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("DescribeNetworkInterfacesError"))
			},
			want:    fmt.Errorf("DescribeNetworkInterfacesError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ec2Mock := client.NewMockIEC2(ctrl)
			tt.prepareMockFn(ec2Mock)

			operator := NewEC2SubnetOperator(ec2Mock)
			operator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::EC2::Subnet"),
				PhysicalResourceId: aws.String("subnet-111"),
			})

			err := operator.DeleteResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want.Error() {
				t.Errorf("err = %#v, want %#v", err.Error(), tt.want.Error())
				return
			}
		})
	}
}
