package operation

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// removeDeletionPolicyFromTemplate removes DeletionPolicy properties with Retain or RetainExceptOnCreate values
// from CloudFormation templates at the resource level only (not within Properties).
//
// This function uses YAML/JSON parsers to structurally understand the template and only removes
// DeletionPolicy from the resource level, preserving any DeletionPolicy within resource Properties.
//
// Supported formats:
// - YAML (both inline and block formats)
// - JSON (both formatted and minified)
//
// Note: This does NOT remove DeletionPolicy with "Delete" or "Snapshot" values.
// Note: Original formatting (indentation, spacing, property order) may not be preserved.
//
// Returns: (modifiedTemplate, changed) where changed is true if any DeletionPolicy was removed.
func removeDeletionPolicyFromTemplate(template *string) (string, bool) {
	if template == nil || *template == "" {
		return "", false
	}

	// Try to parse as JSON first
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*template), &data); err == nil {
		// It's JSON - process and return as JSON
		isMinified := !strings.Contains(*template, "\n")
		changed := removeDeletionPolicyFromResources(data)

		var result []byte
		var marshalErr error
		if isMinified {
			result, marshalErr = json.Marshal(data)
		} else {
			result, marshalErr = json.MarshalIndent(data, "", "  ")
		}

		if marshalErr != nil {
			return *template, false
		}
		return string(result), changed
	}

	// Try to parse as YAML
	if err := yaml.Unmarshal([]byte(*template), &data); err == nil {
		// It's YAML - process and return as YAML
		changed := removeDeletionPolicyFromResources(data)

		result, marshalErr := yaml.Marshal(data)
		if marshalErr != nil {
			return *template, false
		}
		return strings.TrimSuffix(string(result), "\n"), changed
	}

	// If both fail, return original
	return *template, false
}

// removeDeletionPolicyFromResources removes DeletionPolicy (with Retain/RetainExceptOnCreate values)
// from the Resources section at the resource level only.
// Returns true if any changes were made.
func removeDeletionPolicyFromResources(data map[string]interface{}) bool {
	resources, ok := data["Resources"]
	if !ok {
		return false
	}

	resourcesMap, ok := resources.(map[string]interface{})
	if !ok {
		return false
	}

	changed := false
	// Iterate through each resource
	for _, resource := range resourcesMap {
		resourceMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if DeletionPolicy exists at resource level
		deletionPolicy, exists := resourceMap["DeletionPolicy"]
		if !exists {
			continue
		}

		// Check if the value is "Retain" or "RetainExceptOnCreate"
		deletionPolicyStr, ok := deletionPolicy.(string)
		if !ok {
			continue
		}

		if deletionPolicyStr == "Retain" || deletionPolicyStr == "RetainExceptOnCreate" {
			delete(resourceMap, "DeletionPolicy")
			changed = true
		}
	}
	return changed
}
