package cdk

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestParseManifest_SingleStack(t *testing.T) {
	dir := t.TempDir()
	manifestJSON := `{
  "version": "52.0.0",
  "artifacts": {
    "MyStack.assets": {
      "type": "cdk:asset-manifest",
      "properties": { "file": "MyStack.assets.json" }
    },
    "MyStack": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/us-east-1",
      "properties": { "templateFile": "MyStack.template.json" },
      "dependencies": ["MyStack.assets"],
      "displayName": "MyStack"
    },
    "Tree": {
      "type": "cdk:tree",
      "properties": { "file": "tree.json" }
    }
  }
}`
	writeManifest(t, dir, manifestJSON)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}

	s := stacks[0]
	if s.StackName != "MyStack" {
		t.Errorf("expected stack name 'MyStack', got '%s'", s.StackName)
	}
	if s.Region != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got '%s'", s.Region)
	}
	if s.Account != "123456789012" {
		t.Errorf("expected account '123456789012', got '%s'", s.Account)
	}
	if len(s.Dependencies) != 0 {
		t.Errorf("expected 0 dependencies, got %d: %v", len(s.Dependencies), s.Dependencies)
	}
}

func TestParseManifest_MultiStackWithDependencies(t *testing.T) {
	dir := t.TempDir()
	manifestJSON := `{
  "version": "52.0.0",
  "artifacts": {
    "NetworkStack.assets": {
      "type": "cdk:asset-manifest",
      "properties": { "file": "NetworkStack.assets.json" }
    },
    "NetworkStack": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/us-east-1",
      "dependencies": ["NetworkStack.assets"],
      "displayName": "NetworkStack"
    },
    "AppStack.assets": {
      "type": "cdk:asset-manifest",
      "properties": { "file": "AppStack.assets.json" }
    },
    "AppStack": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/us-east-1",
      "dependencies": ["NetworkStack", "AppStack.assets"],
      "displayName": "AppStack"
    }
  }
}`
	writeManifest(t, dir, manifestJSON)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}

	// Find AppStack
	var appStack *StackInfo
	for i := range stacks {
		if stacks[i].StackName == "AppStack" {
			appStack = &stacks[i]
			break
		}
	}
	if appStack == nil {
		t.Fatal("AppStack not found")
	}

	if len(appStack.Dependencies) != 1 {
		t.Fatalf("expected 1 dependency for AppStack, got %d: %v", len(appStack.Dependencies), appStack.Dependencies)
	}
	if appStack.Dependencies[0] != "NetworkStack" {
		t.Errorf("expected dependency 'NetworkStack', got '%s'", appStack.Dependencies[0])
	}
}

func TestParseManifest_UnknownRegion(t *testing.T) {
	dir := t.TempDir()
	manifestJSON := `{
  "version": "52.0.0",
  "artifacts": {
    "MyStack": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://unknown-account/unknown-region",
      "displayName": "MyStack"
    }
  }
}`
	writeManifest(t, dir, manifestJSON)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Region != "unknown-region" {
		t.Errorf("expected region 'unknown-region', got '%s'", stacks[0].Region)
	}
}

func TestParseManifest_MultiRegion(t *testing.T) {
	dir := t.TempDir()
	manifestJSON := `{
  "version": "52.0.0",
  "artifacts": {
    "AppStack": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/ap-northeast-1",
      "displayName": "AppStack"
    },
    "EdgeStack": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/us-east-1",
      "displayName": "EdgeStack"
    }
  }
}`
	writeManifest(t, dir, manifestJSON)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}

	regionMap := make(map[string]string)
	for _, s := range stacks {
		regionMap[s.StackName] = s.Region
	}

	if regionMap["AppStack"] != "ap-northeast-1" {
		t.Errorf("expected AppStack region 'ap-northeast-1', got '%s'", regionMap["AppStack"])
	}
	if regionMap["EdgeStack"] != "us-east-1" {
		t.Errorf("expected EdgeStack region 'us-east-1', got '%s'", regionMap["EdgeStack"])
	}
}

func TestParseManifest_FallbackToArtifactKey(t *testing.T) {
	dir := t.TempDir()
	manifestJSON := `{
  "version": "52.0.0",
  "artifacts": {
    "MyStackArtifactKey": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/us-east-1"
    }
  }
}`
	writeManifest(t, dir, manifestJSON)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stacks[0].StackName != "MyStackArtifactKey" {
		t.Errorf("expected stack name 'MyStackArtifactKey', got '%s'", stacks[0].StackName)
	}
}

func TestParseManifest_NoStacks(t *testing.T) {
	dir := t.TempDir()
	manifestJSON := `{
  "version": "52.0.0",
  "artifacts": {
    "Tree": {
      "type": "cdk:tree",
      "properties": { "file": "tree.json" }
    }
  }
}`
	writeManifest(t, dir, manifestJSON)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 0 {
		t.Errorf("expected 0 stacks, got %d", len(stacks))
	}
}

