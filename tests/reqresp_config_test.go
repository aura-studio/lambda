package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aura-studio/lambda/reqresp"
)

// TestReqRespWithConfig tests YAML configuration loading from bytes
// **Validates: Requirements 3.1**
func TestReqRespWithConfig(t *testing.T) {
	yaml := []byte(`mode:
  debug: true
staticLink:
  - srcPath: /a
    dstPath: /b
prefixLink:
  - srcPrefix: /api
    dstPrefix: /v1
`)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if got := o.StaticLinkMap["/a"]; got != "/b" {
		t.Fatalf("StaticLinkMap['/a'] = %q, expected '/b'", got)
	}
	if got := o.PrefixLinkMap["/api"]; got != "/v1" {
		t.Fatalf("PrefixLinkMap['/api'] = %q, expected '/v1'", got)
	}
}

// TestReqRespWithConfigDebugFalse tests YAML configuration with debug mode disabled
// **Validates: Requirements 3.1**
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
// **Validates: Requirements 3.1**
func TestReqRespWithConfigEmptyYAML(t *testing.T) {
	yaml := []byte(``)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))
	if o.DebugMode {
		t.Fatalf("DebugMode = true, expected false for empty config")
	}
	if len(o.StaticLinkMap) != 0 {
		t.Fatalf("StaticLinkMap should be empty, got %d entries", len(o.StaticLinkMap))
	}
	if len(o.PrefixLinkMap) != 0 {
		t.Fatalf("PrefixLinkMap should be empty, got %d entries", len(o.PrefixLinkMap))
	}
}


// TestReqRespWithConfigMultipleLinks tests YAML configuration with multiple link entries
// **Validates: Requirements 3.1**
func TestReqRespWithConfigMultipleLinks(t *testing.T) {
	yaml := []byte(`mode:
  debug: true
staticLink:
  - srcPath: /static1
    dstPath: /dest1
  - srcPath: /static2
    dstPath: /dest2
  - srcPath: /static3
    dstPath: /dest3
prefixLink:
  - srcPrefix: /api
    dstPrefix: /v1
  - srcPrefix: /admin
    dstPrefix: /v2
`)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))

	// Verify all static links
	expectedStatic := map[string]string{
		"/static1": "/dest1",
		"/static2": "/dest2",
		"/static3": "/dest3",
	}
	for src, expectedDst := range expectedStatic {
		if got := o.StaticLinkMap[src]; got != expectedDst {
			t.Fatalf("StaticLinkMap['%s'] = %q, expected %q", src, got, expectedDst)
		}
	}

	// Verify all prefix links
	expectedPrefix := map[string]string{
		"/api":   "/v1",
		"/admin": "/v2",
	}
	for src, expectedDst := range expectedPrefix {
		if got := o.PrefixLinkMap[src]; got != expectedDst {
			t.Fatalf("PrefixLinkMap['%s'] = %q, expected %q", src, got, expectedDst)
		}
	}
}

// TestReqRespWithConfigSkipsEmptyPaths tests that empty paths in config are skipped
// **Validates: Requirements 3.1**
func TestReqRespWithConfigSkipsEmptyPaths(t *testing.T) {
	yaml := []byte(`staticLink:
  - srcPath: ""
    dstPath: /dest
  - srcPath: /src
    dstPath: ""
  - srcPath: /valid
    dstPath: /valid-dest
prefixLink:
  - srcPrefix: ""
    dstPrefix: /dest
  - srcPrefix: /src
    dstPrefix: ""
  - srcPrefix: /valid-prefix
    dstPrefix: /valid-prefix-dest
`)

	o := reqresp.NewOptions(reqresp.WithConfig(yaml))

	// Only valid entries should be present
	if len(o.StaticLinkMap) != 1 {
		t.Fatalf("StaticLinkMap should have 1 entry, got %d", len(o.StaticLinkMap))
	}
	if got := o.StaticLinkMap["/valid"]; got != "/valid-dest" {
		t.Fatalf("StaticLinkMap['/valid'] = %q, expected '/valid-dest'", got)
	}

	if len(o.PrefixLinkMap) != 1 {
		t.Fatalf("PrefixLinkMap should have 1 entry, got %d", len(o.PrefixLinkMap))
	}
	if got := o.PrefixLinkMap["/valid-prefix"]; got != "/valid-prefix-dest" {
		t.Fatalf("PrefixLinkMap['/valid-prefix'] = %q, expected '/valid-prefix-dest'", got)
	}
}

// TestReqRespWithConfigFile tests YAML configuration loading from file
// **Validates: Requirements 3.1**
func TestReqRespWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "reqresp.yml")

	yaml := []byte(`mode:
  debug: true
staticLink:
  - srcPath: /file-static
    dstPath: /file-dest
prefixLink:
  - srcPrefix: /file-api
    dstPrefix: /file-v1
`)

	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	o := reqresp.NewOptions(reqresp.WithConfigFile(configPath))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if got := o.StaticLinkMap["/file-static"]; got != "/file-dest" {
		t.Fatalf("StaticLinkMap['/file-static'] = %q, expected '/file-dest'", got)
	}
	if got := o.PrefixLinkMap["/file-api"]; got != "/file-v1" {
		t.Fatalf("PrefixLinkMap['/file-api'] = %q, expected '/file-v1'", got)
	}
}

