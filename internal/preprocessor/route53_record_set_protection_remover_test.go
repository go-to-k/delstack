package preprocessor

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"go.uber.org/mock/gomock"
)

func TestRoute53RecordSetProtectionRemover_Preprocess(t *testing.T) {
	io.NewLogger(false)

	recordSet := types.StackResourceSummary{
		ResourceType:       aws.String("AWS::Route53::RecordSet"),
		LogicalResourceId:  aws.String("Www"),
		PhysicalResourceId: aws.String("www.example.com"),
	}
	recordSetGroup := types.StackResourceSummary{
		ResourceType:       aws.String("AWS::Route53::RecordSetGroup"),
		LogicalResourceId:  aws.String("Group"),
		PhysicalResourceId: aws.String("group-id"),
	}
	wildcardRecordSet := types.StackResourceSummary{
		ResourceType:       aws.String("AWS::Route53::RecordSet"),
		LogicalResourceId:  aws.String("Www"),
		PhysicalResourceId: aws.String("*.example.com"),
	}
	createFailedRecordSet := types.StackResourceSummary{
		ResourceType:       aws.String("AWS::Route53::RecordSet"),
		LogicalResourceId:  aws.String("Www"),
		PhysicalResourceId: aws.String("www.example.com"),
		ResourceStatus:     types.ResourceStatusCreateFailed,
	}
	inStackZone := types.StackResourceSummary{
		ResourceType:       aws.String("AWS::Route53::HostedZone"),
		LogicalResourceId:  aws.String("Zone"),
		PhysicalResourceId: aws.String("ZINSTACK"),
	}
	protection := func(name string, recordType route53Types.RRType) route53Types.ResourceRecordSetProtection {
		return route53Types.ResourceRecordSetProtection{RecordName: aws.String(name), RecordType: recordType}
	}
	accessDenied := &client.ClientError{Err: &smithy.GenericAPIError{Code: "AccessDenied", Message: "not authorized"}}

	recordSetJSON := func(properties string) *string {
		return aws.String(`{"Resources":{"Www":{"Type":"AWS::Route53::RecordSet","Properties":` + properties + `}}}`)
	}

	type mocks struct {
		cfn     *client.MockICloudFormation
		route53 *client.MockIRoute53
	}

	cases := []struct {
		name      string
		forceMode bool
		resources []types.StackResourceSummary
		setup     func(m mocks)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "no record sets",
			resources: []types.StackResourceSummary{inStackZone},
			setup:     func(m mocks) {},
		},
		{
			name:      "record set in external hosted zone with no matching protections",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"/hostedzone/ZEXT","Name":"www.example.com","Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection("www.example.com.", route53Types.RRTypeAaaa),
					protection("api.example.com.", route53Types.RRTypeA),
				}, nil)
			},
		},
		{
			name:      "record set in external hosted zone with protection without force mode",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"WWW.example.com","Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection("www.example.com.", route53Types.RRTypeA),
				}, nil)
			},
			wantErr: true,
			errMsg:  "AWS::Route53::RecordSet: Www (www.example.com A in hosted zone ZEXT)",
		},
		{
			name:      "record set in external hosted zone with protection with force mode deletes only matching protections",
			forceMode: true,
			resources: []types.StackResourceSummary{wildcardRecordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"*.example.com","Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection(`\052.example.com.`, route53Types.RRTypeA),
					protection("other.example.com.", route53Types.RRTypeA),
				}, nil)
				m.route53.EXPECT().ListResourceRecordSetsByNameAndType(gomock.Any(), aws.String("ZEXT"), aws.String("*.example.com."), route53Types.RRTypeA).Return(
					[]route53Types.ResourceRecordSet{{Name: aws.String(`\052.example.com.`), Type: route53Types.RRTypeA}}, nil,
				)
				m.route53.EXPECT().DeleteResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT"), []route53Types.ResourceRecordSetProtection{
					protection(`\052.example.com.`, route53Types.RRTypeA),
				}).Return(nil)
			},
		},
		{
			name:      "record set with force mode delete error",
			forceMode: true,
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"www.example.com","Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection("www.example.com.", route53Types.RRTypeA),
				}, nil)
				m.route53.EXPECT().ListResourceRecordSetsByNameAndType(gomock.Any(), aws.String("ZEXT"), aws.String("www.example.com."), route53Types.RRTypeA).Return([]route53Types.ResourceRecordSet{}, nil)
				m.route53.EXPECT().DeleteResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT"), gomock.Any()).Return(fmt.Errorf("DeleteResourceRecordSetProtectionsError"))
			},
			wantErr: true,
			errMsg:  "DeleteResourceRecordSetProtectionsError",
		},
		{
			name:      "record set in hosted zone in the stack is skipped",
			resources: []types.StackResourceSummary{recordSet, inStackZone},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":{"Ref":"Zone"},"Name":"www.example.com","Type":"A"}`), nil)
			},
		},
		{
			name:      "record set name is taken from the physical resource id",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":{"Fn::Sub":"www.${Domain}"},"Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection("www.example.com.", route53Types.RRTypeA),
				}, nil)
			},
			wantErr: true,
			errMsg:  "AWS::Route53::RecordSet: Www (www.example.com A in hosted zone ZEXT)",
		},
		{
			name:      "yaml template with short-form ref to parameter is skipped",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(aws.String(
					"Parameters:\n  ZoneId:\n    Type: String\nResources:\n  Www:\n    Type: AWS::Route53::RecordSet\n    Properties:\n      HostedZoneId: !Ref ZoneId\n      Name: www.example.com\n      Type: A\n",
				), nil)
			},
		},
		{
			name:      "yaml template with short-form import value is skipped",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(aws.String(
					"Resources:\n  Www:\n    Type: AWS::Route53::RecordSet\n    Properties:\n      HostedZoneId: !ImportValue ZoneId\n      Name: www.example.com\n      Type: A\n",
				), nil)
			},
		},
		{
			name:      "record set group resolves hosted zone by name with force mode and dedupes protections",
			forceMode: true,
			resources: []types.StackResourceSummary{recordSetGroup},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(aws.String(
					"Resources:\n  Group:\n    Type: AWS::Route53::RecordSetGroup\n    Properties:\n      HostedZoneName: example.com.\n      RecordSets:\n        - Name: a.example.com\n          Type: A\n          SetIdentifier: one\n        - Name: a.example.com\n          Type: A\n          SetIdentifier: two\n        - Name: b.example.com\n          Type: TXT\n",
				), nil)
				m.route53.EXPECT().ListHostedZonesByName(gomock.Any(), aws.String("example.com.")).Return([]route53Types.HostedZone{
					{Id: aws.String("/hostedzone/ZEXT"), Name: aws.String("example.com.")},
				}, nil).Times(1)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection("a.example.com.", route53Types.RRTypeA),
				}, nil)
				m.route53.EXPECT().ListResourceRecordSetsByNameAndType(gomock.Any(), aws.String("ZEXT"), aws.String("a.example.com."), route53Types.RRTypeA).Return([]route53Types.ResourceRecordSet{
					{Name: aws.String("a.example.com."), Type: route53Types.RRTypeA, SetIdentifier: aws.String("one")},
					{Name: aws.String("a.example.com."), Type: route53Types.RRTypeA, SetIdentifier: aws.String("two")},
				}, nil)
				m.route53.EXPECT().DeleteResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT"), []route53Types.ResourceRecordSetProtection{
					protection("a.example.com.", route53Types.RRTypeA),
				}).Return(nil)
			},
		},
		{
			name:      "record set group with ambiguous hosted zone name is skipped",
			resources: []types.StackResourceSummary{recordSetGroup},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(aws.String(
					`{"Resources":{"Group":{"Type":"AWS::Route53::RecordSetGroup","Properties":{"HostedZoneName":"example.com.","RecordSets":[{"Name":"a.example.com","Type":"A"}]}}}}`,
				), nil)
				m.route53.EXPECT().ListHostedZonesByName(gomock.Any(), aws.String("example.com.")).Return([]route53Types.HostedZone{
					{Id: aws.String("/hostedzone/ZPUBLIC"), Name: aws.String("example.com.")},
					{Id: aws.String("/hostedzone/ZPRIVATE"), Name: aws.String("example.com.")},
				}, nil)
			},
		},
		{
			name:      "record set with force mode refuses to delete a protection also covering record sets not in the stack",
			forceMode: true,
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"www.example.com","Type":"A","SetIdentifier":"mine"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return([]route53Types.ResourceRecordSetProtection{
					protection("www.example.com.", route53Types.RRTypeA),
				}, nil)
				m.route53.EXPECT().ListResourceRecordSetsByNameAndType(gomock.Any(), aws.String("ZEXT"), aws.String("www.example.com."), route53Types.RRTypeA).Return([]route53Types.ResourceRecordSet{
					{Name: aws.String("www.example.com."), Type: route53Types.RRTypeA, SetIdentifier: aws.String("mine")},
					{Name: aws.String("www.example.com."), Type: route53Types.RRTypeA, SetIdentifier: aws.String("theirs")},
				}, nil)
			},
			wantErr: true,
			errMsg:  "www.example.com. A (set identifier: theirs) in hosted zone ZEXT",
		},
		{
			name:      "record set that failed to create is skipped",
			forceMode: true,
			resources: []types.StackResourceSummary{createFailedRecordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"www.example.com","Type":"A"}`), nil)
			},
		},
		{
			name:      "record set with get att to hosted zone in the stack is skipped",
			resources: []types.StackResourceSummary{recordSet, inStackZone},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(aws.String(
					"Resources:\n  Www:\n    Type: AWS::Route53::RecordSet\n    Properties:\n      HostedZoneId: !GetAtt Zone.Id\n      Name: www.example.com\n      Type: A\n",
				), nil)
			},
		},
		{
			name:      "record set group entry overrides the hosted zone of the group",
			resources: []types.StackResourceSummary{recordSetGroup},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(aws.String(
					`{"Resources":{"Group":{"Type":"AWS::Route53::RecordSetGroup","Properties":{"HostedZoneId":"ZGROUP","RecordSets":[{"HostedZoneId":"ZENTRY","Name":"a.example.com","Type":"A"}]}}}}`,
				), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZENTRY")).Return([]route53Types.ResourceRecordSetProtection{}, nil)
			},
		},
		{
			name:      "get template access denied is skipped",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(nil, accessDenied)
			},
		},
		{
			name:      "get template error",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(nil, fmt.Errorf("GetTemplateError"))
			},
			wantErr: true,
			errMsg:  "GetTemplateError",
		},
		{
			name:      "list protections access denied is skipped",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"www.example.com","Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return(nil, accessDenied)
			},
		},
		{
			name:      "list protections error",
			resources: []types.StackResourceSummary{recordSet},
			setup: func(m mocks) {
				m.cfn.EXPECT().GetTemplate(gomock.Any(), aws.String("test-stack")).Return(recordSetJSON(`{"HostedZoneId":"ZEXT","Name":"www.example.com","Type":"A"}`), nil)
				m.route53.EXPECT().ListResourceRecordSetProtections(gomock.Any(), aws.String("ZEXT")).Return(nil, fmt.Errorf("ListResourceRecordSetProtectionsError"))
			},
			wantErr: true,
			errMsg:  "ListResourceRecordSetProtectionsError",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockCfn := client.NewMockICloudFormation(ctrl)
			mockRoute53 := client.NewMockIRoute53(ctrl)
			tt.setup(mocks{cfn: mockCfn, route53: mockRoute53})

			remover := NewRoute53RecordSetProtectionRemover(tt.forceMode, mockCfn, mockRoute53)
			err := remover.Preprocess(context.Background(), aws.String("test-stack"), tt.resources)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Preprocess() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Preprocess() error = %v, should contain %v", err, tt.errMsg)
			}
		})
	}
}

func Test_parseTemplate(t *testing.T) {
	cases := []struct {
		name string
		body string
		want map[string]any
	}{
		{
			name: "json",
			body: `{"Resources":{"A":{"Properties":{"Id":{"Ref":"B"}}}}}`,
			want: map[string]any{"Resources": map[string]any{"A": map[string]any{"Properties": map[string]any{"Id": map[string]any{"Ref": "B"}}}}},
		},
		{
			name: "yaml with short-form intrinsic functions",
			body: "Resources:\n  A:\n    Properties:\n      Ref: !Ref B\n      Sub: !Sub 'x-${B}'\n      Join: !Join ['', [a, !Ref B]]\n      Literal: plain\n      Number: 1\n",
			want: map[string]any{"Resources": map[string]any{"A": map[string]any{"Properties": map[string]any{
				"Ref":     map[string]any{"Ref": "B"},
				"Sub":     map[string]any{"Fn::Sub": "x-${B}"},
				"Join":    map[string]any{"Fn::Join": []any{"", []any{"a", map[string]any{"Ref": "B"}}}},
				"Literal": "plain",
				"Number":  1,
			}}}},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTemplate(tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