func TestParseManifest_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := ParseManifest(dir)
	if err == nil {
		t.Fatal("expected error for missing manifest.json")
	}
}

func TestParseManifest_NestedAssembly_Stage(t *testing.T) {
	dir := t.TempDir()

	// Top-level manifest with a nested cloud assembly (CDK Stage)
	topManifest := `{
  "version": "52.0.0",
  "artifacts": {
    "assembly-MyStage": {
      "type": "cdk:cloud-assembly",
      "properties": { "directoryName": "assembly-MyStage" }
    },
    "Tree": {
      "type": "cdk:tree",
      "properties": { "file": "tree.json" }
    }
  }
}`
	writeManifest(t, dir, topManifest)

	// Nested assembly manifest with stacks
	nestedDir := filepath.Join(dir, "assembly-MyStage")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	nestedManifest := `{
  "version": "52.0.0",
  "artifacts": {
    "MyStage-StackA.assets": {
      "type": "cdk:asset-manifest",
      "properties": { "file": "MyStage-StackA.assets.json" }
    },
    "MyStage-StackA": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/us-east-1",
      "properties": { "stackName": "my-stage-stack-a" },
      "dependencies": ["MyStage-StackA.assets"],
      "displayName": "MyStage/StackA"
    },
    "MyStage-StackB": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/ap-northeast-1",
      "properties": { "stackName": "my-stage-stack-b" },
      "dependencies": ["MyStage-StackA"],
      "displayName": "MyStage/StackB"
    }
  }
}`
	writeManifest(t, nestedDir, nestedManifest)

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}

	stackMap := make(map[string]StackInfo)
	for _, s := range stacks {
		stackMap[s.StackName] = s
	}

	stackA, ok := stackMap["my-stage-stack-a"]
	if !ok {
		t.Fatal("my-stage-stack-a not found")
	}
	if stackA.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %s", stackA.Region)
	}
	if len(stackA.Dependencies) != 0 {
		t.Errorf("expected 0 deps for StackA, got %v", stackA.Dependencies)
	}

	stackB, ok := stackMap["my-stage-stack-b"]
	if !ok {
		t.Fatal("my-stage-stack-b not found")
	}
	if stackB.Region != "ap-northeast-1" {
		t.Errorf("expected region ap-northeast-1, got %s", stackB.Region)
	}
	if len(stackB.Dependencies) != 1 || stackB.Dependencies[0] != "my-stage-stack-a" {
		t.Errorf("expected StackB to depend on my-stage-stack-a, got %v", stackB.Dependencies)
	}
}

func TestParseManifest_MultipleStages(t *testing.T) {
	dir := t.TempDir()

	topManifest := `{
  "version": "52.0.0",
  "artifacts": {
    "assembly-Stage1": {
      "type": "cdk:cloud-assembly",
      "properties": { "directoryName": "assembly-Stage1" }
    },
    "assembly-Stage2": {
      "type": "cdk:cloud-assembly",
      "properties": { "directoryName": "assembly-Stage2" }
    }
  }
}`
	writeManifest(t, dir, topManifest)

	for _, stage := range []struct {
		dir    string
		name   string
		region string
	}{
		{"assembly-Stage1", "Stage1/Stack", "us-east-1"},
		{"assembly-Stage2", "Stage2/Stack", "eu-west-1"},
	} {
		nestedDir := filepath.Join(dir, stage.dir)
		os.MkdirAll(nestedDir, 0755)
		writeManifest(t, nestedDir, fmt.Sprintf(`{
  "version": "52.0.0",
  "artifacts": {
    "%s": {
      "type": "aws:cloudformation:stack",
      "environment": "aws://123456789012/%s",
      "displayName": "%s"
    }
  }
}`, stage.name, stage.region, stage.name))
	}

	stacks, err := ParseManifest(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}
}

func TestParseEnvironment(t *testing.T) {
	tests := []struct {
		name            string
		env             string
		expectedAccount string
		expectedRegion  string
	}{
		{
			name:            "normal",
			env:             "aws://123456789012/us-east-1",
			expectedAccount: "123456789012",
			expectedRegion:  "us-east-1",
		},
		{
			name:            "unknown",
			env:             "aws://unknown-account/unknown-region",
			expectedAccount: "unknown-account",
			expectedRegion:  "unknown-region",
		},
		{
			name:            "empty",
			env:             "",
			expectedAccount: "unknown-account",
			expectedRegion:  "unknown-region",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, region := parseEnvironment(tt.env)
			if account != tt.expectedAccount {
				t.Errorf("expected account '%s', got '%s'", tt.expectedAccount, account)
			}
			if region != tt.expectedRegion {
				t.Errorf("expected region '%s', got '%s'", tt.expectedRegion, region)
			}
		})
	}
}

func writeManifest(t *testing.T, dir, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write manifest.json: %v", err)
	}
}
