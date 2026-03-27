package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aura-studio/lambda/reqresp"
)

// TestReqRespWithConfig tests YAML configuration loading from bytes
func TestReqRespWithConfig(t *testing.T) {
	yaml := []byte(`mode:
  debug: true
`)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
}

// TestReqRespWithConfigDebugFalse tests YAML configuration with debug mode disabled
func TestReqRespWithConfigDebugFalse(t *testing.T) {
	yaml := []byte(`mode:
  debug: false
`)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))
	if o.DebugMode {
		t.Fatalf("DebugMode = true, expected false")
	}
}

// TestReqRespWithConfigEmptyYAML tests that empty YAML produces default options
func TestReqRespWithConfigEmptyYAML(t *testing.T) {
	yaml := []byte(``)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))
	if o.DebugMode {
		t.Fatalf("DebugMode = true, expected false for empty config")
	}
}

// TestReqRespWithConfigFile tests YAML configuration loading from file
func TestReqRespWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "reqresp.yml")

	yaml := []byte(`mode:
  debug: true
`)

	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	o := reqresp.NewOptions(reqresp.WithConfigFile(configPath))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
}

// TestReqRespWithConfigFileNotFound tests that WithConfigFile panics for non-existent file
func TestReqRespWithConfigFileNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for non-existent config file")
		}
	}()

	_ = reqresp.NewOptions(reqresp.WithConfigFile("/non/existent/path/reqresp.yml"))
}

// TestDefaultConfigCandidates tests that DefaultConfigCandidates returns expected paths
func TestDefaultConfigCandidates(t *testing.T) {
	candidates := reqresp.DefaultConfigCandidates()

	if len(candidates) != 4 {
		t.Fatalf("Expected 4 candidates, got %d", len(candidates))
	}

	expected := []string{
		"reqresp.yaml",
		"reqresp.yml",
		filepath.FromSlash("reqresp/reqresp.yaml"),
		filepath.FromSlash("reqresp/reqresp.yml"),
	}

	for i, exp := range expected {
		if candidates[i] != exp {
			t.Fatalf("Candidate[%d] = %q, expected %q", i, candidates[i], exp)
		}
	}
}

// TestFindDefaultConfigFileFound tests that FindDefaultConfigFile finds existing config
func TestFindDefaultConfigFileFound(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "reqresp.yml")

	yaml := []byte(`mode:
  debug: true
`)
	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	foundPath, err := reqresp.FindDefaultConfigFile()
	if err != nil {
		t.Fatalf("FindDefaultConfigFile failed: %v", err)
	}

	if foundPath != "reqresp.yml" {
		t.Fatalf("FindDefaultConfigFile returned %q, expected 'reqresp.yml'", foundPath)
	}
}

// TestFindDefaultConfigFileNotFound tests that FindDefaultConfigFile returns error when no config exists
func TestFindDefaultConfigFileNotFound(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	_, err = reqresp.FindDefaultConfigFile()
	if err == nil {
		t.Fatalf("FindDefaultConfigFile should return error when no config exists")
	}
}
