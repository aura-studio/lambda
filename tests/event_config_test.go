package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aura-studio/lambda/event"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: event-lambda-handler
// Property 13: YAML Config Parsing
// *For any* valid YAML configuration containing debug, run, staticLink, or prefixLink settings,
// the corresponding Options fields SHALL be populated correctly.
//
// **Validates: Requirements 6.3, 6.4, 6.5, 6.6**

// genDebugMode generates random boolean values for debug mode
func genDebugMode() gopter.Gen {
	return gen.Bool()
}

// genRunMode generates valid run mode strings
func genRunMode() gopter.Gen {
	return gen.OneConstOf(
		event.RunModeStrict,
		event.RunModePartial,
		event.RunModeBatch,
		event.RunModeReentrant,
	)
}

// genNonEmptyPath generates non-empty path strings starting with /
func genNonEmptyPath() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}).Map(func(s string) string {
		return "/" + s
	})
}

// staticLinkEntry represents a static link configuration entry
type staticLinkEntry struct {
	SrcPath string
	DstPath string
}

// prefixLinkEntry represents a prefix link configuration entry
type prefixLinkEntry struct {
	SrcPrefix string
	DstPrefix string
}

// genStaticLinkEntry generates a valid static link entry
func genStaticLinkEntry() gopter.Gen {
	return gopter.CombineGens(
		genNonEmptyPath(),
		genNonEmptyPath(),
	).Map(func(values []interface{}) staticLinkEntry {
		return staticLinkEntry{
			SrcPath: values[0].(string),
			DstPath: values[1].(string),
		}
	})
}

// genPrefixLinkEntry generates a valid prefix link entry
func genPrefixLinkEntry() gopter.Gen {
	return gopter.CombineGens(
		genNonEmptyPath(),
		genNonEmptyPath(),
	).Map(func(values []interface{}) prefixLinkEntry {
		return prefixLinkEntry{
			SrcPrefix: values[0].(string),
			DstPrefix: values[1].(string),
		}
	})
}

// genStaticLinks generates a slice of static link entries (0 to 5 entries)
func genStaticLinks() gopter.Gen {
	return gen.SliceOfN(5, genStaticLinkEntry())
}

// genPrefixLinks generates a slice of prefix link entries (0 to 5 entries)
func genPrefixLinks() gopter.Gen {
	return gen.SliceOfN(5, genPrefixLinkEntry())
}

// buildYAMLConfig builds a YAML configuration string from the given parameters
func buildYAMLConfig(debug bool, runMode event.RunMode, staticLinks []staticLinkEntry, prefixLinks []prefixLinkEntry) []byte {
	yaml := "mode:\n"
	if debug {
		yaml += "  debug: true\n"
	} else {
		yaml += "  debug: false\n"
	}
	yaml += "  run: " + string(runMode) + "\n"

	if len(staticLinks) > 0 {
		yaml += "staticLink:\n"
		for _, link := range staticLinks {
			yaml += "  - srcPath: " + link.SrcPath + "\n"
			yaml += "    dstPath: " + link.DstPath + "\n"
		}
	}

	if len(prefixLinks) > 0 {
		yaml += "prefixLink:\n"
		for _, link := range prefixLinks {
			yaml += "  - srcPrefix: " + link.SrcPrefix + "\n"
			yaml += "    dstPrefix: " + link.DstPrefix + "\n"
		}
	}

	return []byte(yaml)
}

// TestEventYAMLConfigDebugMode tests Property 13 for debug mode setting
// Requirement 6.3: WHEN the YAML contains debug mode setting, THE Event_Engine SHALL apply it to DebugMode option
func TestEventYAMLConfigDebugMode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("debug mode setting is correctly applied", prop.ForAll(
		func(debug bool) bool {
			yaml := buildYAMLConfig(debug, event.RunModeBatch, nil, nil)
			opts := event.NewOptions(event.WithConfig(yaml))
			return opts.DebugMode == debug
		},
		genDebugMode(),
	))

	properties.TestingRun(t)
}

