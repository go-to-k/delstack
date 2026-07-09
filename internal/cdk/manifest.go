package cdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StackInfo struct {
	// Identifier is the unique Cloud Assembly artifact key. Unlike StackName it is
	// guaranteed unique even when cross-region stacks share the same CloudFormation
	// stack name (e.g. a CloudFront us-east-1 support stack reusing the main stack
	// name). It is used as the identity for dependency-graph resolution.
	Identifier string
	StackName  string
	Region     string
	Account    string
	// Dependencies holds the Identifier (artifact key) of each stack this stack
	// depends on, not the StackName, to stay unambiguous when names collide.
	Dependencies          []string
	TerminationProtection bool
}

type manifest struct {
	Artifacts map[string]artifact `json:"artifacts"`
}

type artifact struct {
	Type         string        `json:"type"`
	Environment  string        `json:"environment"`
	Properties   artifactProps `json:"properties"`
	Dependencies []string      `json:"dependencies"`
	DisplayName  string        `json:"displayName"`
}

type artifactProps struct {
	TemplateFile  string `json:"templateFile"`
	DirectoryName string `json:"directoryName"`
	StackName     string `json:"stackName"`
}

// ParseManifest parses a Cloud Assembly manifest.json and returns all stacks,
// including stacks inside nested assemblies (CDK Stages).
func ParseManifest(cdkOutDir string) ([]StackInfo, error) {
	return parseManifestRecursive(cdkOutDir)
}

func parseManifestRecursive(dir string) ([]StackInfo, error) {
	manifestPath := filepath.Join(dir, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest.json: %w", err)
	}

	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}

	// Build a set of stack artifact keys for dependency filtering
	stackArtifactKeys := make(map[string]struct{})
	for key, art := range m.Artifacts {
		if art.Type != "aws:cloudformation:stack" {
			continue
		}
		stackArtifactKeys[key] = struct{}{}
	}

	var stacks []StackInfo
	for key, art := range m.Artifacts {
		switch art.Type {
		case "aws:cloudformation:stack":
			name := resolveStackName(key, art)

			account, region := parseEnvironment(art.Environment)

			// Filter dependencies: only include other stack artifacts (exclude .assets etc.).
			// Keep the artifact key (unique) instead of the resolved stack name, which can
			// collide across regions.
			var deps []string
			for _, dep := range art.Dependencies {
				if _, ok := stackArtifactKeys[dep]; ok {
					deps = append(deps, dep)
				}
			}

			stacks = append(stacks, StackInfo{
				Identifier:   key,
				StackName:    name,
				Region:       region,
				Account:      account,
				Dependencies: deps,
			})

		case "cdk:cloud-assembly":
			// Nested assembly (CDK Stage) — recurse into subdirectory
			nestedDir := art.Properties.DirectoryName
			if nestedDir == "" {
				continue
			}
			nestedPath := filepath.Join(dir, nestedDir)
			nestedStacks, err := parseManifestRecursive(nestedPath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nested assembly %s: %w", nestedDir, err)
			}
			stacks = append(stacks, nestedStacks...)
		}
	}

	return stacks, nil
}

// resolveStackName returns the CloudFormation stack name.
// Priority: properties.stackName > artifact key.
//
// displayName is intentionally NOT used as the stack name. For nested constructs
// (CDK Stages, integ-runner DeployAssert stacks) displayName is a human-readable
// construct path like "Foo/Bar/Baz" that violates the CloudFormation stack name
// constraint ([a-zA-Z][-a-zA-Z0-9]*). The real CFN stack name is the artifact key,
// which CDK derives by sanitizing that path.
func resolveStackName(artifactKey string, art artifact) string {
	if art.Properties.StackName != "" {
		return art.Properties.StackName
	}
	return artifactKey
}

func parseEnvironment(env string) (account, region string) {
	// Format: "aws://ACCOUNT/REGION"
	env = strings.TrimPrefix(env, "aws://")
	parts := strings.SplitN(env, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "unknown-account", "unknown-region"
}
