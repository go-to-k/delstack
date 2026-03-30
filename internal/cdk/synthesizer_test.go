package cdk

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSynth_NoCdkJson(t *testing.T) {
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

	err = os.WriteFile(filepath.Join(tmpDir, "cdk.json"), []byte(`{"app":"echo hello"}`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	originalPath := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer t.Setenv("PATH", originalPath)

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