// TestReqRespWithConfigFileNotFound tests that WithConfigFile panics for non-existent file
// **Validates: Requirements 3.1**
func TestReqRespWithConfigFileNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for non-existent config file")
		}
	}()

	_ = reqresp.NewOptions(reqresp.WithConfigFile("/non/existent/path/reqresp.yml"))
}

// TestDefaultConfigCandidates tests that DefaultConfigCandidates returns expected paths
// **Validates: Requirements 3.5**
func TestDefaultConfigCandidates(t *testing.T) {
	candidates := reqresp.DefaultConfigCandidates()

	if len(candidates) != 4 {
		t.Fatalf("Expected 4 candidates, got %d", len(candidates))
	}

	// Check expected candidates (order matters)
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
// **Validates: Requirements 3.5**
func TestFindDefaultConfigFileFound(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "reqresp.yml")

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
	foundPath, err := reqresp.FindDefaultConfigFile()
	if err != nil {
		t.Fatalf("FindDefaultConfigFile failed: %v", err)
	}

	if foundPath != "reqresp.yml" {
		t.Fatalf("FindDefaultConfigFile returned %q, expected 'reqresp.yml'", foundPath)
	}
}

// TestFindDefaultConfigFileNotFound tests that FindDefaultConfigFile returns error when no config exists
// **Validates: Requirements 3.5**
func TestFindDefaultConfigFileNotFound(t *testing.T) {
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
	_, err = reqresp.FindDefaultConfigFile()
	if err == nil {
		t.Fatalf("FindDefaultConfigFile should return error when no config exists")
	}
}

// TestFindDefaultConfigFileInSubdirectory tests that FindDefaultConfigFile finds config in reqresp/ subdirectory
// **Validates: Requirements 3.5**
func TestFindDefaultConfigFileInSubdirectory(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with reqresp/ subdirectory
	tmpDir := t.TempDir()
	reqrespDir := filepath.Join(tmpDir, "reqresp")
	if err := os.MkdirAll(reqrespDir, 0755); err != nil {
		t.Fatalf("Failed to create reqresp directory: %v", err)
	}

	configPath := filepath.Join(reqrespDir, "reqresp.yaml")
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
	foundPath, err := reqresp.FindDefaultConfigFile()
	if err != nil {
		t.Fatalf("FindDefaultConfigFile failed: %v", err)
	}

	expected := filepath.FromSlash("reqresp/reqresp.yaml")
	if foundPath != expected {
		t.Fatalf("FindDefaultConfigFile returned %q, expected %q", foundPath, expected)
	}
}

// TestFindDefaultConfigFilePriority tests that FindDefaultConfigFile respects priority order
// **Validates: Requirements 3.5**
func TestFindDefaultConfigFilePriority(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with multiple config files
	tmpDir := t.TempDir()

	// Create reqresp/reqresp.yml (lower priority)
	reqrespDir := filepath.Join(tmpDir, "reqresp")
	if err := os.MkdirAll(reqrespDir, 0755); err != nil {
		t.Fatalf("Failed to create reqresp directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reqrespDir, "reqresp.yml"), []byte(`mode:\n  debug: false`), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create reqresp.yaml (higher priority)
	if err := os.WriteFile(filepath.Join(tmpDir, "reqresp.yaml"), []byte(`mode:\n  debug: true`), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// FindDefaultConfigFile should find reqresp.yaml first (higher priority)
	foundPath, err := reqresp.FindDefaultConfigFile()
	if err != nil {
		t.Fatalf("FindDefaultConfigFile failed: %v", err)
	}

	if foundPath != "reqresp.yaml" {
		t.Fatalf("FindDefaultConfigFile returned %q, expected 'reqresp.yaml' (higher priority)", foundPath)
	}
}

// TestWithDefaultConfigFilePanicsWhenNotFound tests that WithDefaultConfigFile panics when no config exists
// **Validates: Requirements 3.5**
func TestWithDefaultConfigFilePanicsWhenNotFound(t *testing.T) {
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

	_ = reqresp.NewOptions(reqresp.WithDefaultConfigFile())
}

// TestWithDefaultConfigFileLoadsConfig tests that WithDefaultConfigFile loads config correctly
// **Validates: Requirements 3.5**
func TestWithDefaultConfigFileLoadsConfig(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a temporary directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "reqresp.yml")

	yaml := []byte(`mode:
  debug: true
staticLink:
  - srcPath: /default-static
    dstPath: /default-dest
`)
	if err := os.WriteFile(configPath, yaml, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to the temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	o := reqresp.NewOptions(reqresp.WithDefaultConfigFile())
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if got := o.StaticLinkMap["/default-static"]; got != "/default-dest" {
		t.Fatalf("StaticLinkMap['/default-static'] = %q, expected '/default-dest'", got)
	}
}
