package parser

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"qodex/internal/config"
	"qodex/pkg/models"
)

// Registry walks a repository directory, collects file info, and delegates
// import parsing to language-specific parsers.
type Registry struct {
	cfg      config.ParserConfig
	logger   *slog.Logger
	goParser *GoParser
}

func NewRegistry(cfg config.ParserConfig, logger *slog.Logger) *Registry {
	return &Registry{
		cfg:      cfg,
		logger:   logger,
		goParser: NewGoParser(logger),
	}
}

// Parse walks the repo directory, collects files, and parses dependencies.
func (r *Registry) Parse(ctx context.Context, repoPath string) ([]FileInfo, []Dependency, error) {
	var files []FileInfo
	var deps []Dependency

	var maxSize int64 = 10 * 1024 * 1024 // default 10MB

	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		name := d.Name()

		// Skip ignored directories
		if d.IsDir() {
			if r.shouldIgnore(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-text/source files
		if !isSourceFile(name) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Skip files that are too large
		if info.Size() > maxSize {
			return nil
		}

		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lang := detectLanguage(name)

		fi := FileInfo{
			Path:     relPath,
			Name:     name,
			Size:     int64(countLines(content)),
			Language: lang,
			Content:  content,
		}
		files = append(files, fi)

		// Parse imports for supported languages
		fileDeps := r.parseImports(fi, repoPath)
		if len(fileDeps) > 0 {
			deps = append(deps, Dependency{
				From: relPath,
				To:   fileDeps,
			})
		}

		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("walk repository: %w", err)
	}

	r.logger.Info("parse complete", "files", len(files), "dependencies", len(deps))
	return files, deps, nil
}

// BuildTree constructs a hierarchical TreeNode from the parsed files.
func (r *Registry) BuildTree(repoPath string) ([]*models.TreeNode, error) {
	root := &models.TreeNode{
		Name: filepath.Base(repoPath),
		Path: "",
		Type: "directory",
	}

	err := filepath.WalkDir(repoPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()
		if d.IsDir() && r.shouldIgnore(name) {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(repoPath, path)
		if err != nil || relPath == "." {
			return nil
		}

		nodeType := "file"
		if d.IsDir() {
			nodeType = "directory"
		}

		node := &models.TreeNode{
			Name: name,
			Path: relPath,
			Type: nodeType,
		}

		// Find the parent and attach
		insertIntoTree(root, node, relPath)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("build tree: %w", err)
	}

	if root.Children == nil {
		return []*models.TreeNode{}, nil
	}
	return root.Children, nil
}

func insertIntoTree(root *models.TreeNode, node *models.TreeNode, relPath string) {
	parts := strings.Split(relPath, string(filepath.Separator))
	current := root

	// Navigate to the parent directory
	for i := 0; i < len(parts)-1; i++ {
		found := false
		for _, child := range current.Children {
			if child.Name == parts[i] && child.Type == "directory" {
				current = child
				found = true
				break
			}
		}
		if !found {
			// Parent directory not found, create it
			dirNode := &models.TreeNode{
				Name: parts[i],
				Path: strings.Join(parts[:i+1], "/"),
				Type: "directory",
			}
			current.Children = append(current.Children, dirNode)
			current = dirNode
		}
	}

	current.Children = append(current.Children, node)
}

func (r *Registry) shouldIgnore(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	for _, pattern := range r.cfg.IgnorePatterns {
		if name == pattern {
			return true
		}
	}
	return false
}

func (r *Registry) parseImports(fi FileInfo, repoPath string) []string {
	switch fi.Language {
	case "go":
		return r.goParser.ParseImports(fi.Content, fi.Path)
	case "javascript", "typescript":
		return parseJSImports(fi.Content, fi.Path)
	case "python":
		return parsePythonImports(fi.Content, fi.Path)
	case "rust":
		return parseRustImports(fi.Content, fi.Path)
	default:
		return nil
	}
}

func detectLanguage(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "go"
	case ".rs":
		return "rust"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".rb":
		return "ruby"
	case ".swift":
		return "swift"
	case ".kt", ".kts":
		return "kotlin"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	default:
		return "other"
	}
}

func isSourceFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	sourceExts := map[string]bool{
		".go": true, ".rs": true, ".js": true, ".jsx": true, ".mjs": true, ".cjs": true,
		".ts": true, ".tsx": true, ".py": true, ".java": true,
		".c": true, ".h": true, ".cpp": true, ".cc": true, ".cxx": true, ".hpp": true,
		".rb": true, ".swift": true, ".kt": true, ".kts": true, ".cs": true, ".php": true,
		".yaml": true, ".yml": true, ".json": true, ".toml": true,
		".md": true, ".txt": true, ".cfg": true, ".ini": true,
		".sh": true, ".bash": true, ".zsh": true,
		".html": true, ".css": true, ".scss": true, ".less": true,
		".sql": true, ".graphql": true, ".gql": true,
		".proto": true, ".xml": true, ".svg": true,
		".makefile": true,
	}

	if sourceExts[ext] {
		return true
	}

	// Handle extensionless files like Makefile, Dockerfile
	lower := strings.ToLower(name)
	special := map[string]bool{
		"makefile": true, "dockerfile": true, "jenkinsfile": true,
		"rakefile": true, "gemfile": true, "procfile": true,
	}
	return special[lower]
}

func countLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	return bytes.Count(content, []byte("\n")) + 1
}
