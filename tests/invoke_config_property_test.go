package tests

import (
	"fmt"
	"testing"

	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: invoke-lambda-handler, Property 5: YAML 配置解析正确性**
// **Validates: Requirements 3.1, 3.6**
//
// Property 5: YAML 配置解析正确性
// For any 有效的 YAML 配置字节，解析后 SHALL 产生正确的 Options 对象，
// 其中 DebugMode、StaticLinkMap、PrefixLinkMap 字段与配置内容一致。

// staticLinkConfig represents a static link configuration entry
type staticLinkConfig struct {
	SrcPath string
	DstPath string
}

// prefixLinkConfig represents a prefix link configuration entry
type prefixLinkConfig struct {
	SrcPrefix string
	DstPrefix string
}

// yamlConfigInput represents the input data for generating YAML configuration
type yamlConfigInput struct {
	DebugMode   bool
	StaticLinks []staticLinkConfig
	PrefixLinks []prefixLinkConfig
}

// genValidPath generates valid path strings (non-empty, starting with /)
func genValidPath() gopter.Gen {
	return gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0
	}).Map(func(s string) string {
		return "/" + s
	})
}

// genStaticLink generates a valid static link configuration
func genStaticLink() gopter.Gen {
	return gopter.CombineGens(
		genValidPath(), // srcPath
		genValidPath(), // dstPath
	).Map(func(values []interface{}) staticLinkConfig {
		return staticLinkConfig{
			SrcPath: values[0].(string),
			DstPath: values[1].(string),
		}
	})
}

// genPrefixLink generates a valid prefix link configuration
func genPrefixLink() gopter.Gen {
	return gopter.CombineGens(
		genValidPath(), // srcPrefix
		genValidPath(), // dstPrefix
	).Map(func(values []interface{}) prefixLinkConfig {
		return prefixLinkConfig{
			SrcPrefix: values[0].(string),
			DstPrefix: values[1].(string),
		}
	})
}

// genYAMLConfigInput generates random valid YAML configuration inputs
func genYAMLConfigInput() gopter.Gen {
	return gopter.CombineGens(
		gen.Bool(),                                           // debugMode
		gen.SliceOfN(3, genStaticLink()),                     // staticLinks (0-3 entries)
		gen.SliceOfN(3, genPrefixLink()),                     // prefixLinks (0-3 entries)
	).Map(func(values []interface{}) yamlConfigInput {
		return yamlConfigInput{
			DebugMode:   values[0].(bool),
			StaticLinks: values[1].([]staticLinkConfig),
			PrefixLinks: values[2].([]prefixLinkConfig),
		}
	})
}

// buildYAMLBytes constructs YAML bytes from the configuration input
func buildYAMLBytes(input yamlConfigInput) []byte {
	yaml := fmt.Sprintf("mode:\n  debug: %t\n", input.DebugMode)

	if len(input.StaticLinks) > 0 {
		yaml += "staticLink:\n"
		for _, link := range input.StaticLinks {
			yaml += fmt.Sprintf("  - srcPath: \"%s\"\n    dstPath: \"%s\"\n", link.SrcPath, link.DstPath)
		}
	}

	if len(input.PrefixLinks) > 0 {
		yaml += "prefixLink:\n"
		for _, link := range input.PrefixLinks {
			yaml += fmt.Sprintf("  - srcPrefix: \"%s\"\n    dstPrefix: \"%s\"\n", link.SrcPrefix, link.DstPrefix)
		}
	}

	return []byte(yaml)
}

