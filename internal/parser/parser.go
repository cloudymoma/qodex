package parser

import "context"

// FileInfo holds metadata and content for a parsed file.
type FileInfo struct {
	Path     string
	Name     string
	Size     int64
	Language string
	Content  []byte
}

// Dependency represents a directional dependency from one file to others.
type Dependency struct {
	From string   // source file path
	To   []string // imported file paths
}

// Parser extracts files and dependencies from a repository directory.
type Parser interface {
	Parse(ctx context.Context, repoPath string) ([]FileInfo, []Dependency, error)
}
