//go:generate go tool mockgen -source=$GOFILE -destination=route53_mock.go -package=$GOPACKAGE -write_package_comment=false
package client

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type IRoute53 interface {
	GetHostedZone(ctx context.Context, hostedZoneId *string) (*types.HostedZone, bool, error)
	ListResourceRecordSets(ctx context.Context, hostedZoneId *string) ([]types.ResourceRecordSet, error)
	DeleteResourceRecordSets(ctx context.Context, hostedZoneId *string, recordSets []types.ResourceRecordSet) error
	DeleteHostedZone(ctx context.Context, hostedZoneId *string) error
	DeleteTrafficPolicyInstance(ctx context.Context, trafficPolicyInstanceId *string) error
	ListResourceRecordSetProtections(ctx context.Context, hostedZoneId *string) ([]types.ResourceRecordSetProtection, error)
	DeleteResourceRecordSetProtections(ctx context.Context, hostedZoneId *string, protections []types.ResourceRecordSetProtection) error
	GetDNSSEC(ctx context.Context, hostedZoneId *string) (*types.DNSSECStatus, []types.KeySigningKey, error)
	DisableHostedZoneDNSSEC(ctx context.Context, hostedZoneId *string) error
	DeactivateKeySigningKey(ctx context.Context, hostedZoneId *string, name *string) error
	DeleteKeySigningKey(ctx context.Context, hostedZoneId *string, name *string) error
	ListHostedZonesByName(ctx context.Context, dnsName *string) ([]types.HostedZone, error)
	ListResourceRecordSetsByNameAndType(ctx context.Context, hostedZoneId *string, name *string, recordType types.RRType) ([]types.ResourceRecordSet, error)
}

var _ IRoute53 = (*Route53)(nil)

type Route53 struct {
	client *route53.Client
}

func NewRoute53(client *route53.Client) *Route53 {
	return &Route53{
		client,
	}
}

// GetHostedZone returns exists=false when the hosted zone no longer exists.
func (r *Route53) GetHostedZone(ctx context.Context, hostedZoneId *string) (*types.HostedZone, bool, error) {
	input := &route53.GetHostedZoneInput{
		Id: hostedZoneId,
	}

	output, err := r.client.GetHostedZone(ctx, input)
	if err != nil {
		var notFound *types.NoSuchHostedZone
		if errors.As(err, &notFound) {
			return nil, false, nil
		}
		return nil, false, &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return output.HostedZone, true, nil
}

func (r *Route53) ListResourceRecordSets(ctx context.Context, hostedZoneId *string) ([]types.ResourceRecordSet, error) {
	recordSets := []types.ResourceRecordSet{}

	paginator := route53.NewListResourceRecordSetsPaginator(r.client, &route53.ListResourceRecordSetsInput{
		HostedZoneId: hostedZoneId,
	})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ClientError{
				ResourceName: hostedZoneId,
				Err:          err,
			}
		}
		recordSets = append(recordSets, output.ResourceRecordSets...)
	}

	return recordSets, nil
}

func (r *Route53) DeleteResourceRecordSets(ctx context.Context, hostedZoneId *string, recordSets []types.ResourceRecordSet) error {
	if len(recordSets) == 0 {
		return nil
	}

	changes := make([]types.Change, 0, len(recordSets))
	for _, recordSet := range recordSets {
		changes = append(changes, types.Change{
			Action:            types.ChangeActionDelete,
			ResourceRecordSet: &recordSet,
		})
	}

	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: hostedZoneId,
		ChangeBatch: &types.ChangeBatch{
			Changes: changes,
		},
	}

	_, err := r.client.ChangeResourceRecordSets(ctx, input)
	if err != nil {
		return &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return nil
}

// DeleteHostedZone succeeds without doing anything when the hosted zone no longer exists.
func (r *Route53) DeleteHostedZone(ctx context.Context, hostedZoneId *string) error {
	input := &route53.DeleteHostedZoneInput{
		Id: hostedZoneId,
	}

	_, err := r.client.DeleteHostedZone(ctx, input)
	if err != nil {
		var notFound *types.NoSuchHostedZone
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return nil
}

// DeleteTrafficPolicyInstance succeeds without doing anything when the traffic policy instance no longer exists.
func (r *Route53) DeleteTrafficPolicyInstance(ctx context.Context, trafficPolicyInstanceId *string) error {
	input := &route53.DeleteTrafficPolicyInstanceInput{
		Id: trafficPolicyInstanceId,
	}

	_, err := r.client.DeleteTrafficPolicyInstance(ctx, input)
	if err != nil {
		var notFound *types.NoSuchTrafficPolicyInstance
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: trafficPolicyInstanceId,
			Err:          err,
		}
	}
	return nil
}

// ListResourceRecordSetProtections returns an empty list when the hosted zone no longer exists.
func (r *Route53) ListResourceRecordSetProtections(ctx context.Context, hostedZoneId *string) ([]types.ResourceRecordSetProtection, error) {
	protections := []types.ResourceRecordSetProtection{}

	paginator := route53.NewListResourceRecordSetProtectionsPaginator(r.client, &route53.ListResourceRecordSetProtectionsInput{
		HostedZoneId: hostedZoneId,
	})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			var notFound *types.NoSuchHostedZone
			if errors.As(err, &notFound) {
				return []types.ResourceRecordSetProtection{}, nil
			}
			return nil, &ClientError{
				ResourceName: hostedZoneId,
				Err:          err,
			}
		}
		protections = append(protections, output.ResourceRecordSetProtections...)
	}

	return protections, nil
}

