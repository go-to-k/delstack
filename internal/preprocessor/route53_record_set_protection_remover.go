package preprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/resourcetype"
	"github.com/go-to-k/delstack/pkg/client"
	"gopkg.in/yaml.v3"
)

var _ IPreprocessor = (*Route53RecordSetProtectionRemover)(nil)

// Route53RecordSetProtectionRemover handles resource record set protections on the record sets of the stack
// in hosted zones outside the stack, which block deleting the record sets.
// Protections in hosted zones inside the stack are handled by DeletionProtectionRemover.
type Route53RecordSetProtectionRemover struct {
	forceMode     bool
	cfnClient     client.ICloudFormation
	route53Client client.IRoute53
}

type recordSetTarget struct {
	resourceType      string
	logicalResourceId string
	hostedZoneId      string
	name              string
	recordType        route53Types.RRType
}

type protectedRecordSet struct {
	target     recordSetTarget
	protection route53Types.ResourceRecordSetProtection
}

func NewRoute53RecordSetProtectionRemover(
	forceMode bool,
	cfnClient client.ICloudFormation,
	route53Client client.IRoute53,
) *Route53RecordSetProtectionRemover {
	return &Route53RecordSetProtectionRemover{
		forceMode:     forceMode,
		cfnClient:     cfnClient,
		route53Client: route53Client,
	}
}

func (r *Route53RecordSetProtectionRemover) Preprocess(ctx context.Context, stackName *string, resources []types.StackResourceSummary) error {
	recordSetResources := append(
		FilterResourcesByType(resources, resourcetype.Route53RecordSet),
		FilterResourcesByType(resources, resourcetype.Route53RecordSetGroup)...,
	)
	if len(recordSetResources) == 0 {
		return nil
	}

	targets, err := r.findTargets(ctx, stackName, resources, recordSetResources)
	if err != nil {
		return err
	}

	protectedRecordSets, err := r.findProtectedRecordSets(ctx, stackName, targets)
	if err != nil {
		return err
	}
	if len(protectedRecordSets) == 0 {
		return nil
	}

	if !r.forceMode {
		return r.buildProtectionError(protectedRecordSets)
	}

	return r.deleteProtections(ctx, stackName, protectedRecordSets)
}

func (r *Route53RecordSetProtectionRemover) findTargets(
	ctx context.Context,
	stackName *string,
	resources []types.StackResourceSummary,
	recordSetResources []types.StackResourceSummary,
) ([]recordSetTarget, error) {
	templateBody, err := r.cfnClient.GetTemplate(ctx, stackName)
	if err != nil {
		if isAccessDenied(err) {
			io.Logger.Warn().Msgf("[%v]: Skipped checking resource record set protections for record sets due to missing permission: %v", aws.ToString(stackName), err)
			return nil, nil
		}
		return nil, fmt.Errorf("DeletionProtectionError: failed to get template to check resource record set protections: %w", err)
	}

	template, err := parseTemplate(aws.ToString(templateBody))
	if err != nil {
		io.Logger.Warn().Msgf("[%v]: Skipped checking resource record set protections for record sets because the template could not be parsed: %v", aws.ToString(stackName), err)
		return nil, nil
	}
	templateResources, _ := template["Resources"].(map[string]any)

	stackResources := map[string]types.StackResourceSummary{}
	inStackHostedZoneIds := map[string]bool{}
	for _, resource := range resources {
		stackResources[aws.ToString(resource.LogicalResourceId)] = resource
		if aws.ToString(resource.ResourceType) == resourcetype.Route53HostedZone {
			inStackHostedZoneIds[normalizeHostedZoneId(aws.ToString(resource.PhysicalResourceId))] = true
		}
	}

	resolver := &hostedZoneResolver{
		route53Client:  r.route53Client,
		stackResources: stackResources,
		zoneIdsByName:  map[string]string{},
	}

	targets := []recordSetTarget{}
	for _, resource := range recordSetResources {
		logicalId := aws.ToString(resource.LogicalResourceId)
		resourceType := aws.ToString(resource.ResourceType)
		skip := func(reason string) {
			io.Logger.Warn().Msgf("[%v]: Skipped checking resource record set protections for %s: %s (%s)", aws.ToString(stackName), resourceType, logicalId, reason)
		}

		definition, _ := templateResources[logicalId].(map[string]any)
		properties, _ := definition["Properties"].(map[string]any)
		if properties == nil {
			skip("properties not found in the template")
			continue
		}

		var entries []map[string]any
		var group map[string]any
		if resourceType == resourcetype.Route53RecordSet {
			entries = []map[string]any{properties}
		} else {
			group = properties
			recordSets, _ := properties["RecordSets"].([]any)
			for _, recordSet := range recordSets {
				if entry, ok := recordSet.(map[string]any); ok {
					entries = append(entries, entry)
				}
			}
		}

		for _, entry := range entries {
			hostedZoneId, reason, err := resolver.resolve(ctx, entry, group)
			if err != nil {
				return nil, err
			}
			if hostedZoneId == "" {
				skip(reason)
				continue
			}
			if inStackHostedZoneIds[hostedZoneId] {
				continue
			}

			name, ok := entry["Name"].(string)
			if !ok && resourceType == resourcetype.Route53RecordSet {
				// The physical resource ID of a RecordSet is its record name.
				name, ok = aws.ToString(resource.PhysicalResourceId), resource.PhysicalResourceId != nil
			}
			recordType, typeOk := entry["Type"].(string)
			if !ok || !typeOk {
				skip("record name or type is not a literal value in the template")
				continue
			}

			targets = append(targets, recordSetTarget{
				resourceType:      resourceType,
				logicalResourceId: logicalId,
				hostedZoneId:      hostedZoneId,
				name:              name,
				recordType:        route53Types.RRType(recordType),
			})
		}
	}

	return targets, nil
}

