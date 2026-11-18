package client

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/aws/smithy-go/middleware"
)

type startRecordNameKeyForRoute53 struct{}

type nextTokenKeyForRoute53 struct{}

func getNextTokenForRoute53Initialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	if v, ok := in.Parameters.(*route53.ListResourceRecordSetProtectionsInput); ok {
		ctx = middleware.WithStackValue(ctx, nextTokenKeyForRoute53{}, v.NextToken)
	}
	return next.HandleInitialize(ctx, in)
}

func getStartRecordNameForRoute53Initialize(
	ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler,
) (
	out middleware.InitializeOutput, metadata middleware.Metadata, err error,
) {
	if v, ok := in.Parameters.(*route53.ListResourceRecordSetsInput); ok {
		ctx = middleware.WithStackValue(ctx, startRecordNameKeyForRoute53{}, v.StartRecordName)
	}
	return next.HandleInitialize(ctx, in)
}

func newRoute53ClientForTest(t *testing.T, ctx context.Context, withAPIOptionsFunc func(*middleware.Stack) error) *Route53 {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("ap-northeast-1"),
		config.WithAPIOptions([]func(*middleware.Stack) error{withAPIOptionsFunc}),
	)
	if err != nil {
		t.Fatal(err)
	}
	return NewRoute53(route53.NewFromConfig(cfg))
}

func route53FinalizeMock(name string, result any, err error) func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Finalize.Add(
			middleware.FinalizeMiddlewareFunc(
				name,
				func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
					return middleware.FinalizeOutput{Result: result}, middleware.Metadata{}, err
				},
			),
			middleware.Before,
		)
	}
}

/*
	Test Cases
*/

func TestRoute53_GetHostedZone(t *testing.T) {
	hostedZone := &types.HostedZone{
		Id:   aws.String("/hostedzone/Z0000000000001"),
		Name: aws.String("example.com."),
	}

	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		wantHostedZone     *types.HostedZone
		wantExists         bool
		wantErr            error
	}{
		{
			name:               "get hosted zone successfully",
			withAPIOptionsFunc: route53FinalizeMock("GetHostedZoneMock", &route53.GetHostedZoneOutput{HostedZone: hostedZone}, nil),
			wantHostedZone:     hostedZone,
			wantExists:         true,
		},
		{
			name:               "get hosted zone returns not exists when hosted zone not found",
			withAPIOptionsFunc: route53FinalizeMock("GetHostedZoneNotFoundMock", &route53.GetHostedZoneOutput{}, &types.NoSuchHostedZone{Message: aws.String("No hosted zone found")}),
		},
		{
			name:               "get hosted zone failure",
			withAPIOptionsFunc: route53FinalizeMock("GetHostedZoneErrorMock", &route53.GetHostedZoneOutput{}, fmt.Errorf("GetHostedZoneError")),
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: GetHostedZone, GetHostedZoneError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			output, exists, err := route53Client.GetHostedZone(ctx, aws.String("Z0000000000001"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
			if !reflect.DeepEqual(output, tt.wantHostedZone) {
				t.Errorf("output = %#v, want %#v", output, tt.wantHostedZone)
			}
			if exists != tt.wantExists {
				t.Errorf("exists = %#v, want %#v", exists, tt.wantExists)
			}
		})
	}
}

func TestRoute53_ListResourceRecordSets(t *testing.T) {
	recordSet := func(name string) types.ResourceRecordSet {
		return types.ResourceRecordSet{Name: aws.String(name), Type: types.RRTypeA}
	}

	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		want               []types.ResourceRecordSet
		wantErr            error
	}{
		{
			name: "list resource record sets successfully",
			withAPIOptionsFunc: route53FinalizeMock("ListResourceRecordSetsMock", &route53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []types.ResourceRecordSet{recordSet("a.example.com.")},
				IsTruncated:        false,
			}, nil),
			want: []types.ResourceRecordSet{recordSet("a.example.com.")},
		},
		{
			name: "list resource record sets with pagination successfully",
			withAPIOptionsFunc: func(stack *middleware.Stack) error {
				err := stack.Initialize.Add(
					middleware.InitializeMiddlewareFunc(
						"GetStartRecordName",
						getStartRecordNameForRoute53Initialize,
					), middleware.Before,
				)
				if err != nil {
					return err
				}

				return stack.Finalize.Add(
					middleware.FinalizeMiddlewareFunc(
						"ListResourceRecordSetsWithPaginationMock",
						func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
							startRecordName := middleware.GetStackValue(ctx, startRecordNameKeyForRoute53{}).(*string)
							if startRecordName == nil {
								return middleware.FinalizeOutput{
									Result: &route53.ListResourceRecordSetsOutput{
										ResourceRecordSets: []types.ResourceRecordSet{recordSet("a.example.com.")},
										IsTruncated:        true,
										NextRecordName:     aws.String("b.example.com."),
										NextRecordType:     types.RRTypeA,
									},
								}, middleware.Metadata{}, nil
							}
							return middleware.FinalizeOutput{
								Result: &route53.ListResourceRecordSetsOutput{
									ResourceRecordSets: []types.ResourceRecordSet{recordSet(aws.ToString(startRecordName))},
									IsTruncated:        false,
								},
							}, middleware.Metadata{}, nil
						},
					),
					middleware.Before,
				)
			},
			want: []types.ResourceRecordSet{recordSet("a.example.com."), recordSet("b.example.com.")},
		},
		{
			name:               "list resource record sets failure",
			withAPIOptionsFunc: route53FinalizeMock("ListResourceRecordSetsErrorMock", &route53.ListResourceRecordSetsOutput{}, fmt.Errorf("ListResourceRecordSetsError")),
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: ListResourceRecordSets, ListResourceRecordSetsError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			output, err := route53Client.ListResourceRecordSets(ctx, aws.String("Z0000000000001"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
				}
				return
			}
			if !reflect.DeepEqual(output, tt.want) {
				t.Errorf("output = %#v, want %#v", output, tt.want)
			}
		})
	}
}

