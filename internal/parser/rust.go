package parser

import (
	"regexp"
	"strings"
)

var (
	// use crate::module::submodule;
	rustUseRE = regexp.MustCompile(`(?m)^use\s+(crate::[a-zA-Z_][\w:]*)\s*;`)
	// mod module_name;
	rustModRE = regexp.MustCompile(`(?m)^mod\s+([a-zA-Z_]\w*)\s*;`)
)

// parseRustImports extracts use/mod references from Rust source files.
func parseRustImports(content []byte, filePath string) []string {
	src := string(content)
	seen := make(map[string]bool)

	for _, match := range rustUseRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 {
			// Convert crate::foo::bar to foo/bar
			path := strings.TrimPrefix(match[1], "crate::")
			path = strings.ReplaceAll(path, "::", "/")
			seen[path] = true
		}
	}

	for _, match := range rustModRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 {
			seen[match[1]] = true
		}
	}

	var deps []string
	for dep := range seen {
		deps = append(deps, dep)
	}
	return deps
}