// TestEventYAMLConfigRunMode tests Property 13 for run mode setting
// Requirement 6.4: WHEN the YAML contains run mode setting, THE Event_Engine SHALL apply it to RunMode option
func TestEventYAMLConfigRunMode(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("run mode setting is correctly applied", prop.ForAll(
		func(runMode event.RunMode) bool {
			yaml := buildYAMLConfig(false, runMode, nil, nil)
			opts := event.NewOptions(event.WithConfig(yaml))
			return opts.RunMode == runMode
		},
		genRunMode(),
	))

	properties.TestingRun(t)
}

// TestEventYAMLConfigStaticLink tests Property 13 for staticLink entries
// Requirement 6.5: WHEN the YAML contains staticLink entries, THE Event_Engine SHALL populate StaticLinkMap
func TestEventYAMLConfigStaticLink(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("staticLink entries are correctly populated in StaticLinkMap", prop.ForAll(
		func(staticLinks []staticLinkEntry) bool {
			yaml := buildYAMLConfig(false, event.RunModeBatch, staticLinks, nil)
			opts := event.NewOptions(event.WithConfig(yaml))

			// Build expected map (handling potential duplicate keys - last one wins)
			expected := make(map[string]string)
			for _, link := range staticLinks {
				expected[link.SrcPath] = link.DstPath
			}

			// Verify all expected entries are present
			if len(opts.StaticLinkMap) != len(expected) {
				return false
			}
			for src, dst := range expected {
				if opts.StaticLinkMap[src] != dst {
					return false
				}
			}
			return true
		},
		genStaticLinks(),
	))

	properties.TestingRun(t)
}

// TestEventYAMLConfigPrefixLink tests Property 13 for prefixLink entries
// Requirement 6.6: WHEN the YAML contains prefixLink entries, THE Event_Engine SHALL populate PrefixLinkMap
func TestEventYAMLConfigPrefixLink(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("prefixLink entries are correctly populated in PrefixLinkMap", prop.ForAll(
		func(prefixLinks []prefixLinkEntry) bool {
			yaml := buildYAMLConfig(false, event.RunModeBatch, nil, prefixLinks)
			opts := event.NewOptions(event.WithConfig(yaml))

			// Build expected map (handling potential duplicate keys - last one wins)
			expected := make(map[string]string)
			for _, link := range prefixLinks {
				expected[link.SrcPrefix] = link.DstPrefix
			}

			// Verify all expected entries are present
			if len(opts.PrefixLinkMap) != len(expected) {
				return false
			}
			for src, dst := range expected {
				if opts.PrefixLinkMap[src] != dst {
					return false
				}
			}
			return true
		},
		genPrefixLinks(),
	))

	properties.TestingRun(t)
}

// TestEventYAMLConfigCombined tests Property 13 with all settings combined
// Validates: Requirements 6.3, 6.4, 6.5, 6.6
func TestEventYAMLConfigCombined(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all config settings are correctly applied together", prop.ForAll(
		func(debug bool, runMode event.RunMode, staticLinks []staticLinkEntry, prefixLinks []prefixLinkEntry) bool {
			yaml := buildYAMLConfig(debug, runMode, staticLinks, prefixLinks)
			opts := event.NewOptions(event.WithConfig(yaml))

			// Verify debug mode (Requirement 6.3)
			if opts.DebugMode != debug {
				return false
			}

			// Verify run mode (Requirement 6.4)
			if opts.RunMode != runMode {
				return false
			}

			// Verify static links (Requirement 6.5)
			expectedStatic := make(map[string]string)
			for _, link := range staticLinks {
				expectedStatic[link.SrcPath] = link.DstPath
			}
			if len(opts.StaticLinkMap) != len(expectedStatic) {
				return false
			}
			for src, dst := range expectedStatic {
				if opts.StaticLinkMap[src] != dst {
					return false
				}
			}

			// Verify prefix links (Requirement 6.6)
			expectedPrefix := make(map[string]string)
			for _, link := range prefixLinks {
				expectedPrefix[link.SrcPrefix] = link.DstPrefix
			}
			if len(opts.PrefixLinkMap) != len(expectedPrefix) {
				return false
			}
			for src, dst := range expectedPrefix {
				if opts.PrefixLinkMap[src] != dst {
					return false
				}
			}

			return true
		},
		genDebugMode(),
		genRunMode(),
		genStaticLinks(),
		genPrefixLinks(),
	))

	properties.TestingRun(t)
}