func (r *Route53RecordSetProtectionRemover) findProtectedRecordSets(ctx context.Context, stackName *string, targets []recordSetTarget) ([]protectedRecordSet, error) {
	targetsByZone := map[string][]recordSetTarget{}
	hostedZoneIds := []string{}
	for _, target := range targets {
		if _, ok := targetsByZone[target.hostedZoneId]; !ok {
			hostedZoneIds = append(hostedZoneIds, target.hostedZoneId)
		}
		targetsByZone[target.hostedZoneId] = append(targetsByZone[target.hostedZoneId], target)
	}

	results := []protectedRecordSet{}
	for _, hostedZoneId := range hostedZoneIds {
		protections, err := r.route53Client.ListResourceRecordSetProtections(ctx, aws.String(hostedZoneId))
		if err != nil {
			if isAccessDenied(err) {
				io.Logger.Warn().Msgf("[%v]: Skipped checking resource record set protections in hosted zone %s due to missing permission: %v", aws.ToString(stackName), hostedZoneId, err)
				continue
			}
			return nil, fmt.Errorf("DeletionProtectionError: failed to list resource record set protections in hosted zone %s: %w", hostedZoneId, err)
		}

		for _, target := range targetsByZone[hostedZoneId] {
			for _, protection := range protections {
				if normalizeRecordName(aws.ToString(protection.RecordName)) == normalizeRecordName(target.name) &&
					protection.RecordType == target.recordType {
					results = append(results, protectedRecordSet{target: target, protection: protection})
				}
			}
		}
	}

	return results, nil
}

func (r *Route53RecordSetProtectionRemover) deleteProtections(ctx context.Context, stackName *string, protectedRecordSets []protectedRecordSet) error {
	protectionsByZone := map[string][]route53Types.ResourceRecordSetProtection{}
	hostedZoneIds := []string{}
	for _, p := range protectedRecordSets {
		protections, ok := protectionsByZone[p.target.hostedZoneId]
		if !ok {
			hostedZoneIds = append(hostedZoneIds, p.target.hostedZoneId)
		}
		// The same protection can match multiple record sets in the stack, and each can be deleted only once.
		if slices.ContainsFunc(protections, func(existing route53Types.ResourceRecordSetProtection) bool {
			return aws.ToString(existing.RecordName) == aws.ToString(p.protection.RecordName) && existing.RecordType == p.protection.RecordType
		}) {
			continue
		}
		protectionsByZone[p.target.hostedZoneId] = append(protections, p.protection)
	}

	for _, hostedZoneId := range hostedZoneIds {
		for chunk := range slices.Chunk(protectionsByZone[hostedZoneId], route53MaxProtectionsPerRequest) {
			if err := r.route53Client.DeleteResourceRecordSetProtections(ctx, aws.String(hostedZoneId), chunk); err != nil {
				return fmt.Errorf("DeletionProtectionError: failed to delete resource record set protections in hosted zone %s: %w", hostedZoneId, err)
			}
		}
	}

	for _, p := range protectedRecordSets {
		io.Logger.Info().Msgf("[%v]: Deleted resource record set protection for %s: %s (%s %s in hosted zone %s)",
			aws.ToString(stackName), p.target.resourceType, p.target.logicalResourceId, p.target.name, p.target.recordType, p.target.hostedZoneId)
	}
	return nil
}

func (r *Route53RecordSetProtectionRemover) buildProtectionError(protectedRecordSets []protectedRecordSet) error {
	lines := make([]string, 0, len(protectedRecordSets))
	for _, p := range protectedRecordSets {
		lines = append(lines, fmt.Sprintf("- %s: %s (%s %s in hosted zone %s)", p.target.resourceType, p.target.logicalResourceId, p.target.name, p.target.recordType, p.target.hostedZoneId))
	}
	return fmt.Errorf("DeletionProtectionError: the following record sets have resource record set protections enabled:\n  %s\nuse the -f option to force delete the protections and delete the stack", strings.Join(lines, "\n  "))
}