func TestRoute53_DeleteResourceRecordSets(t *testing.T) {
	recordSets := []types.ResourceRecordSet{
		{Name: aws.String("a.example.com."), Type: types.RRTypeA},
		{Name: aws.String("b.example.com."), Type: types.RRTypeCname},
	}

	cases := []struct {
		name        string
		recordSets  []types.ResourceRecordSet
		mockErr     error
		wantChanges []types.Change
		wantErr     error
	}{
		{
			name:       "delete resource record sets successfully",
			recordSets: recordSets,
			wantChanges: []types.Change{
				{Action: types.ChangeActionDelete, ResourceRecordSet: &recordSets[0]},
				{Action: types.ChangeActionDelete, ResourceRecordSet: &recordSets[1]},
			},
		},
		{
			name:        "delete resource record sets does nothing for empty record sets",
			recordSets:  []types.ResourceRecordSet{},
			mockErr:     fmt.Errorf("must not be called"),
			wantChanges: nil,
		},
		{
			name:       "delete resource record sets failure",
			recordSets: recordSets,
			mockErr:    fmt.Errorf("ChangeResourceRecordSetsError"),
			wantChanges: []types.Change{
				{Action: types.ChangeActionDelete, ResourceRecordSet: &recordSets[0]},
				{Action: types.ChangeActionDelete, ResourceRecordSet: &recordSets[1]},
			},
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: ChangeResourceRecordSets, ChangeResourceRecordSetsError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var gotChanges []types.Change
			route53Client := newRoute53ClientForTest(t, ctx, func(stack *middleware.Stack) error {
				err := stack.Initialize.Add(
					middleware.InitializeMiddlewareFunc(
						"CaptureChanges",
						func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
							if v, ok := in.Parameters.(*route53.ChangeResourceRecordSetsInput); ok {
								gotChanges = v.ChangeBatch.Changes
							}
							return next.HandleInitialize(ctx, in)
						},
					), middleware.Before,
				)
				if err != nil {
					return err
				}
				return route53FinalizeMock("ChangeResourceRecordSetsMock", &route53.ChangeResourceRecordSetsOutput{}, tt.mockErr)(stack)
			})

			err := route53Client.DeleteResourceRecordSets(ctx, aws.String("Z0000000000001"), tt.recordSets)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
			if !reflect.DeepEqual(gotChanges, tt.wantChanges) {
				t.Errorf("changes = %#v, want %#v", gotChanges, tt.wantChanges)
			}
		})
	}
}

func TestRoute53_DeleteHostedZone(t *testing.T) {
	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		wantErr            error
	}{
		{
			name:               "delete hosted zone successfully",
			withAPIOptionsFunc: route53FinalizeMock("DeleteHostedZoneMock", &route53.DeleteHostedZoneOutput{}, nil),
		},
		{
			name:               "delete hosted zone successfully when hosted zone not found",
			withAPIOptionsFunc: route53FinalizeMock("DeleteHostedZoneNotFoundMock", &route53.DeleteHostedZoneOutput{}, &types.NoSuchHostedZone{Message: aws.String("No hosted zone found")}),
		},
		{
			name:               "delete hosted zone failure",
			withAPIOptionsFunc: route53FinalizeMock("DeleteHostedZoneErrorMock", &route53.DeleteHostedZoneOutput{}, &types.HostedZoneNotEmpty{Message: aws.String("not empty")}),
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: DeleteHostedZone, HostedZoneNotEmpty: not empty"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			err := route53Client.DeleteHostedZone(ctx, aws.String("Z0000000000001"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
		})
	}
}

