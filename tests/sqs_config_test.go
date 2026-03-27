package tests

import (
	"os"
	"path/filepath"
	"testing"

	lambdasqs "github.com/aura-studio/lambda/sqs"
)

// TestSQSWithConfig tests YAML configuration loading from bytes
func TestSQSWithConfig(t *testing.T) {
	yaml := []byte(`mode:
  debug: true
  run: partial
  reply: true
`)

	o := lambdasqs.NewOptions(lambdasqs.WithConfig(yaml))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if o.RunMode != lambdasqs.RunModePartial {
		t.Fatalf("RunMode = %q, expected 'partial'", o.RunMode)
	}
	if !o.ReplyMode {
		t.Fatalf("ReplyMode = false, expected true")
	}
}

// TestSQSWithConfigAllRunModes tests all run modes in config
func TestSQSWithConfigAllRunModes(t *testing.T) {
	testCases := []struct {
		name     string
		yaml     string
		expected lambdasqs.RunMode
	}{
		{"strict", `mode:\n  run: strict`, lambdasqs.RunModeStrict},
		{"partial", `mode:\n  run: partial`, lambdasqs.RunModePartial},
		{"batch", `mode:\n  run: batch`, lambdasqs.RunModeBatch},
		{"reentrant", `mode:\n  run: reentrant`, lambdasqs.RunModeReentrant},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			yaml := []byte("mode:\n  run: " + string(tc.expected))
			o := lambdasqs.NewOptions(lambdasqs.WithConfig(yaml))
			if o.RunMode != tc.expected {
				t.Errorf("RunMode = %q, expected %q", o.RunMode, tc.expected)
			}
		})
	}
}

// TestSQSWithConfigDebugFalse tests YAML configuration with debug mode disabled
func TestSQSWithConfigDebugFalse(t *testing.T) {
	yaml := []byte(`mode:
  debug: false
`)

	o := lambdasqs.NewOptions(lambdasqs.WithConfig(yaml))
	if o.DebugMode {
		t.Fatalf("DebugMode = true, expected false")
	}
}

// TestSQSWithConfigEmptyYAML tests that empty YAML produces default options
func TestSQSWithConfigEmptyYAML(t *testing.T) {
	yaml := []byte(``)

	o := lambdasqs.NewOptions(lambdasqs.WithConfig(yaml))
	if o.DebugMode {
		t.Fatalf("DebugMode = true, expected false for empty config")
	}
	if o.ReplyMode {
		t.Fatalf("ReplyMode = true, expected false for empty config")
	}
}



// TestSQSWithConfigFile tests YAML configuration loading from file
func TestSQSWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sqs.yml")

	yaml := []byte(`mode:
  debug: true
  run: strict
  reply: true
`)

	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	o := lambdasqs.NewOptions(lambdasqs.WithConfigFile(configPath))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if o.RunMode != lambdasqs.RunModeStrict {
		t.Fatalf("RunMode = %q, expected 'strict'", o.RunMode)
	}
	if !o.ReplyMode {
		t.Fatalf("ReplyMode = false, expected true")
	}
}

// TestSQSWithConfigFileNotFound tests that WithConfigFile panics for non-existent file
func TestSQSWithConfigFileNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for non-existent config file")
		}
	}()

	_ = lambdasqs.NewOptions(lambdasqs.WithConfigFile("/non/existent/path/sqs.yml"))
}

// TestSQSDefaultConfigCandidates tests that DefaultConfigCandidates returns expected paths
func TestSQSDefaultConfigCandidates(t *testing.T) {
	candidates := lambdasqs.DefaultConfigCandidates()

	if len(candidates) != 4 {
		t.Fatalf("Expected 4 candidates, got %d", len(candidates))
	}

	// Check expected candidates (order matters)
	expected := []string{
		"sqs.yaml",
		"sqs.yml",
		filepath.FromSlash("sqs/sqs.yaml"),
		filepath.FromSlash("sqs/sqs.yml"),
	}

	for i, exp := range expected {
		if candidates[i] != exp {
			t.Fatalf("Candidate[%d] = %q, expected %q", i, candidates[i], exp)
		}
	}
}

// TestSQSFindDefaultConfigFileFound tests that FindDefaultConfigFile finds existing config
func TestSQSFindDefaultConfigFileFound(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sqs.yml")

	yaml := []byte(`mode:
  debug: true
`)
	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// FindDefaultConfigFile should find the config
	foundPath, err := lambdasqs.FindDefaultConfigFile()
	if err != nil {
		t.Fatalf("FindDefaultConfigFile failed: %v", err)
	}

	if foundPath != "sqs.yml" {
		t.Fatalf("FindDefaultConfigFile returned %q, expected 'sqs.yml'", foundPath)
	}
}

// TestSQSFindDefaultConfigFileNotFound tests that FindDefaultConfigFile returns error when no config exists
func TestSQSFindDefaultConfigFileNotFound(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create an empty temporary directory
	tmpDir := t.TempDir()

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// FindDefaultConfigFile should return an error
	_, err = lambdasqs.FindDefaultConfigFile()
	if err == nil {
		t.Fatalf("FindDefaultConfigFile should return error when no config exists")
	}
}

// TestSQSFindDefaultConfigFileInSubdirectory tests that FindDefaultConfigFile finds config in sqs/ subdirectory
func TestSQSFindDefaultConfigFileInSubdirectory(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with sqs/ subdirectory
	tmpDir := t.TempDir()
	sqsDir := filepath.Join(tmpDir, "sqs")
	if err := os.MkdirAll(sqsDir, 0755); err != nil {
		t.Fatalf("Failed to create sqs directory: %v", err)
	}

	configPath := filepath.Join(sqsDir, "sqs.yaml")
	yaml := []byte(`mode:
  debug: true
`)
	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// FindDefaultConfigFile should find the config in subdirectory
	foundPath, err := lambdasqs.FindDefaultConfigFile()
	if err != nil {
		t.Fatalf("FindDefaultConfigFile failed: %v", err)
	}

	expected := filepath.FromSlash("sqs/sqs.yaml")
	if foundPath != expected {
		t.Fatalf("FindDefaultConfigFile returned %q, expected %q", foundPath, expected)
	}
}

// TestSQSWithDefaultConfigFilePanicsWhenNotFound tests that WithDefaultConfigFile panics when no config exists
func TestSQSWithDefaultConfigFilePanicsWhenNotFound(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create an empty temporary directory
	tmpDir := t.TempDir()

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when no default config file exists")
		}
	}()

	_ = lambdasqs.NewOptions(lambdasqs.WithDefaultConfigFile())
}

// TestSQSWithDefaultConfigFileLoadsConfig tests that WithDefaultConfigFile loads config correctly
func TestSQSWithDefaultConfigFileLoadsConfig(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "sqs.yml")

	yaml := []byte(`mode:
  debug: true
  run: reentrant
`)
	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	o := lambdasqs.NewOptions(lambdasqs.WithDefaultConfigFile())
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if o.RunMode != lambdasqs.RunModeReentrant {
		t.Fatalf("RunMode = %q, expected 'reentrant'", o.RunMode)
	}
}
