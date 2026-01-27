package event

import (
	"fmt"
	"log"
	"strings"
)

// InstallHandlers sets up the default router and handlers.
// This method is called automatically by NewEngine.
//
// Validates: Requirements 2.6, 2.7, 3.1, 3.2, 3.3, 3.4
func (e *Engine) InstallHandlers() {
	e.r = NewRouter()

	// Register middleware for path rewriting
	e.r.Use(e.HeaderLink)
	e.r.Use(e.StaticLink)
	e.r.Use(e.PrefixLink)

	// Register route handlers
	e.r.Handle("/", e.OK)
	e.r.Handle("/_/api/*path", e.Debug, e.WAPI)
	e.r.Handle("/api/*path", e.API)

	// Set NoRoute handler
	e.r.NoRoute(e.PageNotFound)
	e.r.NoMethod(e.MethodNotAllowed)
}

// HeaderLink is a middleware placeholder for header-based path rewriting.
// Reserved for future use.
func (e *Engine) HeaderLink(c *Context) {
	// Placeholder for header-based link rewriting
}

// StaticLink rewrites paths based on StaticLinkMap configuration.
//
// Validates: Requirement 2.6
func (e *Engine) StaticLink(c *Context) {
	if e.StaticLinkMap == nil {
		return
	}
	if dst, ok := e.StaticLinkMap[c.Path]; ok {
		c.Path = dst
	}
}

// PrefixLink rewrites paths based on PrefixLinkMap configuration.
//
// Validates: Requirement 2.7
func (e *Engine) PrefixLink(c *Context) {
	if e.PrefixLinkMap == nil {
		return
	}
	for srcPrefix, dstPrefix := range e.PrefixLinkMap {
		if strings.HasPrefix(c.Path, srcPrefix) {
			c.Path = strings.Replace(c.Path, srcPrefix, dstPrefix, 1)
			return
		}
	}
}

// OK handles the root path request.
func (e *Engine) OK(c *Context) {
	// Root path handler - no action needed for event mode
}

// Debug is a middleware that enables debug mode for the request.
func (e *Engine) Debug(c *Context) {
	c.DebugMode = true
}

// API handles API requests by invoking the Dynamic package.
//
// Validates: Requirements 3.1, 3.2, 3.3
func (e *Engine) API(c *Context) {
	e.handle(c)
}

// WAPI handles wrapped API requests (debug mode).
//
// Validates: Requirement 3.4
func (e *Engine) WAPI(c *Context) {
	e.handle(c)
}

// PageNotFound handles requests that don't match any route.
//
// Validates: Requirement 2.2
func (e *Engine) PageNotFound(c *Context) {
	c.Err = fmt.Errorf("event: page not found: %s", c.Path)
}

// MethodNotAllowed handles method not allowed errors (reserved for future use).
func (e *Engine) MethodNotAllowed(c *Context) {
	c.Err = fmt.Errorf("event: method not allowed: %s", c.Path)
}

// handle invokes the Dynamic package based on the request path.
//
// Validates: Requirements 3.1, 3.2, 3.3, 8.3
func (e *Engine) handle(c *Context) {
	// Extract package name and version from path
	// Path format: /package/version/route or /package/route (uses default version)
	pkg, version, route, err := parsePath(c.ParamPath)
	if err != nil {
		c.Err = err
		return
	}

	if c.DebugMode {
		log.Printf("[Event] Dynamic call: pkg=%s version=%s route=%s", pkg, version, route)
	}

	// Get the package tunnel
	tunnel, err := e.GetPackage(pkg, version)
	if err != nil {
		c.Err = fmt.Errorf("event: package not found: %s_%s: %w", pkg, version, err)
		return
	}

	// Invoke the tunnel - response is ignored in event mode (Requirement 1.3)
	_ = tunnel.Invoke(route, c.Request)
}

// parsePath extracts package, version, and route from the path.
// Path format: /package/version/route or /package/route
//
// Validates: Requirements 3.1, 8.3
func parsePath(path string) (pkg, version, route string, err error) {
	if path == "" || path == "/" {
		return "", "", "", fmt.Errorf("event: invalid path format: path is empty")
	}

	// Remove leading slash
	if path[0] == '/' {
		path = path[1:]
	}

	// Split path into parts
	parts := splitPath(path)
	if len(parts) < 1 {
		return "", "", "", fmt.Errorf("event: invalid path format: missing package name")
	}

	pkg = parts[0]
	if pkg == "" {
		return "", "", "", fmt.Errorf("event: invalid path format: empty package name")
	}

	// Determine version and route based on path structure
	if len(parts) >= 2 {
		// Check if second part looks like a version (e.g., "v1", "1.0", "commit")
		// For simplicity, we treat the second part as version if there are 3+ parts
		// or if it matches common version patterns
		if len(parts) >= 3 || isVersion(parts[1]) {
			version = parts[1]
			if len(parts) > 2 {
				route = "/" + joinPath(parts[2:])
			} else {
				route = "/"
			}
		} else {
			// Second part is the route, use default version
			version = ""
			route = "/" + joinPath(parts[1:])
		}
	} else {
		version = ""
		route = "/"
	}

	return pkg, version, route, nil
}

// isVersion checks if a string looks like a version identifier.
func isVersion(s string) bool {
	if s == "" {
		return false
	}
	// Common version patterns: v1, v2, commit, latest, or semantic versions
	if s[0] == 'v' || s == "commit" || s == "latest" {
		return true
	}
	// Check for numeric version (e.g., "1", "1.0", "1.0.0")
	for _, c := range s {
		if c != '.' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// splitPath splits a path by '/' and returns non-empty parts.
func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	return parts
}

// joinPath joins path parts with '/'.
func joinPath(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "/")
}
