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

func TestEC2SecurityGroupOperator_DeleteEC2SecurityGroup(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx             context.Context
		securityGroupId *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIEC2)
		want          error
		wantErr       bool
	}{
		{
			name: "delete security group successfully after cleaning orphan ENIs",
			args: args{
				ctx:             context.Background(),
				securityGroupId: aws.String("sg-111"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DeleteOrphanLambdaENIsByFilter(gomock.Any(), "group-id", "sg-111").Return(nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), aws.String("sg-111")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "orphan ENI cleanup failure aborts security group deletion",
			args: args{
				ctx:             context.Background(),
				securityGroupId: aws.String("sg-222"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DeleteOrphanLambdaENIsByFilter(gomock.Any(), "group-id", "sg-222").Return(fmt.Errorf("DescribeNetworkInterfacesError"))
			},
			want:    fmt.Errorf("DescribeNetworkInterfacesError"),
			wantErr: true,
		},
		{
			name: "security group deletion failure (e.g. non-Lambda dependency) propagates",
			args: args{
				ctx:             context.Background(),
				securityGroupId: aws.String("sg-333"),
			},
			prepareMockFn: func(m *client.MockIEC2) {
				m.EXPECT().DeleteOrphanLambdaENIsByFilter(gomock.Any(), "group-id", "sg-333").Return(nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), aws.String("sg-333")).Return(fmt.Errorf("DependencyViolation"))
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

			operator := NewEC2SecurityGroupOperator(ec2Mock)

			err := operator.DeleteEC2SecurityGroup(tt.args.ctx, tt.args.securityGroupId)
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

func TestEC2SecurityGroupOperator_DeleteResourcesForEC2SecurityGroup(t *testing.T) {
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
				m.EXPECT().DeleteOrphanLambdaENIsByFilter(gomock.Any(), "group-id", "sg-111").Return(nil)
				m.EXPECT().DeleteSecurityGroup(gomock.Any(), aws.String("sg-111")).Return(nil)
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
				m.EXPECT().DeleteOrphanLambdaENIsByFilter(gomock.Any(), "group-id", "sg-111").Return(fmt.Errorf("DescribeNetworkInterfacesError"))
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

			operator := NewEC2SecurityGroupOperator(ec2Mock)
			operator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::EC2::SecurityGroup"),
				PhysicalResourceId: aws.String("sg-111"),
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
