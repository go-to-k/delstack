package operation

import (
	"regexp"
	"strings"
)

// Regular expression pattern components for DeletionPolicy removal
const (
	optionalQuote      = `["']?`
	retainValues       = `(?:Retain|RetainExceptOnCreate)`
	deletionPolicyKey  = optionalQuote + `DeletionPolicy` + optionalQuote
	deletionPolicyPair = deletionPolicyKey + `\s*:\s*` + optionalQuote + retainValues + optionalQuote
)

// removeDeletionPolicyFromTemplate removes DeletionPolicy properties with Retain or RetainExceptOnCreate values
// from CloudFormation templates while preserving the original formatting.
//
// This function uses a line-based string processing approach instead of YAML/JSON parsers to ensure that:
// - Original indentation (spaces/tabs) is completely preserved
// - Property order remains unchanged
// - Line breaks and whitespace are maintained exactly as in the input
//
// Supported formats:
// - YAML inline: "DeletionPolicy: Retain"
// - YAML block: "DeletionPolicy:\n  Retain"
// - JSON formatted: "\"DeletionPolicy\": \"Retain\""
// - JSON minified: single-line JSON without newlines
//
// Note: This does NOT remove DeletionPolicy with "Delete" or "Snapshot" values.
func removeDeletionPolicyFromTemplate(template *string) string {
	// Handle minified JSON (single line)
	if !strings.Contains(*template, "\n") {
		return removeFromMinifiedJSON(*template)
	}

	// Handle multi-line templates (YAML or formatted JSON)
	return removeFromMultiLine(*template)
}

// removeFromMinifiedJSON removes DeletionPolicy from single-line (minified) JSON templates.
// It handles comma placement to maintain valid JSON syntax after removal.
func removeFromMinifiedJSON(template string) string {
	// For minified JSON, use a simpler approach: match the entire key-value with surrounding commas
	// Match: "DeletionPolicy":"Retain", or ,"DeletionPolicy":"Retain" or "DeletionPolicy":"Retain"
	result := regexp.MustCompile(deletionPolicyPair+`\s*,\s*`).ReplaceAllString(template, "")
	result = regexp.MustCompile(`,\s*`+deletionPolicyPair+`\s*`).ReplaceAllString(result, "")
	return result
}

// removeFromMultiLine removes DeletionPolicy from multi-line templates (formatted JSON or YAML).
// It preserves the original indentation, line breaks, and property order by processing line by line.
// Supports both YAML inline format ("DeletionPolicy: Retain") and block format ("DeletionPolicy:\n  Retain").
func removeFromMultiLine(template string) string {
	lines := strings.Split(template, "\n")
	result := make([]string, 0, len(lines))

	// Pattern to match DeletionPolicy lines with Retain or RetainExceptOnCreate (inline format)
	inlinePattern := regexp.MustCompile(`^\s*` + deletionPolicyPair + `\s*,?\s*$`)
	// Pattern for YAML block format: DeletionPolicy key without value on same line
	keyOnlyPattern := regexp.MustCompile(`^\s*` + deletionPolicyKey + `\s*:\s*$`)
	// Pattern for the value line (indented Retain or RetainExceptOnCreate)
	valueOnlyPattern := regexp.MustCompile(`^\s+` + optionalQuote + retainValues + optionalQuote + `\s*$`)
	// Patterns for trailing comma cleanup
	closingBracketPattern := regexp.MustCompile(`^\s*[}\]]`)
	trailingCommaPattern := regexp.MustCompile(`,\s*$`)
	trailingCommaRemover := regexp.MustCompile(`,(\s*)$`)

	skipNext := false
	for i, line := range lines {
		// Skip this line if it was marked by previous iteration
		if skipNext {
			skipNext = false
			continue
		}

		// Check for YAML block format (key on one line, value on next)
		if keyOnlyPattern.MatchString(line) {
			if i+1 < len(lines) && valueOnlyPattern.MatchString(lines[i+1]) {
				// Skip both the key and value lines
				skipNext = true
				continue
			}
		}

		// Check for inline format (key and value on same line)
		if inlinePattern.MatchString(line) {
			// Remove trailing comma from previous line if next line is closing bracket
			if len(result) > 0 && i+1 < len(lines) {
				if closingBracketPattern.MatchString(lines[i+1]) && trailingCommaPattern.MatchString(result[len(result)-1]) {
					result[len(result)-1] = trailingCommaRemover.ReplaceAllString(result[len(result)-1], "$1")
				}
			}
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
