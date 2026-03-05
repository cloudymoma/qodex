package graph

import (
	"path/filepath"
	"strings"

	"qodex/internal/parser"
	"qodex/pkg/models"
)

// Build creates a GraphData from parsed files and their dependencies.
// It resolves import module paths to actual file paths and filters out
// any links whose source or target doesn't correspond to an existing node.
func (b *Builder) Build(files []parser.FileInfo, deps []parser.Dependency) *models.GraphData {
	graph := &models.GraphData{
		Nodes: make([]models.Node, 0, len(files)),
		Links: make([]models.Link, 0),
	}

	// Build node set and lookup maps for resolving import paths
	nodeIDs := make(map[string]bool, len(files))
	// Map directory-style module paths to actual file paths
	// e.g. "src/error" could resolve to "src/error.rs" or "src/error/mod.rs"
	moduleIndex := buildModuleIndex(files)

	for _, f := range files {
		nodeIDs[f.Path] = true
		graph.Nodes = append(graph.Nodes, models.Node{
			ID:    f.Path,
			Name:  f.Name,
			Group: languageGroup(f.Language),
			Val:   int(f.Size),
		})
	}

	// Resolve and filter links
	for _, dep := range deps {
		if !nodeIDs[dep.From] {
			continue
		}
		sourceDir := filepath.Dir(dep.From)

		for _, to := range dep.To {
			target := resolveTarget(to, sourceDir, nodeIDs, moduleIndex)
			if target == "" || target == dep.From {
				continue // skip dangling or self-referencing links
			}
			graph.Links = append(graph.Links, models.Link{
				Source: dep.From,
				Target: target,
			})
		}
	}

	return graph
}

// buildModuleIndex creates mappings from module-style paths to file paths.
// For a file "src/error/mod.rs", it indexes:
//   - "src/error/mod" -> "src/error/mod.rs"
//   - "src/error" -> "src/error/mod.rs"
//   - "error" -> "src/error/mod.rs" (short name)
//
// For "src/utils.go":
//   - "src/utils" -> "src/utils.go"
//   - "utils" -> "src/utils.go"
func buildModuleIndex(files []parser.FileInfo) map[string]string {
	idx := make(map[string]string, len(files)*3)

	for _, f := range files {
		path := f.Path
		ext := filepath.Ext(path)
		noExt := strings.TrimSuffix(path, ext)

		// Full path without extension: "src/utils" -> "src/utils.go"
		idx[noExt] = path

		// Basename without extension: "utils" -> "src/utils.go"
		base := filepath.Base(noExt)
		if _, exists := idx[base]; !exists {
			idx[base] = path
		}

		// For mod.rs / __init__.py / index.js: map parent dir
		baseLower := strings.ToLower(base)
		if baseLower == "mod" || baseLower == "__init__" || baseLower == "index" {
			parentDir := filepath.Dir(noExt)
			if parentDir != "." {
				idx[parentDir] = path
				// Also the short parent name
				shortParent := filepath.Base(parentDir)
				if _, exists := idx[shortParent]; !exists {
					idx[shortParent] = path
				}
			}
		}
	}

	return idx
}

// resolveTarget tries to resolve an import target to an actual file path.
// It attempts multiple resolution strategies:
//  1. Direct match (target is already a file path)
//  2. Relative to source directory (e.g. "./utils" from "src/main.go")
//  3. Module index lookup (e.g. "error/Result" -> strip last segment -> "error")
//  4. Prefixed with "src/" (common in Rust/Go projects)
func resolveTarget(target, sourceDir string, nodeIDs map[string]bool, moduleIndex map[string]string) string {
	// Clean up the target
	target = strings.TrimPrefix(target, "./")
	target = strings.TrimPrefix(target, "../")

	// 1. Direct match
	if nodeIDs[target] {
		return target
	}

	// 2. Relative to source directory
	rel := filepath.Join(sourceDir, target)
	if nodeIDs[rel] {
		return rel
	}

	// 3. Module index lookup (exact)
	if resolved, ok := moduleIndex[target]; ok {
		return resolved
	}

	// 4. Relative in module index
	if resolved, ok := moduleIndex[rel]; ok {
		return resolved
	}

	// 5. With "src/" prefix (Rust crate-relative imports)
	withSrc := filepath.Join("src", target)
	if nodeIDs[withSrc] {
		return withSrc
	}
	if resolved, ok := moduleIndex[withSrc]; ok {
		return resolved
	}

	// 6. Strip last segment (e.g. "error/Result" -> try "error")
	if idx := strings.LastIndex(target, "/"); idx > 0 {
		parent := target[:idx]
		return resolveTarget(parent, sourceDir, nodeIDs, moduleIndex)
	}

	// Unresolvable
	return ""
}

func languageGroup(lang string) int {
	switch lang {
	case "go":
		return models.GroupGo
	case "rust":
		return models.GroupRust
	case "javascript":
		return models.GroupJS
	case "typescript":
		return models.GroupTS
	case "python":
		return models.GroupPython
	case "java":
		return models.GroupJava
	default:
		return models.GroupOther
	}
}
