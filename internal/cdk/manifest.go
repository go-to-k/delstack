package cdk

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StackInfo struct {
	StackName    string
	Region       string
	Account      string
	Dependencies []string
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
	stackArtifactKeys := make(map[string]string) // artifact key -> stack name
	for key, art := range m.Artifacts {
		if art.Type != "aws:cloudformation:stack" {
			continue
		}
		stackArtifactKeys[key] = resolveStackName(key, art)
	}

	var stacks []StackInfo
	for key, art := range m.Artifacts {
		switch art.Type {
		case "aws:cloudformation:stack":
			name := resolveStackName(key, art)

			account, region := parseEnvironment(art.Environment)

			// Filter dependencies: only include other stack artifacts (exclude .assets etc.)
			var deps []string
			for _, dep := range art.Dependencies {
				if depName, ok := stackArtifactKeys[dep]; ok {
					deps = append(deps, depName)
				}
			}

			stacks = append(stacks, StackInfo{
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
// Priority: properties.stackName > displayName > artifact key
func resolveStackName(artifactKey string, art artifact) string {
	if art.Properties.StackName != "" {
		return art.Properties.StackName
	}
	if art.DisplayName != "" {
		return art.DisplayName
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
