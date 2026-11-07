package operation

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestRoute53ResolverConfigOperator_DeleteResolverConfig(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx        context.Context
		resourceId *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIRoute53Resolver)
		want          error
		wantErr       bool
	}{
		{
			name: "reset autodefined reverse to enable successfully",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-test")).Return(&types.ResolverConfig{
					AutodefinedReverse: types.ResolverAutodefinedReverseStatusDisabled,
				}, true, nil)
				m.EXPECT().UpdateResolverConfig(gomock.Any(), aws.String("vpc-test"), types.AutodefinedReverseFlagEnable).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "do nothing when VPC no longer exists",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-test")).Return(nil, false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "do nothing when autodefined reverse is already enabled",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-test")).Return(&types.ResolverConfig{
					AutodefinedReverse: types.ResolverAutodefinedReverseStatusEnabled,
				}, true, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "do nothing when autodefined reverse is enabling",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-test")).Return(&types.ResolverConfig{
					AutodefinedReverse: types.ResolverAutodefinedReverseStatusEnabling,
				}, true, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "get resolver config failure",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-test")).Return(nil, false, fmt.Errorf("GetResolverConfigError"))
			},
			want:    fmt.Errorf("GetResolverConfigError"),
			wantErr: true,
		},
		{
			name: "update resolver config failure",
			args: args{
				ctx:        context.Background(),
				resourceId: aws.String("vpc-test"),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-test")).Return(&types.ResolverConfig{
					AutodefinedReverse: types.ResolverAutodefinedReverseStatusDisabled,
				}, true, nil)
				m.EXPECT().UpdateResolverConfig(gomock.Any(), aws.String("vpc-test"), types.AutodefinedReverseFlagEnable).Return(fmt.Errorf("UpdateResolverConfigError"))
			},
			want:    fmt.Errorf("UpdateResolverConfigError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			route53ResolverMock := client.NewMockIRoute53Resolver(ctrl)
			tt.prepareMockFn(route53ResolverMock)

			route53ResolverConfigOperator := NewRoute53ResolverConfigOperator(route53ResolverMock)

			err := route53ResolverConfigOperator.DeleteResolverConfig(tt.args.ctx, tt.args.resourceId)
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

func TestRoute53ResolverConfigOperator_DeleteResourcesForRoute53ResolverConfig(t *testing.T) {
	io.NewLogger(false)

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIRoute53Resolver)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-1")).Return(nil, false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIRoute53Resolver) {
				m.EXPECT().GetResolverConfig(gomock.Any(), aws.String("vpc-1")).Return(nil, false, fmt.Errorf("GetResolverConfigError"))
			},
			want:    fmt.Errorf("GetResolverConfigError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			route53ResolverMock := client.NewMockIRoute53Resolver(ctrl)
			tt.prepareMockFn(route53ResolverMock)

			route53ResolverConfigOperator := NewRoute53ResolverConfigOperator(route53ResolverMock)
			route53ResolverConfigOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::Route53Resolver::ResolverConfig"),
				PhysicalResourceId: aws.String("vpc-1"),
			})

			err := route53ResolverConfigOperator.DeleteResources(tt.args.ctx)
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