func TestRoute53_DeleteTrafficPolicyInstance(t *testing.T) {
	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		wantErr            error
	}{
		{
			name:               "delete traffic policy instance successfully",
			withAPIOptionsFunc: route53FinalizeMock("DeleteTrafficPolicyInstanceMock", &route53.DeleteTrafficPolicyInstanceOutput{}, nil),
		},
		{
			name:               "delete traffic policy instance successfully when traffic policy instance not found",
			withAPIOptionsFunc: route53FinalizeMock("DeleteTrafficPolicyInstanceNotFoundMock", &route53.DeleteTrafficPolicyInstanceOutput{}, &types.NoSuchTrafficPolicyInstance{Message: aws.String("not found")}),
		},
		{
			name:               "delete traffic policy instance failure",
			withAPIOptionsFunc: route53FinalizeMock("DeleteTrafficPolicyInstanceErrorMock", &route53.DeleteTrafficPolicyInstanceOutput{}, fmt.Errorf("DeleteTrafficPolicyInstanceError")),
			wantErr: &ClientError{
				ResourceName: aws.String("tpi-1"),
				Err:          fmt.Errorf("operation error Route 53: DeleteTrafficPolicyInstance, DeleteTrafficPolicyInstanceError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			err := route53Client.DeleteTrafficPolicyInstance(ctx, aws.String("tpi-1"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
		})
	}
}

func TestRoute53_ListResourceRecordSetProtections(t *testing.T) {
	protection := func(name string) types.ResourceRecordSetProtection {
		return types.ResourceRecordSetProtection{RecordName: aws.String(name), RecordType: types.RRTypeA}
	}

	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		want               []types.ResourceRecordSetProtection
		wantErr            error
	}{
		{
			name: "list resource record set protections with pagination successfully",
			withAPIOptionsFunc: func(stack *middleware.Stack) error {
				err := stack.Initialize.Add(
					middleware.InitializeMiddlewareFunc(
						"GetNextToken",
						getNextTokenForRoute53Initialize,
					), middleware.Before,
				)
				if err != nil {
					return err
				}

				return stack.Finalize.Add(
					middleware.FinalizeMiddlewareFunc(
						"ListResourceRecordSetProtectionsWithPaginationMock",
						func(ctx context.Context, input middleware.FinalizeInput, handler middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
							token := middleware.GetStackValue(ctx, nextTokenKeyForRoute53{}).(*string)
							if token == nil {
								return middleware.FinalizeOutput{
									Result: &route53.ListResourceRecordSetProtectionsOutput{
										ResourceRecordSetProtections: []types.ResourceRecordSetProtection{protection("a.example.com.")},
										NextToken:                    aws.String("NextToken"),
									},
								}, middleware.Metadata{}, nil
							}
							return middleware.FinalizeOutput{
								Result: &route53.ListResourceRecordSetProtectionsOutput{
									ResourceRecordSetProtections: []types.ResourceRecordSetProtection{protection("b.example.com.")},
								},
							}, middleware.Metadata{}, nil
						},
					),
					middleware.Before,
				)
			},
			want: []types.ResourceRecordSetProtection{protection("a.example.com."), protection("b.example.com.")},
		},
		{
			name:               "list resource record set protections returns empty when hosted zone not found",
			withAPIOptionsFunc: route53FinalizeMock("ListResourceRecordSetProtectionsNotFoundMock", &route53.ListResourceRecordSetProtectionsOutput{}, &types.NoSuchHostedZone{Message: aws.String("No hosted zone found")}),
			want:               []types.ResourceRecordSetProtection{},
		},
		{
			name:               "list resource record set protections failure",
			withAPIOptionsFunc: route53FinalizeMock("ListResourceRecordSetProtectionsErrorMock", &route53.ListResourceRecordSetProtectionsOutput{}, fmt.Errorf("ListResourceRecordSetProtectionsError")),
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: ListResourceRecordSetProtections, ListResourceRecordSetProtectionsError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			output, err := route53Client.ListResourceRecordSetProtections(ctx, aws.String("Z0000000000001"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
				}
				return
			}
			if !reflect.DeepEqual(output, tt.want) {
				t.Errorf("output = %#v, want %#v", output, tt.want)
			}
		})
	}
}

