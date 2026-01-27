package tests

import (
	"strings"
	"testing"

	"github.com/aura-studio/lambda/invoke"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: invoke-lambda-handler, Property 3: 路径映射正确性**
// **Validates: Requirements 2.5, 2.6**
//
// Property 3: 路径映射正确性
// For any 配置的 StaticLink 或 PrefixLink 映射，当请求路径匹配源路径/前缀时，
// 路由器 SHALL 将路径转换为目标路径/前缀。

// genPathMappingSegment generates a valid path segment for mapping tests
func genPathMappingSegment() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		// Limit length to avoid overly long paths
		if len(s) > 12 {
			return s[:12]
		}
		return s
	})
}

// genStaticPath generates a simple path like "/foo" for static link testing
func genStaticPath() gopter.Gen {
	return genPathMappingSegment().Map(func(segment string) string {
		return "/" + segment
	})
}

// genPrefixPath generates a prefix path like "/v1/api" for prefix link testing
func genPrefixPath() gopter.Gen {
	return gopter.CombineGens(
		genPathMappingSegment(),
		genPathMappingSegment(),
	).Map(func(values []interface{}) string {
		return "/" + values[0].(string) + "/" + values[1].(string)
	})
}

// TestStaticLinkExactMapping tests that StaticLink maps exact paths correctly
// **Validates: Requirements 2.5**
func TestStaticLinkExactMapping(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("StaticLink: exact path is mapped to destination path", prop.ForAll(
		func(srcPath, dstPath string) bool {
			// Create engine with static link mapping
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithStaticLink(srcPath, dstPath),
				},
				nil,
			)

			// Create context with source path
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: srcPath,
				Path:    srcPath,
			}

			// Apply StaticLink middleware
			engine.StaticLink(ctx)

			// Path should be transformed to destination
			if ctx.Path != dstPath {
				t.Logf("StaticLink mapping failed: srcPath=%q, expected dstPath=%q, got=%q",
					srcPath, dstPath, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
	))

	properties.TestingRun(t)
}

// TestStaticLinkNonMatchingPath tests that StaticLink does not modify non-matching paths
// **Validates: Requirements 2.5**
func TestStaticLinkNonMatchingPath(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("StaticLink: non-matching path is not modified", prop.ForAll(
		func(srcPath, dstPath, requestPath string) bool {
			// Skip if request path happens to match source path
			if srcPath == requestPath {
				return true
			}

			// Create engine with static link mapping
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithStaticLink(srcPath, dstPath),
				},
				nil,
			)

			// Create context with non-matching path
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath,
				Path:    requestPath,
			}

			// Apply StaticLink middleware
			engine.StaticLink(ctx)

			// Path should remain unchanged
			if ctx.Path != requestPath {
				t.Logf("StaticLink should not modify non-matching path: srcPath=%q, requestPath=%q, got=%q",
					srcPath, requestPath, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
		genStaticPath(),
	))

	properties.TestingRun(t)
}

// TestStaticLinkMultipleMappings tests that multiple StaticLink mappings work correctly
// **Validates: Requirements 2.5**
func TestStaticLinkMultipleMappings(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("StaticLink: multiple mappings work correctly", prop.ForAll(
		func(src1, dst1, src2, dst2 string) bool {
			// Skip if sources are the same (would overwrite)
			if src1 == src2 {
				return true
			}

			// Create engine with multiple static link mappings
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithStaticLinkMap(map[string]string{
						src1: dst1,
						src2: dst2,
					}),
				},
				nil,
			)

			// Test first mapping
			ctx1 := &invoke.Context{
				Engine:  engine,
				RawPath: src1,
				Path:    src1,
			}
			engine.StaticLink(ctx1)
			if ctx1.Path != dst1 {
				t.Logf("First mapping failed: src=%q, expected=%q, got=%q", src1, dst1, ctx1.Path)
				return false
			}

			// Test second mapping
			ctx2 := &invoke.Context{
				Engine:  engine,
				RawPath: src2,
				Path:    src2,
			}
			engine.StaticLink(ctx2)
			if ctx2.Path != dst2 {
				t.Logf("Second mapping failed: src=%q, expected=%q, got=%q", src2, dst2, ctx2.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
		genStaticPath(),
		genStaticPath(),
	))

	properties.TestingRun(t)
}

// TestPrefixLinkPrefixMapping tests that PrefixLink maps paths with matching prefix correctly
// **Validates: Requirements 2.6**
func TestPrefixLinkPrefixMapping(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLink: path with matching prefix is transformed", prop.ForAll(
		func(srcPrefix, dstPrefix, suffix string) bool {
			// Create engine with prefix link mapping
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithPrefixLink(srcPrefix, dstPrefix),
				},
				nil,
			)

			// Create request path with source prefix
			requestPath := srcPrefix + "/" + suffix

			// Create context
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath,
				Path:    requestPath,
			}

			// Apply PrefixLink middleware
			engine.PrefixLink(ctx)

			// Path should have prefix replaced
			expectedPath := dstPrefix + "/" + suffix
			if ctx.Path != expectedPath {
				t.Logf("PrefixLink mapping failed: srcPrefix=%q, dstPrefix=%q, suffix=%q, expected=%q, got=%q",
					srcPrefix, dstPrefix, suffix, expectedPath, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
		genPathMappingSegment(),
	))

	properties.TestingRun(t)
}

