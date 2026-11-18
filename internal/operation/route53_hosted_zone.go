package operation

import (
	"cmp"
	"context"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/pkg/client"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// ChangeResourceRecordSets limits per request.
// See: https://docs.aws.amazon.com/Route53/latest/APIReference/API_ChangeResourceRecordSets.html
const (
	route53MaxResourceRecordsPerBatch = 1000
	route53MaxValueCharsPerBatch      = 32000
)

const (
	route53WaitMaxAttempts = 60
	route53WaitInterval    = 5 * time.Second
)

var _ IOperator = (*Route53HostedZoneOperator)(nil)

type Route53HostedZoneOperator struct {
	client       client.IRoute53
	resources    []*types.StackResourceSummary
	waitInterval time.Duration
}

func NewRoute53HostedZoneOperator(client client.IRoute53) *Route53HostedZoneOperator {
	return &Route53HostedZoneOperator{
		client:       client,
		resources:    []*types.StackResourceSummary{},
		waitInterval: route53WaitInterval,
	}
}

func (o *Route53HostedZoneOperator) AddResource(resource *types.StackResourceSummary) {
	o.resources = append(o.resources, resource)
}

func (o *Route53HostedZoneOperator) GetResourcesLength() int {
	return len(o.resources)
}

func (o *Route53HostedZoneOperator) DeleteResources(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))

	for _, resource := range o.resources {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		eg.Go(func() (err error) {
			defer sem.Release(1)

			return o.DeleteRoute53HostedZone(ctx, resource.PhysicalResourceId)
		})
	}

	err := eg.Wait()
	return err
}

func (o *Route53HostedZoneOperator) DeleteRoute53HostedZone(ctx context.Context, hostedZoneId *string) error {
	hostedZone, exists, err := o.client.GetHostedZone(ctx, hostedZoneId)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	hostedZoneName := aws.ToString(hostedZone.Name)

	// DNSSEC is not supported for private hosted zones.
	if hostedZone.Config == nil || !hostedZone.Config.PrivateZone {
		if err = o.cleanupDNSSEC(ctx, hostedZoneId); err != nil {
			return err
		}
	}

	recordSets, err := o.client.ListResourceRecordSets(ctx, hostedZoneId)
	if err != nil {
		return err
	}

	// Record sets created by traffic policy instances must be deleted via the instances,
	// otherwise the instances remain and continue to be charged.
	if trafficPolicyInstanceIds := collectTrafficPolicyInstanceIds(recordSets); len(trafficPolicyInstanceIds) > 0 {
		for _, id := range trafficPolicyInstanceIds {
			if err = o.client.DeleteTrafficPolicyInstance(ctx, aws.String(id)); err != nil {
				return err
			}
		}
		recordSets, err = o.waitForTrafficPolicyRecordSetsDeleted(ctx, hostedZoneId)
		if err != nil {
			return err
		}
	}

	nonDefaultRecordSets := []route53Types.ResourceRecordSet{}
	for _, recordSet := range recordSets {
		if isDefaultRecordSet(recordSet, hostedZoneName) {
			continue
		}
		nonDefaultRecordSets = append(nonDefaultRecordSets, recordSet)
	}

	// Delete alias records first in case they target other record sets in the same zone and the deletion spans multiple batches.
	slices.SortStableFunc(nonDefaultRecordSets, func(a, b route53Types.ResourceRecordSet) int {
		return cmp.Compare(aliasSortKey(a), aliasSortKey(b))
	})

	// Sequential to avoid PriorRequestNotComplete and throttling on the same hosted zone.
	for _, batch := range chunkRecordSetsForChangeBatch(nonDefaultRecordSets) {
		if err = o.client.DeleteResourceRecordSets(ctx, hostedZoneId, batch); err != nil {
			return err
		}
	}

	return o.client.DeleteHostedZone(ctx, hostedZoneId)
}

// A hosted zone with DNSSEC signing enabled or with key-signing keys cannot be deleted.
func (o *Route53HostedZoneOperator) cleanupDNSSEC(ctx context.Context, hostedZoneId *string) error {
	status, keySigningKeys, err := o.client.GetDNSSEC(ctx, hostedZoneId)
	if err != nil {
		// Do not block deletions that worked before for roles without the permission.
		// If DNSSEC is in use, deleting the hosted zone fails instead.
		if client.IsAccessDeniedError(err) {
			io.Logger.Warn().Msgf("Skipped checking DNSSEC for hosted zone %s due to missing permission: %v", aws.ToString(hostedZoneId), err)
			return nil
		}
		return err
	}

	if !isDNSSECNotSigning(status) {
		if err = o.client.DisableHostedZoneDNSSEC(ctx, hostedZoneId); err != nil {
			return err
		}
		keySigningKeys, err = o.waitForDNSSEC(ctx, hostedZoneId, "DNSSEC signing to be disabled", func(status *route53Types.DNSSECStatus, _ []route53Types.KeySigningKey) bool {
			return isDNSSECNotSigning(status)
		})
		if err != nil {
			return err
		}
	}

	// Key-signing keys must be deactivated before they are deleted.
	deactivated := false
	for _, keySigningKey := range keySigningKeys {
		if aws.ToString(keySigningKey.Status) != "ACTIVE" {
			continue
		}
		if err = o.client.DeactivateKeySigningKey(ctx, hostedZoneId, keySigningKey.Name); err != nil {
			return err
		}
		deactivated = true
	}
	if deactivated {
		keySigningKeys, err = o.waitForDNSSEC(ctx, hostedZoneId, "key-signing keys to be deactivated", func(_ *route53Types.DNSSECStatus, keySigningKeys []route53Types.KeySigningKey) bool {
			return !slices.ContainsFunc(keySigningKeys, func(k route53Types.KeySigningKey) bool {
				return aws.ToString(k.Status) == "ACTIVE"
			})
		})
		if err != nil {
			return err
		}
	}

	for _, keySigningKey := range keySigningKeys {
		if err = o.client.DeleteKeySigningKey(ctx, hostedZoneId, keySigningKey.Name); err != nil {
			return err
		}
	}
	return nil
}