// TestEventWithConfig tests YAML configuration loading from bytes (unit test)
func TestEventWithConfig(t *testing.T) {
	yaml := []byte(`mode:
  debug: true
  run: partial
staticLink:
  - srcPath: /a
    dstPath: /b
prefixLink:
  - srcPrefix: /api
    dstPrefix: /v1
`)

	o := event.NewOptions(event.WithConfig(yaml))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if o.RunMode != event.RunModePartial {
		t.Fatalf("RunMode = %q, expected 'partial'", o.RunMode)
	}
	if got := o.StaticLinkMap["/a"]; got != "/b" {
		t.Fatalf("StaticLinkMap['/a'] = %q, expected '/b'", got)
	}
	if got := o.PrefixLinkMap["/api"]; got != "/v1" {
		t.Fatalf("PrefixLinkMap['/api'] = %q, expected '/v1'", got)
	}
}

// TestEventWithConfigAllRunModes tests all run modes in config
func TestEventWithConfigAllRunModes(t *testing.T) {
	testCases := []struct {
		name     string
		expected event.RunMode
	}{
		{"strict", event.RunModeStrict},
		{"partial", event.RunModePartial},
		{"batch", event.RunModeBatch},
		{"reentrant", event.RunModeReentrant},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			yaml := []byte("mode:\n  run: " + string(tc.expected))
			o := event.NewOptions(event.WithConfig(yaml))
			if o.RunMode != tc.expected {
				t.Errorf("RunMode = %q, expected %q", o.RunMode, tc.expected)
			}
		})
	}
}

// TestEventWithConfigDebugFalse tests YAML configuration with debug mode disabled
func TestEventWithConfigDebugFalse(t *testing.T) {
	yaml := []byte(`mode:
  debug: false
`)

	o := event.NewOptions(event.WithConfig(yaml))
	if o.DebugMode {
		t.Fatalf("DebugMode = true, expected false")
	}
}

// TestEventWithConfigEmptyYAML tests that empty YAML produces default options
func TestEventWithConfigEmptyYAML(t *testing.T) {
	yaml := []byte(``)

	o := event.NewOptions(event.WithConfig(yaml))
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

// TestEventWithConfigMultipleLinks tests YAML configuration with multiple link entries
func TestEventWithConfigMultipleLinks(t *testing.T) {
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

	o := event.NewOptions(event.WithConfig(yaml))

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

// TestEventWithConfigSkipsEmptyPaths tests that empty paths in config are skipped
func TestEventWithConfigSkipsEmptyPaths(t *testing.T) {
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

	o := event.NewOptions(event.WithConfig(yaml))

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

// TestEventWithConfigFile tests YAML configuration loading from file
func TestEventWithConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "event.yml")

	yaml := []byte(`mode:
  debug: true
  run: strict
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

	o := event.NewOptions(event.WithConfigFile(configPath))
	if !o.DebugMode {
		t.Fatalf("DebugMode = false, expected true")
	}
	if o.RunMode != event.RunModeStrict {
		t.Fatalf("RunMode = %q, expected 'strict'", o.RunMode)
	}
	if got := o.StaticLinkMap["/file-static"]; got != "/file-dest" {
		t.Fatalf("StaticLinkMap['/file-static'] = %q, expected '/file-dest'", got)
	}
	if got := o.PrefixLinkMap["/file-api"]; got != "/file-v1" {
		t.Fatalf("PrefixLinkMap['/file-api'] = %q, expected '/file-v1'", got)
	}
}

// TestEventWithConfigFileNotFound tests that WithConfigFile panics for non-existent file
func TestEventWithConfigFileNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for non-existent config file")
		}
	}()

	_ = event.NewOptions(event.WithConfigFile("/non/existent/path/event.yml"))
}

// TestEventWithConfigInvalidRunMode tests that invalid run mode causes panic
func TestEventWithConfigInvalidRunMode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid run mode")
		}
	}()

	yaml := []byte(`mode:
  run: invalid_mode
`)
	_ = event.NewOptions(event.WithConfig(yaml))
}