// TestPrefixLinkExactPrefixMatch tests that PrefixLink maps exact prefix paths
// **Validates: Requirements 2.6**
func TestPrefixLinkExactPrefixMatch(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLink: exact prefix path is transformed", prop.ForAll(
		func(srcPrefix, dstPrefix string) bool {
			// Create engine with prefix link mapping
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithPrefixLink(srcPrefix, dstPrefix),
				},
				nil,
			)

			// Create context with exact prefix path
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: srcPrefix,
				Path:    srcPrefix,
			}

			// Apply PrefixLink middleware
			engine.PrefixLink(ctx)

			// Path should be transformed to destination prefix
			if ctx.Path != dstPrefix {
				t.Logf("PrefixLink exact prefix mapping failed: srcPrefix=%q, expected=%q, got=%q",
					srcPrefix, dstPrefix, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
	))

	properties.TestingRun(t)
}

// TestPrefixLinkNonMatchingPath tests that PrefixLink does not modify non-matching paths
// **Validates: Requirements 2.6**
func TestPrefixLinkNonMatchingPath(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLink: non-matching path is not modified", prop.ForAll(
		func(srcPrefix, dstPrefix, otherPrefix string) bool {
			// Skip if other prefix starts with source prefix
			if strings.HasPrefix(otherPrefix, srcPrefix) {
				return true
			}

			// Create engine with prefix link mapping
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithPrefixLink(srcPrefix, dstPrefix),
				},
				nil,
			)

			// Create context with non-matching path
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: otherPrefix,
				Path:    otherPrefix,
			}

			// Apply PrefixLink middleware
			engine.PrefixLink(ctx)

			// Path should remain unchanged
			if ctx.Path != otherPrefix {
				t.Logf("PrefixLink should not modify non-matching path: srcPrefix=%q, otherPrefix=%q, got=%q",
					srcPrefix, otherPrefix, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
		genStaticPath(),
	))

	properties.TestingRun(t)
}

// TestPrefixLinkMultipleMappings tests that multiple PrefixLink mappings work correctly
// **Validates: Requirements 2.6**
func TestPrefixLinkMultipleMappings(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLink: multiple mappings work correctly (first match wins)", prop.ForAll(
		func(src1, dst1, src2, dst2, suffix string) bool {
			// Skip if sources are the same or one is prefix of another
			if src1 == src2 || strings.HasPrefix(src1, src2) || strings.HasPrefix(src2, src1) {
				return true
			}

			// Create engine with multiple prefix link mappings
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithPrefixLinkMap(map[string]string{
						src1: dst1,
						src2: dst2,
					}),
				},
				nil,
			)

			// Test first mapping
			requestPath1 := src1 + "/" + suffix
			ctx1 := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath1,
				Path:    requestPath1,
			}
			engine.PrefixLink(ctx1)
			expectedPath1 := dst1 + "/" + suffix
			if ctx1.Path != expectedPath1 {
				t.Logf("First prefix mapping failed: src=%q, expected=%q, got=%q",
					requestPath1, expectedPath1, ctx1.Path)
				return false
			}

			// Test second mapping
			requestPath2 := src2 + "/" + suffix
			ctx2 := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath2,
				Path:    requestPath2,
			}
			engine.PrefixLink(ctx2)
			expectedPath2 := dst2 + "/" + suffix
			if ctx2.Path != expectedPath2 {
				t.Logf("Second prefix mapping failed: src=%q, expected=%q, got=%q",
					requestPath2, expectedPath2, ctx2.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
		genStaticPath(),
		genStaticPath(),
		genPathMappingSegment(),
	))

	properties.TestingRun(t)
}

