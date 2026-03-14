package preprocessor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestDeletionProtectionRemover_Preprocess(t *testing.T) {
	// Initialize logger for tests
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	io.Logger = &logger
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx       context.Context
		stackName *string
		resources []types.StackResourceSummary
	}

	type mocks struct {
		ec2     *client.MockIEC2
		rds     *client.MockIRDS
		cognito *client.MockICognito
		logs    *client.MockICloudWatchLogs
		elbv2   *client.MockIELBV2
	}

	cases := []struct {
		name      string
		forceMode bool
		args      args
		setup     func(m mocks)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "no target resources",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::S3::Bucket"),
						LogicalResourceId:  aws.String("MyBucket"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::Lambda::Function"),
						LogicalResourceId:  aws.String("MyFunction"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup:   func(m mocks) {},
			wantErr: false,
		},
		{
			name:      "target resources with no protection",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::RDS::DBInstance"),
						LogicalResourceId:  aws.String("MyDBInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::RDS::DBCluster"),
						LogicalResourceId:  aws.String("MyDBCluster"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::Cognito::UserPool"),
						LogicalResourceId:  aws.String("MyUserPool"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::Logs::LogGroup"),
						LogicalResourceId:  aws.String("MyLogGroup"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::ElasticLoadBalancingV2::LoadBalancer"),
						LogicalResourceId:  aws.String("MyLoadBalancer"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
				m.rds.EXPECT().CheckDBInstanceDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
				m.rds.EXPECT().CheckDBClusterDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
				m.cognito.EXPECT().CheckUserPoolDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
				m.logs.EXPECT().CheckLogGroupDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
				m.elbv2.EXPECT().CheckLoadBalancerDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
			},
			wantErr: false,
		},
		{
			name:      "protected resources without force mode",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::RDS::DBInstance"),
						LogicalResourceId:  aws.String("MyDBInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.rds.EXPECT().CheckDBInstanceDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
			},
			wantErr: true,
			errMsg:  "DeletionProtectionError",
		},
		{
			name:      "protected resources with force mode",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::RDS::DBInstance"),
						LogicalResourceId:  aws.String("MyDBInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.rds.EXPECT().CheckDBInstanceDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.ec2.EXPECT().DisableTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
				m.rds.EXPECT().DisableDBInstanceDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "protected resources with force mode disable fails",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.ec2.EXPECT().DisableTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(fmt.Errorf("access denied"))
			},
			wantErr: true,
			errMsg:  "DeletionProtectionError",
		},
		{
			name:      "check protection error treats resource as protected",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(false, fmt.Errorf("api error"))
			},
			wantErr: true,
			errMsg:  "DeletionProtectionError",
		},
		{
			name:      "check protection error with force mode attempts disable",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(false, fmt.Errorf("api error"))
				m.ec2.EXPECT().DisableTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "DELETE_COMPLETE resources are skipped",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
						ResourceStatus:     types.ResourceStatusDeleteComplete,
					},
				},
			},
			setup:   func(m mocks) {},
			wantErr: false,
		},
		{
			name:      "multiple resource types mixed",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::S3::Bucket"),
						LogicalResourceId:  aws.String("MyBucket"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::RDS::DBCluster"),
						LogicalResourceId:  aws.String("MyDBCluster"),
						PhysicalResourceId: aws.String("physical-id"),
					},
					{
						ResourceType:       aws.String("AWS::Cognito::UserPool"),
						LogicalResourceId:  aws.String("MyUserPool"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.rds.EXPECT().CheckDBClusterDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(false, nil)
				m.cognito.EXPECT().CheckUserPoolDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
			},
			wantErr: true,
			errMsg:  "DeletionProtectionError",
		},
		{
			name:      "ec2 termination protection check and disable",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::EC2::Instance"),
						LogicalResourceId:  aws.String("MyInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.ec2.EXPECT().CheckTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.ec2.EXPECT().DisableTerminationProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "rds db instance deletion protection",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::RDS::DBInstance"),
						LogicalResourceId:  aws.String("MyDBInstance"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.rds.EXPECT().CheckDBInstanceDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.rds.EXPECT().DisableDBInstanceDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "rds db cluster deletion protection",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::RDS::DBCluster"),
						LogicalResourceId:  aws.String("MyDBCluster"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.rds.EXPECT().CheckDBClusterDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.rds.EXPECT().DisableDBClusterDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "cognito user pool deletion protection",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Cognito::UserPool"),
						LogicalResourceId:  aws.String("MyUserPool"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.cognito.EXPECT().CheckUserPoolDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.cognito.EXPECT().DisableUserPoolDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "cloudwatch logs deletion protection",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::Logs::LogGroup"),
						LogicalResourceId:  aws.String("MyLogGroup"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.logs.EXPECT().CheckLogGroupDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.logs.EXPECT().DisableLogGroupDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "elbv2 load balancer deletion protection",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: []types.StackResourceSummary{
					{
						ResourceType:       aws.String("AWS::ElasticLoadBalancingV2::LoadBalancer"),
						LogicalResourceId:  aws.String("MyLoadBalancer"),
						PhysicalResourceId: aws.String("physical-id"),
					},
				},
			},
			setup: func(m mocks) {
				m.elbv2.EXPECT().CheckLoadBalancerDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(true, nil)
				m.elbv2.EXPECT().DisableLoadBalancerDeletionProtection(gomock.Any(), aws.String("physical-id")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockEC2 := client.NewMockIEC2(ctrl)
			mockRDS := client.NewMockIRDS(ctrl)
			mockCognito := client.NewMockICognito(ctrl)
			mockLogs := client.NewMockICloudWatchLogs(ctrl)
			mockELBV2 := client.NewMockIELBV2(ctrl)

			tt.setup(mocks{
				ec2:     mockEC2,
				rds:     mockRDS,
				cognito: mockCognito,
				logs:    mockLogs,
				elbv2:   mockELBV2,
			})

			remover := NewDeletionProtectionRemover(tt.forceMode, mockEC2, mockRDS, mockCognito, mockLogs, mockELBV2)
			err := remover.Preprocess(tt.args.ctx, tt.args.stackName, tt.args.resources)

			if (err != nil) != tt.wantErr {
				t.Errorf("Preprocess() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errMsg != "" && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Preprocess() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}
