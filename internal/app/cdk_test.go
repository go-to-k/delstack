package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-to-k/delstack/internal/io"
)

func TestCdkAction_Validation(t *testing.T) {
	io.NewLogger(false)

	tests := []struct {
		name    string
		action  *CdkAction
		wantErr string
	}{
		{
			name:    "stack names with interactive mode",
			action:  NewCdkAction([]string{"Stack1"}, "", "", true, false, true, 0, "./cdk.out", nil),
			wantErr: "InvalidOptionError",
		},
		{
			name:    "negative concurrency number",
			action:  NewCdkAction(nil, "", "", false, false, true, -1, "./cdk.out", nil),
			wantErr: "InvalidOptionError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Run(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestCdkAction_NoCdkJson(t *testing.T) {
	io.NewLogger(false)

	// Change to temp dir without cdk.json
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(original) }()

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	action := NewCdkAction(nil, "", "", false, false, true, 0, "", nil)
	err = action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsString(err.Error(), "cdk.json not found") {
		t.Errorf("expected cdk.json error, got %q", err.Error())
	}
}

func TestCdkAction_AppPathNoManifest(t *testing.T) {
	io.NewLogger(false)

	tmpDir := t.TempDir()

	action := NewCdkAction(nil, "", "", false, false, true, 0, tmpDir, nil)
	err := action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsString(err.Error(), "manifest.json") {
		t.Errorf("expected manifest.json error, got %q", err.Error())
	}
}

func TestCdkAction_EmptyManifest(t *testing.T) {
	io.NewLogger(false)

	tmpDir := t.TempDir()
	manifest := `{"version": "52.0.0", "artifacts": {"Tree": {"type": "cdk:tree", "properties": {"file": "tree.json"}}}}`
	err := os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte(manifest), 0600)
	if err != nil {
		t.Fatal(err)
	}

	action := NewCdkAction(nil, "", "", false, false, true, 0, tmpDir, nil)
	err = action.Run(context.Background())
	// No error — just logs "No stacks found" and returns nil
	if err != nil {
		t.Errorf("expected nil error for empty manifest, got %q", err.Error())
	}
}

func TestCdkAction_StackNameNotInManifest(t *testing.T) {
	io.NewLogger(false)

	tmpDir := t.TempDir()
	manifest := `{
		"version": "52.0.0",
		"artifacts": {
			"MyStack": {
				"type": "aws:cloudformation:stack",
				"environment": "aws://123456789012/us-east-1",
				"displayName": "MyStack"
			}
		}
	}`
	err := os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte(manifest), 0600)
	if err != nil {
		t.Fatal(err)
	}

	action := NewCdkAction([]string{"NonExistentStack"}, "", "us-east-1", false, false, true, 0, tmpDir, nil)
	err = action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !containsString(err.Error(), "stacks not found in CDK app") {
		t.Errorf("expected 'stacks not found' error, got %q", err.Error())
	}
}

func TestCdkAction_AppPathDirectory(t *testing.T) {
	io.NewLogger(false)

	// -a with a directory path should read manifest directly (no synth)
	tmpDir := t.TempDir()
	manifest := `{"version": "52.0.0", "artifacts": {"Tree": {"type": "cdk:tree", "properties": {"file": "tree.json"}}}}`
	err := os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte(manifest), 0600)
	if err != nil {
		t.Fatal(err)
	}

	action := NewCdkAction(nil, "", "", false, false, true, 0, tmpDir, nil)
	err = action.Run(context.Background())
	// No stacks in manifest, should return nil (no error, just "No stacks found")
	if err != nil {
		t.Errorf("expected nil error for directory appPath, got %q", err.Error())
	}
}

func TestCdkAction_AppPathCommand(t *testing.T) {
	io.NewLogger(false)

	// -a with a non-directory string should be treated as an app command
	// This will fail because "echo hello" won't produce a valid cdk.out,
	// but it verifies the command path is taken (not the directory path)
	action := NewCdkAction(nil, "", "", false, false, true, 0, "echo hello", nil)
	err := action.Run(context.Background())
	if err == nil {
		t.Fatal("expected error for command appPath (no valid cdk.out produced)")
	}
	// Should fail at cdk synth or manifest parsing, not at "cdk.json not found"
	if containsString(err.Error(), "cdk.json not found") {
		t.Errorf("should not check cdk.json when appPath is a command, got %q", err.Error())
	}
}

func TestIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		appPath string
		want    bool
	}{
		{"directory", tmpDir, true},
		{"file", tmpFile, false},
		{"nonexistent path", "/nonexistent/path", false},
		{"command string", "npx ts-node bin/app.ts", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &CdkAction{appPath: tt.appPath}
			if got := a.isDirectory(); got != tt.want {
				t.Errorf("isDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}
