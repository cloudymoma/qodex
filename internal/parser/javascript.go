package parser

import (
	"regexp"
)

var (
	// import ... from './path' or import ... from 'module'
	jsImportFromRE = regexp.MustCompile(`(?m)import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
	// require('./path') or require('module')
	jsRequireRE = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
)

// parseJSImports extracts import/require paths from JavaScript/TypeScript files.
func parseJSImports(content []byte, filePath string) []string {
	src := string(content)
	seen := make(map[string]bool)

	for _, match := range jsImportFromRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 && isRelativePath(match[1]) {
			seen[match[1]] = true
		}
	}

	for _, match := range jsRequireRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 && isRelativePath(match[1]) {
			seen[match[1]] = true
		}
	}

	var deps []string
	for dep := range seen {
		// Normalize: strip leading ./ and resolve to a rough path
		deps = append(deps, normalizeJSPath(dep))
	}
	return deps
}

func isRelativePath(p string) bool {
	return len(p) > 0 && (p[0] == '.' || p[0] == '/')
}

func normalizeJSPath(p string) string {
	if len(p) > 2 && p[:2] == "./" {
		return p[2:]
	}
	if len(p) > 3 && p[:3] == "../" {
		return p[3:]
	}
	return p
}
