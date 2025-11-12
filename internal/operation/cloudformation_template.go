package operation

import (
	"encoding/json"
	"fmt"
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
// Returns: (modifiedTemplate, changed, error) where changed is true if any DeletionPolicy was removed.
func removeDeletionPolicyFromTemplate(template *string) (string, bool, error) {
	if template == nil || *template == "" {
		return "", false, nil
	}

	// Try to parse as JSON first
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(*template), &data); err == nil {
		// It's JSON - process and return as JSON
		changed := removeDeletionPolicyFromResources(data)
		if !changed {
			return *template, false, nil
		}

		// Check if template is minified by looking for actual newline characters (line breaks).
		// Note: Escaped newlines (\n) within JSON string property values are two characters
		// (backslash and 'n') and do not affect this check.
		isMinified := !strings.Contains(*template, "\n")

		var result []byte
		var marshalErr error
		if isMinified {
			result, marshalErr = json.Marshal(data)
		} else {
			result, marshalErr = json.MarshalIndent(data, "", "  ")
		}

		// Note: This error should not occur in practice because data that was successfully
		// unmarshaled can always be marshaled back. This check is defensive programming.
		if marshalErr != nil {
			return "", false, fmt.Errorf("RemoveDeletionPolicyError: failed to update template for DeletionPolicy removal: %w", marshalErr)
		}
		return string(result), changed, nil
	}

	// Try to parse as YAML
	if err := yaml.Unmarshal([]byte(*template), &data); err == nil {
		// It's YAML - process and return as YAML
		changed := removeDeletionPolicyFromResources(data)
		if !changed {
			return *template, false, nil
		}

		result, marshalErr := yaml.Marshal(data)
		// Note: This error should not occur in practice because data that was successfully
		// unmarshaled can always be marshaled back. This check is defensive programming.
		if marshalErr != nil {
			return "", false, fmt.Errorf("RemoveDeletionPolicyError: failed to update template for DeletionPolicy removal: %w", marshalErr)
		}
		return strings.TrimSuffix(string(result), "\n"), changed, nil
	}

	// Note: Never reached: template must be either valid JSON or valid YAML
	return "", false, fmt.Errorf("RemoveDeletionPolicyError: failed to update template for DeletionPolicy removal because template is neither valid JSON nor valid YAML")
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
