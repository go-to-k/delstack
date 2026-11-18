package operation

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfnTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	gomock "go.uber.org/mock/gomock"
)

/*
	Test Cases
*/

func TestRoute53HostedZoneOperator_DeleteRoute53HostedZone(t *testing.T) {
	io.NewLogger(false)

	// Private by default, so that DNSSEC is not involved except in the DNSSEC cases.
	hostedZone := &types.HostedZone{
		Id:     aws.String("/hostedzone/Z0000000000001"),
		Name:   aws.String("example.com."),
		Config: &types.HostedZoneConfig{PrivateZone: true},
	}
	publicHostedZone := &types.HostedZone{
		Id:     aws.String("/hostedzone/Z0000000000001"),
		Name:   aws.String("example.com."),
		Config: &types.HostedZoneConfig{PrivateZone: false},
	}
	signing := &types.DNSSECStatus{ServeSignature: aws.String("SIGNING")}
	deleting := &types.DNSSECStatus{ServeSignature: aws.String("DELETING")}
	notSigning := &types.DNSSECStatus{ServeSignature: aws.String("NOT_SIGNING")}
	soa := types.ResourceRecordSet{
		Name:            aws.String("example.com."),
		Type:            types.RRTypeSoa,
		ResourceRecords: []types.ResourceRecord{{Value: aws.String("ns-1.awsdns-00.com. awsdns-hostmaster.amazon.com. 1 7200 900 1209600 86400")}},
	}
	ns := types.ResourceRecordSet{
		Name:            aws.String("example.com."),
		Type:            types.RRTypeNs,
		ResourceRecords: []types.ResourceRecord{{Value: aws.String("ns-1.awsdns-00.com.")}},
	}
	a := types.ResourceRecordSet{
		Name:            aws.String("www.example.com."),
		Type:            types.RRTypeA,
		ResourceRecords: []types.ResourceRecord{{Value: aws.String("192.0.2.1")}},
	}
	delegatedNs := types.ResourceRecordSet{
		Name:            aws.String("sub.example.com."),
		Type:            types.RRTypeNs,
		ResourceRecords: []types.ResourceRecord{{Value: aws.String("ns-2.awsdns-00.com.")}},
	}
	apexTxt := types.ResourceRecordSet{
		Name:            aws.String("example.com."),
		Type:            types.RRTypeTxt,
		ResourceRecords: []types.ResourceRecord{{Value: aws.String("\"v=spf1 -all\"")}},
	}
	upperCaseNs := types.ResourceRecordSet{
		Name:            aws.String("Example.COM."),
		Type:            types.RRTypeNs,
		ResourceRecords: []types.ResourceRecord{{Value: aws.String("ns-1.awsdns-00.com.")}},
	}
	alias := types.ResourceRecordSet{
		Name:        aws.String("alias.example.com."),
		Type:        types.RRTypeA,
		AliasTarget: &types.AliasTarget{DNSName: aws.String("www.example.com."), HostedZoneId: aws.String("Z0000000000001")},
	}
	trafficPolicyA := types.ResourceRecordSet{
		Name:                    aws.String("tp.example.com."),
		Type:                    types.RRTypeA,
		TrafficPolicyInstanceId: aws.String("tpi-1"),
	}
	trafficPolicyAAAA := types.ResourceRecordSet{
		Name:                    aws.String("tp.example.com."),
		Type:                    types.RRTypeAaaa,
		TrafficPolicyInstanceId: aws.String("tpi-1"),
	}
	trafficPolicyOther := types.ResourceRecordSet{
		Name:                    aws.String("tp2.example.com."),
		Type:                    types.RRTypeA,
		TrafficPolicyInstanceId: aws.String("tpi-2"),
	}
	manyRecordSets := make([]types.ResourceRecordSet, 1001)
	for i := range manyRecordSets {
		manyRecordSets[i] = types.ResourceRecordSet{
			Name:            aws.String(fmt.Sprintf("a%d.example.com.", i)),
			Type:            types.RRTypeA,
			ResourceRecords: []types.ResourceRecord{{Value: aws.String("192.0.2.1")}},
		}
	}

	type args struct {
		ctx          context.Context
		hostedZoneId *string
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIRoute53)
		want          error
		wantErr       bool
	}{
		{
			name: "delete hosted zone successfully with non-default record sets",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns, apexTxt, a, delegatedNs}, nil,
				)
				m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), []types.ResourceRecordSet{apexTxt, a, delegatedNs}).Return(nil)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone successfully with apex default record sets compared case-insensitively",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, upperCaseNs, a}, nil,
				)
				m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), []types.ResourceRecordSet{a}).Return(nil)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone successfully with alias record sets deleted first",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns, a, alias, apexTxt}, nil,
				)
				m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), []types.ResourceRecordSet{alias, a, apexTxt}).Return(nil)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone successfully with multiple batches in order",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					append([]types.ResourceRecordSet{soa, ns}, manyRecordSets...), nil,
				)
				gomock.InOrder(
					m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), manyRecordSets[:1000]).Return(nil),
					m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), manyRecordSets[1000:]).Return(nil),
					m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone successfully with traffic policy instances",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				gomock.InOrder(
					m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
						[]types.ResourceRecordSet{soa, ns, trafficPolicyA, trafficPolicyAAAA, trafficPolicyOther, a}, nil,
					),
					m.EXPECT().DeleteTrafficPolicyInstance(gomock.Any(), aws.String("tpi-1")).Return(nil).Times(1),
					m.EXPECT().DeleteTrafficPolicyInstance(gomock.Any(), aws.String("tpi-2")).Return(nil).Times(1),
					m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
						[]types.ResourceRecordSet{soa, ns, trafficPolicyA, a}, nil,
					),
					m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
						[]types.ResourceRecordSet{soa, ns, a}, nil,
					),
					m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), []types.ResourceRecordSet{a}).Return(nil),
					m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone failure for delete traffic policy instance errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns, trafficPolicyA}, nil,
				)
				m.EXPECT().DeleteTrafficPolicyInstance(gomock.Any(), aws.String("tpi-1")).Return(fmt.Errorf("DeleteTrafficPolicyInstanceError"))
			},
			want:    fmt.Errorf("DeleteTrafficPolicyInstanceError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for list resource record sets errors while waiting for traffic policy instances",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				gomock.InOrder(
					m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
						[]types.ResourceRecordSet{soa, ns, trafficPolicyA}, nil,
					),
					m.EXPECT().DeleteTrafficPolicyInstance(gomock.Any(), aws.String("tpi-1")).Return(nil),
					m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(nil, fmt.Errorf("ListResourceRecordSetsError")),
				)
			},
			want:    fmt.Errorf("ListResourceRecordSetsError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for traffic policy record sets not deleted in time",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns, trafficPolicyA}, nil,
				).Times(1 + route53WaitMaxAttempts)
				m.EXPECT().DeleteTrafficPolicyInstance(gomock.Any(), aws.String("tpi-1")).Return(nil)
			},
			want:    fmt.Errorf("[resource Z0000000000001] timed out waiting for record sets of traffic policy instances to be deleted"),
			wantErr: true,
		},
		{
			name: "delete hosted zone successfully for public hosted zone without DNSSEC signing",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(publicHostedZone, true, nil)
				m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(notSigning, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return([]types.ResourceRecordSet{soa, ns}, nil)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone successfully after disabling DNSSEC signing",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(publicHostedZone, true, nil)
				gomock.InOrder(
					m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(signing, nil),
					m.EXPECT().DisableHostedZoneDNSSEC(gomock.Any(), aws.String("Z0000000000001")).Return(nil),
					m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(deleting, nil),
					m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(notSigning, nil),
					m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return([]types.ResourceRecordSet{soa, ns, a}, nil),
					m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), []types.ResourceRecordSet{a}).Return(nil),
					m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil),
				)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone failure for get DNSSEC status errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(publicHostedZone, true, nil)
				m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(nil, fmt.Errorf("GetDNSSECError"))
			},
			want:    fmt.Errorf("GetDNSSECError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for disable DNSSEC errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(publicHostedZone, true, nil)
				m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(signing, nil)
				m.EXPECT().DisableHostedZoneDNSSEC(gomock.Any(), aws.String("Z0000000000001")).Return(fmt.Errorf("DisableHostedZoneDNSSECError"))
			},
			want:    fmt.Errorf("DisableHostedZoneDNSSECError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for DNSSEC signing not disabled in time",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(publicHostedZone, true, nil)
				m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(signing, nil)
				m.EXPECT().DisableHostedZoneDNSSEC(gomock.Any(), aws.String("Z0000000000001")).Return(nil)
				m.EXPECT().GetDNSSECStatus(gomock.Any(), aws.String("Z0000000000001")).Return(deleting, nil).Times(route53WaitMaxAttempts)
			},
			want:    fmt.Errorf("[resource Z0000000000001] timed out waiting for DNSSEC signing to be disabled"),
			wantErr: true,
		},
		{
			name: "delete hosted zone successfully with only default record sets",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns}, nil,
				)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone successfully for hosted zone not exists",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil, false, nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete hosted zone failure for get hosted zone errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(nil, false, fmt.Errorf("GetHostedZoneError"))
			},
			want:    fmt.Errorf("GetHostedZoneError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for list resource record sets errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(nil, fmt.Errorf("ListResourceRecordSetsError"))
			},
			want:    fmt.Errorf("ListResourceRecordSetsError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for delete resource record sets errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns, a}, nil,
				)
				m.EXPECT().DeleteResourceRecordSets(gomock.Any(), aws.String("Z0000000000001"), []types.ResourceRecordSet{a}).Return(fmt.Errorf("ChangeResourceRecordSetsError"))
			},
			want:    fmt.Errorf("ChangeResourceRecordSetsError"),
			wantErr: true,
		},
		{
			name: "delete hosted zone failure for delete hosted zone errors",
			args: args{
				ctx:          context.Background(),
				hostedZoneId: aws.String("Z0000000000001"),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return(
					[]types.ResourceRecordSet{soa, ns}, nil,
				)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(fmt.Errorf("DeleteHostedZoneError"))
			},
			want:    fmt.Errorf("DeleteHostedZoneError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			route53Mock := client.NewMockIRoute53(ctrl)
			tt.prepareMockFn(route53Mock)

			route53HostedZoneOperator := NewRoute53HostedZoneOperator(route53Mock)
			route53HostedZoneOperator.waitInterval = 0

			err := route53HostedZoneOperator.DeleteRoute53HostedZone(tt.args.ctx, tt.args.hostedZoneId)
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

func TestRoute53HostedZoneOperator_DeleteResourcesForRoute53HostedZone(t *testing.T) {
	io.NewLogger(false)

	hostedZone := &types.HostedZone{
		Id:     aws.String("/hostedzone/PhysicalResourceId1"),
		Name:   aws.String("example.com."),
		Config: &types.HostedZoneConfig{PrivateZone: true},
	}

	type args struct {
		ctx context.Context
	}

	cases := []struct {
		name          string
		args          args
		prepareMockFn func(m *client.MockIRoute53)
		want          error
		wantErr       bool
	}{
		{
			name: "delete resources successfully",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("PhysicalResourceId1")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("PhysicalResourceId1")).Return([]types.ResourceRecordSet{}, nil)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("PhysicalResourceId1")).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "delete resources failure",
			args: args{
				ctx: context.Background(),
			},
			prepareMockFn: func(m *client.MockIRoute53) {
				m.EXPECT().GetHostedZone(gomock.Any(), aws.String("PhysicalResourceId1")).Return(hostedZone, true, nil)
				m.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("PhysicalResourceId1")).Return([]types.ResourceRecordSet{}, nil)
				m.EXPECT().DeleteHostedZone(gomock.Any(), aws.String("PhysicalResourceId1")).Return(fmt.Errorf("DeleteHostedZoneError"))
			},
			want:    fmt.Errorf("DeleteHostedZoneError"),
			wantErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			route53Mock := client.NewMockIRoute53(ctrl)
			tt.prepareMockFn(route53Mock)

			route53HostedZoneOperator := NewRoute53HostedZoneOperator(route53Mock)
			route53HostedZoneOperator.AddResource(&cfnTypes.StackResourceSummary{
				LogicalResourceId:  aws.String("LogicalResourceId1"),
				ResourceStatus:     "DELETE_FAILED",
				ResourceType:       aws.String("AWS::Route53::HostedZone"),
				PhysicalResourceId: aws.String("PhysicalResourceId1"),
			})

			err := route53HostedZoneOperator.DeleteResources(tt.args.ctx)
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

func Test_chunkRecordSetsForChangeBatch(t *testing.T) {
	newRecordSet := func(name string, values ...string) types.ResourceRecordSet {
		records := []types.ResourceRecord{}
		for _, v := range values {
			records = append(records, types.ResourceRecord{Value: aws.String(v)})
		}
		return types.ResourceRecordSet{Name: aws.String(name), Type: types.RRTypeTxt, ResourceRecords: records}
	}
	alias := types.ResourceRecordSet{
		Name:        aws.String("alias.example.com."),
		Type:        types.RRTypeA,
		AliasTarget: &types.AliasTarget{DNSName: aws.String("d111111abcdef8.cloudfront.net.")},
	}

	manyAliases := make([]types.ResourceRecordSet, 1001)
	for i := range manyAliases {
		manyAliases[i] = alias
	}

	longValue := strings.Repeat("a", 20000)

	cases := []struct {
		name       string
		recordSets []types.ResourceRecordSet
		want       []int
	}{
		{
			name:       "no record sets",
			recordSets: []types.ResourceRecordSet{},
			want:       []int{},
		},
		{
			name:       "single batch",
			recordSets: []types.ResourceRecordSet{alias, newRecordSet("a.example.com.", "x", "y")},
			want:       []int{2},
		},
		{
			name:       "split by resource record count",
			recordSets: manyAliases,
			want:       []int{1000, 1},
		},
		{
			name:       "split by value characters",
			recordSets: []types.ResourceRecordSet{newRecordSet("a.example.com.", longValue), newRecordSet("b.example.com.", longValue)},
			want:       []int{1, 1},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			batches := chunkRecordSetsForChangeBatch(tt.recordSets)
			got := []int{}
			flattened := []types.ResourceRecordSet{}
			for _, batch := range batches {
				got = append(got, len(batch))
				flattened = append(flattened, batch...)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(flattened, tt.recordSets) {
				t.Errorf("order of record sets is not preserved")
			}
		})
	}
}

func TestRoute53HostedZoneOperator_DeleteRoute53HostedZone_CanceledWhileWaiting(t *testing.T) {
	io.NewLogger(false)

	ctx, cancel := context.WithCancel(context.Background())
	trafficPolicyA := types.ResourceRecordSet{
		Name:                    aws.String("tp.example.com."),
		Type:                    types.RRTypeA,
		TrafficPolicyInstanceId: aws.String("tpi-1"),
	}

	ctrl := gomock.NewController(t)
	route53Mock := client.NewMockIRoute53(ctrl)
	route53Mock.EXPECT().GetHostedZone(gomock.Any(), aws.String("Z0000000000001")).Return(&types.HostedZone{Name: aws.String("example.com."), Config: &types.HostedZoneConfig{PrivateZone: true}}, true, nil)
	route53Mock.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).Return([]types.ResourceRecordSet{trafficPolicyA}, nil)
	route53Mock.EXPECT().DeleteTrafficPolicyInstance(gomock.Any(), aws.String("tpi-1")).Return(nil)
	route53Mock.EXPECT().ListResourceRecordSets(gomock.Any(), aws.String("Z0000000000001")).DoAndReturn(
		func(context.Context, *string) ([]types.ResourceRecordSet, error) {
			cancel()
			return []types.ResourceRecordSet{trafficPolicyA}, nil
		},
	)

	route53HostedZoneOperator := NewRoute53HostedZoneOperator(route53Mock)
	route53HostedZoneOperator.waitInterval = time.Hour

	err := route53HostedZoneOperator.DeleteRoute53HostedZone(ctx, aws.String("Z0000000000001"))
	want := "[resource Z0000000000001] context canceled"
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %v", err, want)
	}
}