func (o *Route53HostedZoneOperator) waitForDNSSEC(
	ctx context.Context,
	hostedZoneId *string,
	description string,
	done func(*route53Types.DNSSECStatus, []route53Types.KeySigningKey) bool,
) ([]route53Types.KeySigningKey, error) {
	for range route53WaitMaxAttempts {
		status, keySigningKeys, err := o.client.GetDNSSEC(ctx, hostedZoneId)
		if err != nil {
			return nil, err
		}
		if done(status, keySigningKeys) {
			return keySigningKeys, nil
		}

		select {
		case <-ctx.Done():
			return nil, &client.ClientError{
				ResourceName: hostedZoneId,
				Err:          ctx.Err(),
			}
		case <-time.After(o.waitInterval):
		}
	}

	return nil, &client.ClientError{
		ResourceName: hostedZoneId,
		Err:          fmt.Errorf("timed out waiting for %s", description),
	}
}

func isDNSSECNotSigning(status *route53Types.DNSSECStatus) bool {
	return status == nil || aws.ToString(status.ServeSignature) == "NOT_SIGNING"
}

func (o *Route53HostedZoneOperator) waitForTrafficPolicyRecordSetsDeleted(ctx context.Context, hostedZoneId *string) ([]route53Types.ResourceRecordSet, error) {
	for range route53WaitMaxAttempts {
		recordSets, err := o.client.ListResourceRecordSets(ctx, hostedZoneId)
		if err != nil {
			return nil, err
		}
		if len(collectTrafficPolicyInstanceIds(recordSets)) == 0 {
			return recordSets, nil
		}

		select {
		case <-ctx.Done():
			return nil, &client.ClientError{
				ResourceName: hostedZoneId,
				Err:          ctx.Err(),
			}
		case <-time.After(o.waitInterval):
		}
	}

	return nil, &client.ClientError{
		ResourceName: hostedZoneId,
		Err:          fmt.Errorf("timed out waiting for record sets of traffic policy instances to be deleted"),
	}
}

func collectTrafficPolicyInstanceIds(recordSets []route53Types.ResourceRecordSet) []string {
	ids := []string{}
	for _, recordSet := range recordSets {
		if id := aws.ToString(recordSet.TrafficPolicyInstanceId); id != "" && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

func aliasSortKey(recordSet route53Types.ResourceRecordSet) int {
	if recordSet.AliasTarget != nil {
		return 0
	}
	return 1
}

// The SOA and NS records at the zone apex are created by Route 53 and cannot be deleted.
func isDefaultRecordSet(recordSet route53Types.ResourceRecordSet, hostedZoneName string) bool {
	if !strings.EqualFold(aws.ToString(recordSet.Name), hostedZoneName) {
		return false
	}
	return recordSet.Type == route53Types.RRTypeSoa || recordSet.Type == route53Types.RRTypeNs
}

func chunkRecordSetsForChangeBatch(recordSets []route53Types.ResourceRecordSet) [][]route53Types.ResourceRecordSet {
	batches := [][]route53Types.ResourceRecordSet{}
	batch := []route53Types.ResourceRecordSet{}
	records, chars := 0, 0

	for _, recordSet := range recordSets {
		// Alias records have no ResourceRecord elements, but count them as one.
		recordSetRecords := max(len(recordSet.ResourceRecords), 1)
		recordSetChars := 0
		for _, record := range recordSet.ResourceRecords {
			recordSetChars += len(aws.ToString(record.Value))
		}

		if len(batch) > 0 && (records+recordSetRecords > route53MaxResourceRecordsPerBatch || chars+recordSetChars > route53MaxValueCharsPerBatch) {
			batches = append(batches, batch)
			batch = []route53Types.ResourceRecordSet{}
			records, chars = 0, 0
		}

		batch = append(batch, recordSet)
		records += recordSetRecords
		chars += recordSetChars
	}
	if len(batch) > 0 {
		batches = append(batches, batch)
	}

	return batches
}