// TestYAMLConfigParsing tests that valid YAML configurations are parsed correctly
// into Options objects with matching DebugMode, StaticLinkMap, and PrefixLinkMap fields.
// **Validates: Requirements 3.1, 3.6**
func TestYAMLConfigParsing(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("YAML config parsing: parsed Options matches input configuration", prop.ForAll(
		func(input yamlConfigInput) bool {
			// Build YAML bytes from input
			yamlBytes := buildYAMLBytes(input)

			// Parse using WithConfig and create Options
			opts := invoke.NewOptions(invoke.WithConfig(yamlBytes))

			// Verify DebugMode
			if opts.DebugMode != input.DebugMode {
				t.Logf("DebugMode mismatch: expected %t, got %t", input.DebugMode, opts.DebugMode)
				return false
			}

			// Verify StaticLinkMap
			expectedStaticLinks := make(map[string]string)
			for _, link := range input.StaticLinks {
				expectedStaticLinks[link.SrcPath] = link.DstPath
			}
			if len(opts.StaticLinkMap) != len(expectedStaticLinks) {
				t.Logf("StaticLinkMap length mismatch: expected %d, got %d",
					len(expectedStaticLinks), len(opts.StaticLinkMap))
				return false
			}
			for srcPath, expectedDstPath := range expectedStaticLinks {
				if actualDstPath, ok := opts.StaticLinkMap[srcPath]; !ok || actualDstPath != expectedDstPath {
					t.Logf("StaticLinkMap mismatch for %s: expected %s, got %s",
						srcPath, expectedDstPath, actualDstPath)
					return false
				}
			}

			// Verify PrefixLinkMap
			expectedPrefixLinks := make(map[string]string)
			for _, link := range input.PrefixLinks {
				expectedPrefixLinks[link.SrcPrefix] = link.DstPrefix
			}
			if len(opts.PrefixLinkMap) != len(expectedPrefixLinks) {
				t.Logf("PrefixLinkMap length mismatch: expected %d, got %d",
					len(expectedPrefixLinks), len(opts.PrefixLinkMap))
				return false
			}
			for srcPrefix, expectedDstPrefix := range expectedPrefixLinks {
				if actualDstPrefix, ok := opts.PrefixLinkMap[srcPrefix]; !ok || actualDstPrefix != expectedDstPrefix {
					t.Logf("PrefixLinkMap mismatch for %s: expected %s, got %s",
						srcPrefix, expectedDstPrefix, actualDstPrefix)
					return false
				}
			}

			return true
		},
		genYAMLConfigInput(),
	))

	properties.TestingRun(t)
}

// TestYAMLConfigParsingEmptyConfig tests that empty YAML configuration produces default Options
// **Validates: Requirements 3.1**
func TestYAMLConfigParsingEmptyConfig(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Empty YAML config: produces default Options with empty maps and false debug mode", prop.ForAll(
		func(_ bool) bool {
			// Empty YAML configuration
			yamlBytes := []byte("")

			// Parse using WithConfig and create Options
			opts := invoke.NewOptions(invoke.WithConfig(yamlBytes))

			// Verify defaults
			if opts.DebugMode != false {
				t.Logf("DebugMode should be false for empty config, got %t", opts.DebugMode)
				return false
			}
			if len(opts.StaticLinkMap) != 0 {
				t.Logf("StaticLinkMap should be empty for empty config, got %d entries", len(opts.StaticLinkMap))
				return false
			}
			if len(opts.PrefixLinkMap) != 0 {
				t.Logf("PrefixLinkMap should be empty for empty config, got %d entries", len(opts.PrefixLinkMap))
				return false
			}

			return true
		},
		gen.Bool(), // Dummy generator to run the test
	))

	properties.TestingRun(t)
}

// TestYAMLConfigParsingDebugModeOnly tests that debug mode only configuration is parsed correctly
// **Validates: Requirements 3.1**
func TestYAMLConfigParsingDebugModeOnly(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("Debug mode only config: DebugMode matches input, maps are empty", prop.ForAll(
		func(debugMode bool) bool {
			// Build YAML with only debug mode
			yamlBytes := []byte(fmt.Sprintf("mode:\n  debug: %t\n", debugMode))

			// Parse using WithConfig and create Options
			opts := invoke.NewOptions(invoke.WithConfig(yamlBytes))

			// Verify DebugMode
			if opts.DebugMode != debugMode {
				t.Logf("DebugMode mismatch: expected %t, got %t", debugMode, opts.DebugMode)
				return false
			}

			// Verify maps are empty
			if len(opts.StaticLinkMap) != 0 {
				t.Logf("StaticLinkMap should be empty, got %d entries", len(opts.StaticLinkMap))
				return false
			}
			if len(opts.PrefixLinkMap) != 0 {
				t.Logf("PrefixLinkMap should be empty, got %d entries", len(opts.PrefixLinkMap))
				return false
			}

			return true
		},
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// TestYAMLConfigParsingInvalidYAML tests that invalid YAML configuration causes panic
// **Validates: Requirements 3.6**
func TestYAMLConfigParsingInvalidYAML(t *testing.T) {
	// Test cases with invalid YAML
	// Note: YAML parser is lenient with some formats, so we test cases that definitely fail
	invalidYAMLCases := []struct {
		name string
		yaml string
	}{
		{"invalid boolean", "mode:\n  debug: notabool"},
		{"unclosed quote", "mode:\n  debug: \"true"},
		{"invalid structure", "mode: [debug: true]"},
		{"malformed yaml", "mode:\n  debug: true\n  invalid: ["},
	}

	for _, tc := range invalidYAMLCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Expected panic for invalid YAML: %s", tc.name)
				}
			}()

			// This should panic for invalid YAML
			_ = invoke.NewOptions(invoke.WithConfig([]byte(tc.yaml)))
		})
	}
}