type hostedZoneResolver struct {
	route53Client  client.IRoute53
	stackResources map[string]types.StackResourceSummary
	zoneIdsByName  map[string]string
}

// resolve returns the hosted zone ID of a record set, or an empty ID with the reason when it cannot be resolved.
// A record set in a RecordSetGroup inherits the hosted zone of the group when it does not specify its own.
func (h *hostedZoneResolver) resolve(ctx context.Context, entry map[string]any, group map[string]any) (string, string, error) {
	for _, properties := range []map[string]any{entry, group} {
		if properties == nil {
			continue
		}

		if value, ok := properties["HostedZoneId"]; ok {
			if id, ok := value.(string); ok {
				return normalizeHostedZoneId(id), "", nil
			}
			if ref, ok := value.(map[string]any)["Ref"].(string); ok {
				if resource, ok := h.stackResources[ref]; ok && aws.ToString(resource.ResourceType) == resourcetype.Route53HostedZone {
					return normalizeHostedZoneId(aws.ToString(resource.PhysicalResourceId)), "", nil
				}
			}
			return "", "HostedZoneId cannot be resolved from the template", nil
		}

		if value, ok := properties["HostedZoneName"]; ok {
			name, ok := value.(string)
			if !ok {
				return "", "HostedZoneName is not a literal value in the template", nil
			}
			return h.resolveByName(ctx, name)
		}
	}
	return "", "neither HostedZoneId nor HostedZoneName is specified", nil
}

func (h *hostedZoneResolver) resolveByName(ctx context.Context, name string) (string, string, error) {
	key := normalizeRecordName(name)
	if id, ok := h.zoneIdsByName[key]; ok {
		return id, "", nil
	}

	hostedZones, err := h.route53Client.ListHostedZonesByName(ctx, aws.String(name))
	if err != nil {
		if isAccessDenied(err) {
			return "", fmt.Sprintf("missing permission to find hosted zone %s: %v", name, err), nil
		}
		return "", "", fmt.Errorf("DeletionProtectionError: failed to find hosted zone %s: %w", name, err)
	}
	if len(hostedZones) != 1 {
		return "", fmt.Sprintf("found %d hosted zones named %s", len(hostedZones), name), nil
	}

	id := normalizeHostedZoneId(aws.ToString(hostedZones[0].Id))
	h.zoneIdsByName[key] = id
	return id, "", nil
}

func normalizeHostedZoneId(id string) string {
	return strings.TrimPrefix(id, "/hostedzone/")
}

// Route 53 returns names in lowercase with a trailing dot, and "*" escaped as "\052".
func normalizeRecordName(name string) string {
	name = strings.ToLower(strings.ReplaceAll(name, `\052`, "*"))
	if !strings.HasSuffix(name, ".") {
		name += "."
	}
	return name
}

// parseTemplate parses a JSON or YAML template. YAML short-form intrinsic functions (e.g. !Ref) are converted
// to their long form (e.g. {"Ref": ...}), so that they are not mistaken for literal values.
func parseTemplate(body string) (map[string]any, error) {
	var template map[string]any
	if err := json.Unmarshal([]byte(body), &template); err == nil {
		return template, nil
	}

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(body), &node); err != nil {
		return nil, err
	}
	value, err := yamlNodeToValue(&node)
	if err != nil {
		return nil, err
	}
	template, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("template is not a mapping")
	}
	return template, nil
}

func yamlNodeToValue(node *yaml.Node) (any, error) {
	var value any
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, fmt.Errorf("empty YAML document")
		}
		return yamlNodeToValue(node.Content[0])
	case yaml.AliasNode:
		return yamlNodeToValue(node.Alias)
	case yaml.MappingNode:
		mapping := map[string]any{}
		for i := 0; i+1 < len(node.Content); i += 2 {
			v, err := yamlNodeToValue(node.Content[i+1])
			if err != nil {
				return nil, err
			}
			mapping[node.Content[i].Value] = v
		}
		value = mapping
	case yaml.SequenceNode:
		sequence := []any{}
		for _, child := range node.Content {
			v, err := yamlNodeToValue(child)
			if err != nil {
				return nil, err
			}
			sequence = append(sequence, v)
		}
		value = sequence
	case yaml.ScalarNode:
		if isIntrinsicFunctionTag(node.Tag) {
			value = node.Value
		} else if err := node.Decode(&value); err != nil {
			return nil, err
		}
	}

	if !isIntrinsicFunctionTag(node.Tag) {
		return value, nil
	}
	name := strings.TrimPrefix(node.Tag, "!")
	if name == "Ref" || name == "Condition" {
		return map[string]any{name: value}, nil
	}
	return map[string]any{"Fn::" + name: value}, nil
}

func isIntrinsicFunctionTag(tag string) bool {
	return strings.HasPrefix(tag, "!") && !strings.HasPrefix(tag, "!!")
}