func TestRoute53_DeleteResourceRecordSetProtections(t *testing.T) {
	protections := []types.ResourceRecordSetProtection{
		{RecordName: aws.String("a.example.com."), RecordType: types.RRTypeA, Comment: aws.String("comment")},
	}
	wantKeys := []types.ResourceRecordSetProtectionKey{
		{RecordName: aws.String("a.example.com."), RecordType: types.RRTypeA},
	}

	cases := []struct {
		name        string
		protections []types.ResourceRecordSetProtection
		mockErr     error
		wantKeys    []types.ResourceRecordSetProtectionKey
		wantErr     error
	}{
		{
			name:        "delete resource record set protections successfully",
			protections: protections,
			wantKeys:    wantKeys,
		},
		{
			name:        "delete resource record set protections does nothing for empty protections",
			protections: []types.ResourceRecordSetProtection{},
			mockErr:     fmt.Errorf("must not be called"),
		},
		{
			name:        "delete resource record set protections successfully when hosted zone not found",
			protections: protections,
			mockErr:     &types.NoSuchHostedZone{Message: aws.String("No hosted zone found")},
			wantKeys:    wantKeys,
		},
		{
			name:        "delete resource record set protections failure",
			protections: protections,
			mockErr:     fmt.Errorf("DeleteResourceRecordSetProtectionsError"),
			wantKeys:    wantKeys,
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: DeleteResourceRecordSetProtections, DeleteResourceRecordSetProtectionsError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var gotKeys []types.ResourceRecordSetProtectionKey
			route53Client := newRoute53ClientForTest(t, ctx, func(stack *middleware.Stack) error {
				err := stack.Initialize.Add(
					middleware.InitializeMiddlewareFunc(
						"CaptureProtections",
						func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
							if v, ok := in.Parameters.(*route53.DeleteResourceRecordSetProtectionsInput); ok {
								gotKeys = v.Protections
							}
							return next.HandleInitialize(ctx, in)
						},
					), middleware.Before,
				)
				if err != nil {
					return err
				}
				return route53FinalizeMock("DeleteResourceRecordSetProtectionsMock", &route53.DeleteResourceRecordSetProtectionsOutput{}, tt.mockErr)(stack)
			})

			err := route53Client.DeleteResourceRecordSetProtections(ctx, aws.String("Z0000000000001"), tt.protections)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
			if !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("keys = %#v, want %#v", gotKeys, tt.wantKeys)
			}
		})
	}
}

func TestRoute53_GetDNSSECStatus(t *testing.T) {
	status := &types.DNSSECStatus{ServeSignature: aws.String("SIGNING")}

	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		want               *types.DNSSECStatus
		wantErr            error
	}{
		{
			name:               "get DNSSEC status successfully",
			withAPIOptionsFunc: route53FinalizeMock("GetDNSSECMock", &route53.GetDNSSECOutput{Status: status}, nil),
			want:               status,
		},
		{
			name:               "get DNSSEC status failure",
			withAPIOptionsFunc: route53FinalizeMock("GetDNSSECErrorMock", &route53.GetDNSSECOutput{}, fmt.Errorf("GetDNSSECError")),
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: GetDNSSEC, GetDNSSECError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			output, err := route53Client.GetDNSSECStatus(ctx, aws.String("Z0000000000001"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
			if !reflect.DeepEqual(output, tt.want) {
				t.Errorf("output = %#v, want %#v", output, tt.want)
			}
		})
	}
}

func TestRoute53_DisableHostedZoneDNSSEC(t *testing.T) {
	cases := []struct {
		name               string
		withAPIOptionsFunc func(*middleware.Stack) error
		wantErr            error
	}{
		{
			name:               "disable hosted zone DNSSEC successfully",
			withAPIOptionsFunc: route53FinalizeMock("DisableHostedZoneDNSSECMock", &route53.DisableHostedZoneDNSSECOutput{}, nil),
		},
		{
			name:               "disable hosted zone DNSSEC successfully when DNSSEC not found",
			withAPIOptionsFunc: route53FinalizeMock("DisableHostedZoneDNSSECNotFoundMock", &route53.DisableHostedZoneDNSSECOutput{}, &types.DNSSECNotFound{Message: aws.String("not found")}),
		},
		{
			name:               "disable hosted zone DNSSEC failure",
			withAPIOptionsFunc: route53FinalizeMock("DisableHostedZoneDNSSECErrorMock", &route53.DisableHostedZoneDNSSECOutput{}, fmt.Errorf("DisableHostedZoneDNSSECError")),
			wantErr: &ClientError{
				ResourceName: aws.String("Z0000000000001"),
				Err:          fmt.Errorf("operation error Route 53: DisableHostedZoneDNSSEC, DisableHostedZoneDNSSECError"),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			route53Client := newRoute53ClientForTest(t, ctx, tt.withAPIOptionsFunc)

			err := route53Client.DisableHostedZoneDNSSEC(ctx, aws.String("Z0000000000001"))
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("error = %#v, wantErr %#v", err, tt.wantErr)
			}
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Fatalf("err = %#v, want %#v", err.Error(), tt.wantErr.Error())
			}
		})
	}
}
