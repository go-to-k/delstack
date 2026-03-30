package cdk

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSynth_NoCdkJson(t *testing.T) {
	// Change to temp dir without cdk.json
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	s := NewSynthesizer()
	err = s.Synth(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when cdk.json is missing")
	}
	if err.Error() != "cdk.json not found in current directory" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSynth_CdkNotInstalled(t *testing.T) {
	// Change to temp dir with cdk.json but no cdk CLI
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(original)

	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// Create a cdk.json
	err = os.WriteFile(filepath.Join(tmpDir, "cdk.json"), []byte(`{"app":"echo hello"}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Set PATH to empty to ensure cdk command is not found
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	s := NewSynthesizer()
	err = s.Synth(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when cdk is not installed")
	}
}

func TestNewSynthesizer(t *testing.T) {
	s := NewSynthesizer()
	if s.CdkOutDir != DefaultCdkOutDir {
		t.Errorf("expected CdkOutDir '%s', got '%s'", DefaultCdkOutDir, s.CdkOutDir)
	}
}