// TestPrefixLinkPreservesRemainder tests that PrefixLink preserves the path remainder after prefix
// **Validates: Requirements 2.6**
func TestPrefixLinkPreservesRemainder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLink: preserves path remainder after prefix replacement", prop.ForAll(
		func(srcPrefix, dstPrefix, seg1, seg2 string) bool {
			// Create engine with prefix link mapping
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithPrefixLink(srcPrefix, dstPrefix),
				},
				nil,
			)

			// Create request path with multiple segments after prefix
			remainder := "/" + seg1 + "/" + seg2
			requestPath := srcPrefix + remainder

			// Create context
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath,
				Path:    requestPath,
			}

			// Apply PrefixLink middleware
			engine.PrefixLink(ctx)

			// Path should have prefix replaced but remainder preserved
			expectedPath := dstPrefix + remainder
			if ctx.Path != expectedPath {
				t.Logf("PrefixLink remainder preservation failed: expected=%q, got=%q",
					expectedPath, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genStaticPath(),
		genPathMappingSegment(),
		genPathMappingSegment(),
	))

	properties.TestingRun(t)
}

// TestStaticLinkAndPrefixLinkCombined tests that StaticLink and PrefixLink work together
// StaticLink is applied first, then PrefixLink
// **Validates: Requirements 2.5, 2.6**
func TestStaticLinkAndPrefixLinkCombined(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("StaticLink and PrefixLink: both mappings work in sequence", prop.ForAll(
		func(staticSrc, staticDst, prefixSrc, prefixDst, suffix string) bool {
			// Skip if static destination doesn't start with prefix source
			// (we want to test the chaining behavior)
			if !strings.HasPrefix(staticDst, prefixSrc) {
				return true
			}

			// Create engine with both static and prefix link mappings
			engine := invoke.NewEngine(
				[]invoke.Option{
					invoke.WithStaticLink(staticSrc, staticDst),
					invoke.WithPrefixLink(prefixSrc, prefixDst),
				},
				nil,
			)

			// Create context with static source path
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: staticSrc,
				Path:    staticSrc,
			}

			// Apply StaticLink first
			engine.StaticLink(ctx)

			// After StaticLink, path should be staticDst
			if ctx.Path != staticDst {
				t.Logf("StaticLink step failed: expected=%q, got=%q", staticDst, ctx.Path)
				return false
			}

			// Apply PrefixLink
			engine.PrefixLink(ctx)

			// After PrefixLink, prefix should be replaced
			expectedFinal := strings.Replace(staticDst, prefixSrc, prefixDst, 1)
			if ctx.Path != expectedFinal {
				t.Logf("PrefixLink step failed: expected=%q, got=%q", expectedFinal, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
		genPrefixPath(), // staticDst is a longer path so it can have a prefix
		genStaticPath(),
		genStaticPath(),
		genPathMappingSegment(),
	))

	properties.TestingRun(t)
}

// TestStaticLinkEmptyMap tests that StaticLink with empty map does not modify path
// **Validates: Requirements 2.5**
func TestStaticLinkEmptyMap(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("StaticLink: empty map does not modify any path", prop.ForAll(
		func(requestPath string) bool {
			// Create engine with empty static link map
			engine := invoke.NewEngine(nil, nil)

			// Create context
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath,
				Path:    requestPath,
			}

			// Apply StaticLink middleware
			engine.StaticLink(ctx)

			// Path should remain unchanged
			if ctx.Path != requestPath {
				t.Logf("StaticLink with empty map should not modify path: expected=%q, got=%q",
					requestPath, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
	))

	properties.TestingRun(t)
}

// TestPrefixLinkEmptyMap tests that PrefixLink with empty map does not modify path
// **Validates: Requirements 2.6**
func TestPrefixLinkEmptyMap(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	parameters.Rng.Seed(42) // For reproducibility

	properties := gopter.NewProperties(parameters)

	properties.Property("PrefixLink: empty map does not modify any path", prop.ForAll(
		func(requestPath string) bool {
			// Create engine with empty prefix link map
			engine := invoke.NewEngine(nil, nil)

			// Create context
			ctx := &invoke.Context{
				Engine:  engine,
				RawPath: requestPath,
				Path:    requestPath,
			}

			// Apply PrefixLink middleware
			engine.PrefixLink(ctx)

			// Path should remain unchanged
			if ctx.Path != requestPath {
				t.Logf("PrefixLink with empty map should not modify path: expected=%q, got=%q",
					requestPath, ctx.Path)
				return false
			}

			return true
		},
		genStaticPath(),
	))

	properties.TestingRun(t)
}