// DeleteResourceRecordSetProtections succeeds without doing anything when the hosted zone no longer exists.
func (r *Route53) DeleteResourceRecordSetProtections(ctx context.Context, hostedZoneId *string, protections []types.ResourceRecordSetProtection) error {
	if len(protections) == 0 {
		return nil
	}

	keys := make([]types.ResourceRecordSetProtectionKey, 0, len(protections))
	for _, protection := range protections {
		keys = append(keys, types.ResourceRecordSetProtectionKey{
			RecordName: protection.RecordName,
			RecordType: protection.RecordType,
		})
	}

	input := &route53.DeleteResourceRecordSetProtectionsInput{
		HostedZoneId: hostedZoneId,
		Protections:  keys,
	}

	_, err := r.client.DeleteResourceRecordSetProtections(ctx, input)
	if err != nil {
		var notFound *types.NoSuchHostedZone
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return nil
}

func (r *Route53) GetDNSSEC(ctx context.Context, hostedZoneId *string) (*types.DNSSECStatus, []types.KeySigningKey, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: hostedZoneId,
	}

	output, err := r.client.GetDNSSEC(ctx, input)
	if err != nil {
		return nil, nil, &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return output.Status, output.KeySigningKeys, nil
}

// DeactivateKeySigningKey succeeds without doing anything when the key-signing key no longer exists.
func (r *Route53) DeactivateKeySigningKey(ctx context.Context, hostedZoneId *string, name *string) error {
	input := &route53.DeactivateKeySigningKeyInput{
		HostedZoneId: hostedZoneId,
		Name:         name,
	}

	_, err := r.client.DeactivateKeySigningKey(ctx, input)
	if err != nil {
		var notFound *types.NoSuchKeySigningKey
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return nil
}

// DeleteKeySigningKey succeeds without doing anything when the key-signing key no longer exists.
func (r *Route53) DeleteKeySigningKey(ctx context.Context, hostedZoneId *string, name *string) error {
	input := &route53.DeleteKeySigningKeyInput{
		HostedZoneId: hostedZoneId,
		Name:         name,
	}

	_, err := r.client.DeleteKeySigningKey(ctx, input)
	if err != nil {
		var notFound *types.NoSuchKeySigningKey
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return nil
}

// DisableHostedZoneDNSSEC succeeds without doing anything when DNSSEC signing is not enabled.
func (r *Route53) DisableHostedZoneDNSSEC(ctx context.Context, hostedZoneId *string) error {
	input := &route53.DisableHostedZoneDNSSECInput{
		HostedZoneId: hostedZoneId,
	}

	_, err := r.client.DisableHostedZoneDNSSEC(ctx, input)
	if err != nil {
		var notFound *types.DNSSECNotFound
		if errors.As(err, &notFound) {
			return nil
		}
		return &ClientError{
			ResourceName: hostedZoneId,
			Err:          err,
		}
	}
	return nil
}

// ListHostedZonesByName returns the hosted zones whose name is exactly the specified DNS name.
func (r *Route53) ListHostedZonesByName(ctx context.Context, dnsName *string) ([]types.HostedZone, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: dnsName,
	}

	output, err := r.client.ListHostedZonesByName(ctx, input)
	if err != nil {
		return nil, &ClientError{
			ResourceName: dnsName,
			Err:          err,
		}
	}

	// Hosted zones are sorted by name starting from the specified name, so stop at the first other name.
	hostedZones := []types.HostedZone{}
	for _, hostedZone := range output.HostedZones {
		if !strings.EqualFold(strings.TrimSuffix(aws.ToString(hostedZone.Name), "."), strings.TrimSuffix(aws.ToString(dnsName), ".")) {
			break
		}
		hostedZones = append(hostedZones, hostedZone)
	}
	return hostedZones, nil
}

// ListResourceRecordSetsByNameAndType returns the record sets with exactly the specified name and type
// (e.g. weighted record sets with different set identifiers).
func (r *Route53) ListResourceRecordSetsByNameAndType(ctx context.Context, hostedZoneId *string, name *string, recordType types.RRType) ([]types.ResourceRecordSet, error) {
	recordSets := []types.ResourceRecordSet{}

	paginator := route53.NewListResourceRecordSetsPaginator(r.client, &route53.ListResourceRecordSetsInput{
		HostedZoneId:    hostedZoneId,
		StartRecordName: name,
		StartRecordType: recordType,
	})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, &ClientError{
				ResourceName: hostedZoneId,
				Err:          err,
			}
		}
		// Record sets are sorted by name and type starting from the specified ones, so stop at the first other one.
		for _, recordSet := range output.ResourceRecordSets {
			if !strings.EqualFold(strings.TrimSuffix(aws.ToString(recordSet.Name), "."), strings.TrimSuffix(aws.ToString(name), ".")) || recordSet.Type != recordType {
				return recordSets, nil
			}
			recordSets = append(recordSets, recordSet)
		}
	}

	return recordSets, nil
}
