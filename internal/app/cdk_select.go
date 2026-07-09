package app

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-to-k/delstack/internal/cdk"
	"github.com/go-to-k/delstack/internal/io"
	"github.com/go-to-k/delstack/internal/operation"
)

type CdkStackSelector struct {
	stackNames      []string
	interactiveMode bool
	forceMode       bool
}

func NewCdkStackSelector(stackNames []string, interactiveMode, forceMode bool) *CdkStackSelector {
	return &CdkStackSelector{
		stackNames:      stackNames,
		interactiveMode: interactiveMode,
		forceMode:       forceMode,
	}
}

func (s *CdkStackSelector) Select(stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	if len(s.stackNames) > 0 {
		selected, unmatched, err := s.matchByPatterns(stacks)
		if err != nil {
			return nil, err
		}
		if len(unmatched) > 0 {
			return nil, fmt.Errorf("stacks not found in CDK app: %s", strings.Join(unmatched, ", "))
		}
		return selected, nil
	}

	if s.interactiveMode {
		return s.selectInteractively(stacks)
	}

	return stacks, nil
}

// isGlobPattern returns true if the pattern contains glob special characters.
func (s *CdkStackSelector) isGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

// matchByPatterns matches stack names against the given patterns.
// Patterns without glob characters are matched exactly.
// Patterns with glob characters (*, ?, [...]) use path.Match semantics.
// Returns matched stacks, unmatched patterns, and any error from invalid patterns.
func (s *CdkStackSelector) matchByPatterns(stacks []cdk.StackInfo) ([]cdk.StackInfo, []string, error) {
	var selected []cdk.StackInfo
	seen := make(map[string]struct{})

	// Split patterns into exact names and glob patterns
	exactSet := make(map[string]struct{})
	matchedExact := make(map[string]struct{})
	var globs []string
	for _, p := range s.stackNames {
		if s.isGlobPattern(p) {
			globs = append(globs, p)
		} else {
			exactSet[p] = struct{}{}
		}
	}

	for _, st := range stacks {
		// Dedup by Identifier, not StackName, so cross-region stacks that share the
		// same CloudFormation stack name are both selectable.
		id := stackIdentity(st)
		if _, ok := seen[id]; ok {
			continue
		}

		// Check exact match. The name is not removed from exactSet so it can also
		// match another stack with the same name in a different region.
		if _, ok := exactSet[st.StackName]; ok {
			selected = append(selected, st)
			seen[id] = struct{}{}
			matchedExact[st.StackName] = struct{}{}
			continue
		}

		// Check glob patterns
		for _, g := range globs {
			matched, err := path.Match(g, st.StackName)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid glob pattern %q: %w", g, err)
			}
			if matched {
				selected = append(selected, st)
				seen[id] = struct{}{}
				break
			}
		}
	}

	// Collect unmatched exact names (glob patterns that match nothing are not errors)
	var unmatched []string
	for name := range exactSet {
		if _, ok := matchedExact[name]; !ok {
			unmatched = append(unmatched, name)
		}
	}

	return selected, unmatched, nil
}

func (s *CdkStackSelector) selectInteractively(stacks []cdk.StackInfo) ([]cdk.StackInfo, error) {
	// Build display names and filter out TP stacks without forceMode
	var displayStacks []cdk.StackInfo
	var displayNames []string
	for _, st := range stacks {
		if st.TerminationProtection && !s.forceMode {
			continue
		}
		displayStacks = append(displayStacks, st)
		name := fmt.Sprintf("%s (%s)", st.StackName, st.Region)
		if st.TerminationProtection {
			name = operation.TerminationProtectionMarker + name
		}
		displayNames = append(displayNames, name)
	}

	if len(displayStacks) == 0 {
		return nil, nil
	}

	label := []string{"Select stacks to delete."}
	if s.forceMode {
		label = append(label, "(* = TerminationProtection)")
	} else {
		label = append(label, "EnableTerminationProtection stacks are not displayed.")
	}

	selectedNames, continuation, err := io.GetCheckboxes(label, displayNames, false)
	if err != nil {
		return nil, err
	}
	if !continuation {
		return nil, nil
	}

	selectedSet := make(map[string]struct{})
	for _, name := range selectedNames {
		selectedSet[name] = struct{}{}
	}

	var selected []cdk.StackInfo
	for i, st := range displayStacks {
		if _, ok := selectedSet[displayNames[i]]; ok {
			selected = append(selected, st)
		}
	}

	return selected, nil
}
