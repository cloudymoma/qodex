package parser

import (
	"context"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Match single import: import "path/to/pkg"
	singleImportRE = regexp.MustCompile(`import\s+"([^"]+)"`)

	// Match multi-line import block: import ( ... )
	multiImportRE = regexp.MustCompile(`(?s)import\s+\((.*?)\)`)

	// Extract individual import path from within import block
	importPathRE = regexp.MustCompile(`"([^"]+)"`)
)

// GoParser parses Go source files for import dependencies.
type GoParser struct {
	logger *slog.Logger
}

func NewGoParser(logger *slog.Logger) *GoParser {
	return &GoParser{logger: logger}
}

// Parse implements the Parser interface for Go files.
func (p *GoParser) Parse(ctx context.Context, repoPath string) ([]FileInfo, []Dependency, error) {
	// This method is not directly used — the Registry handles walking.
	// Kept for interface compliance.
	return nil, nil, nil
}

// ParseImports extracts import paths from Go source content and maps them
// to relative file paths within the repository when possible.
func (p *GoParser) ParseImports(content []byte, filePath string) []string {
	src := string(content)
	importPaths := make(map[string]bool)

	// Find single-line imports
	for _, match := range singleImportRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 {
			importPaths[match[1]] = true
		}
	}

	// Find multi-line import blocks
	for _, block := range multiImportRE.FindAllStringSubmatch(src, -1) {
		if len(block) > 1 {
			for _, pathMatch := range importPathRE.FindAllStringSubmatch(block[1], -1) {
				if len(pathMatch) > 1 {
					importPaths[pathMatch[1]] = true
				}
			}
		}
	}

	// Convert import paths to relative file references
	var deps []string
	for imp := range importPaths {
		// Skip standard library imports (no dots in path)
		if isStdLib(imp) {
			continue
		}
		// Use the last path segment as a rough file reference
		// This creates links between packages within the repo
		rel := lastPathSegment(imp)
		if rel != "" && rel != filepath.Dir(filePath) {
			deps = append(deps, rel)
		}
	}

	return deps
}

// isStdLib returns true if the import path looks like a Go standard library package.
func isStdLib(path string) bool {
	// Standard library packages don't contain a dot in the first segment
	firstSlash := strings.IndexByte(path, '/')
	firstPart := path
	if firstSlash > 0 {
		firstPart = path[:firstSlash]
	}
	return !strings.Contains(firstPart, ".")
}

// lastPathSegment returns the last segment of an import path.
func lastPathSegment(imp string) string {
	parts := strings.Split(imp, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
