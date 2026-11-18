package preprocessor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go"
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
		route53 *client.MockIRoute53
	}

	hostedZoneResources := []types.StackResourceSummary{
		{
			ResourceType:       aws.String("AWS::Route53::HostedZone"),
			LogicalResourceId:  aws.String("MyHostedZone"),
			PhysicalResourceId: aws.String("Z0000000000001"),
		},
	}
	protection := func(i int) route53Types.ResourceRecordSetProtection {
		return route53Types.ResourceRecordSetProtection{
			RecordName: aws.String(fmt.Sprintf("r%d.example.com.", i)),
			RecordType: route53Types.RRTypeA,
		}
	}
	manyProtections := make([]route53Types.ResourceRecordSetProtection, 101)
	for i := range manyProtections {
		manyProtections[i] = protection(i)
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
		{
			name:      "route53 hosted zone with no record set protections",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: hostedZoneResources,
			},
			setup: func(m mocks) {
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return([]route53Types.ResourceRecordSetProtection{}, nil)
			},
			wantErr: false,
		},
		{
			name:      "route53 hosted zone with record set protections without force mode",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: hostedZoneResources,
			},
			setup: func(m mocks) {
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return([]route53Types.ResourceRecordSetProtection{protection(0)}, nil)
			},
			wantErr: true,
			errMsg:  "AWS::Route53::HostedZone: MyHostedZone (physical: Z0000000000001)",
		},
		{
			name:      "route53 hosted zone with record set protections check error without force mode",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: hostedZoneResources,
			},
			setup: func(m mocks) {
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return(nil, fmt.Errorf("ListResourceRecordSetProtectionsError"))
			},
			wantErr: true,
			errMsg:  "DeletionProtectionError",
		},
		{
			name:      "route53 hosted zone with record set protections check access denied is skipped",
			forceMode: false,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: hostedZoneResources,
			},
			setup: func(m mocks) {
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return(nil, &client.ClientError{
					ResourceName: aws.String("Z0000000000001"),
					Err:          &smithy.GenericAPIError{Code: "AccessDenied", Message: "not authorized"},
				})
			},
			wantErr: false,
		},
		{
			name:      "route53 hosted zone with record set protections with force mode",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: hostedZoneResources,
			},
			setup: func(m mocks) {
				gomock.InOrder(
					m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return(manyProtections, nil),
					m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return(manyProtections, nil),
					m.route53.EXPECT().DeleteResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001"), manyProtections[:100]).Return(nil),
					m.route53.EXPECT().DeleteResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001"), manyProtections[100:]).Return(nil),
				)
			},
			wantErr: false,
		},
		{
			name:      "route53 hosted zone with record set protections with force mode delete error",
			forceMode: true,
			args: args{
				ctx:       context.Background(),
				stackName: aws.String("test-stack"),
				resources: hostedZoneResources,
			},
			setup: func(m mocks) {
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001")).Return([]route53Types.ResourceRecordSetProtection{protection(0)}, nil).Times(2)
				m.route53.EXPECT().DeleteResourceRecordSetProtections(gomock.Any(), aws.String("Z0000000000001"), []route53Types.ResourceRecordSetProtection{protection(0)}).Return(fmt.Errorf("DeleteResourceRecordSetProtectionsError"))
			},
			wantErr: true,
			errMsg:  "DeleteResourceRecordSetProtectionsError",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockEC2 := client.NewMockIEC2(ctrl)
			mockRDS := client.NewMockIRDS(ctrl)
			mockCognito := client.NewMockICognito(ctrl)
			mockLogs := client.NewMockICloudWatchLogs(ctrl)
			mockELBV2 := client.NewMockIELBV2(ctrl)
			mockRoute53 := client.NewMockIRoute53(ctrl)

			tt.setup(mocks{
				ec2:     mockEC2,
				rds:     mockRDS,
				cognito: mockCognito,
				logs:    mockLogs,
				elbv2:   mockELBV2,
				route53: mockRoute53,
			})

			remover := NewDeletionProtectionRemover(tt.forceMode, mockEC2, mockRDS, mockCognito, mockLogs, mockELBV2, mockRoute53)
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
