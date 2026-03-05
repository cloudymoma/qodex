package parser

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"qodex/internal/config"
)

func TestRegistryParse(t *testing.T) {
	dir := t.TempDir()

	// Create a small Go project structure
	if err := os.MkdirAll(filepath.Join(dir, "cmd"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "pkg", "util"), 0o755); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import (
	"fmt"
	"github.com/example/myapp/pkg/util"
)

func main() {
	fmt.Println(util.Hello())
}
`
	utilGo := `package util

func Hello() string {
	return "hello"
}
`
	os.WriteFile(filepath.Join(dir, "cmd", "main.go"), []byte(mainGo), 0o644)
	os.WriteFile(filepath.Join(dir, "pkg", "util", "util.go"), []byte(utilGo), 0o644)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := config.ParserConfig{
		MaxDepth:       100,
		IgnorePatterns: []string{"node_modules", "vendor"},
	}

	reg := NewRegistry(cfg, logger)
	files, deps, err := reg.Parse(context.Background(), dir)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(files) < 2 {
		t.Errorf("expected at least 2 files, got %d", len(files))
	}

	// main.go should have a dependency on "util"
	foundDep := false
	for _, d := range deps {
		if d.From == filepath.Join("cmd", "main.go") {
			for _, to := range d.To {
				if to == "util" {
					foundDep = true
				}
			}
		}
	}
	if !foundDep {
		t.Errorf("expected main.go -> util dependency, deps: %+v", deps)
	}
}

func TestRegistryBuildTree(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure
	os.MkdirAll(filepath.Join(dir, "src", "lib"), 0o755)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "app.go"), []byte("package src"), 0o644)
	os.WriteFile(filepath.Join(dir, "src", "lib", "util.go"), []byte("package lib"), 0o644)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := config.ParserConfig{
		MaxDepth:       100,
		IgnorePatterns: []string{},
	}

	reg := NewRegistry(cfg, logger)
	tree, err := reg.BuildTree(dir)
	if err != nil {
		t.Fatalf("BuildTree() error: %v", err)
	}

	if len(tree) == 0 {
		t.Fatal("expected non-empty tree")
	}

	// Should have main.go and src/ at top level
	foundMainGo := false
	foundSrc := false
	for _, node := range tree {
		if node.Name == "main.go" && node.Type == "file" {
			foundMainGo = true
		}
		if node.Name == "src" && node.Type == "directory" {
			foundSrc = true
			// Check nested structure
			if len(node.Children) == 0 {
				t.Error("src/ should have children")
			}
		}
	}

	if !foundMainGo {
		t.Error("expected main.go in tree")
	}
	if !foundSrc {
		t.Error("expected src/ directory in tree")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"main.go", "go"},
		{"app.rs", "rust"},
		{"index.js", "javascript"},
		{"App.tsx", "typescript"},
		{"script.py", "python"},
		{"Main.java", "java"},
		{"unknown.xyz", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.name)
			if got != tt.want {
				t.Errorf("detectLanguage(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"main.go", true},
		{"index.ts", true},
		{"image.png", false},
		{"binary.exe", false},
		{"Makefile", true},
		{"Dockerfile", true},
		{"data.json", true},
		{"style.css", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSourceFile(tt.name)
			if got != tt.want {
				t.Errorf("isSourceFile(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"empty", "", 0},
		{"single line", "hello", 1},
		{"two lines", "hello\nworld", 2},
		{"trailing newline", "hello\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countLines([]byte(tt.content))
			if got != tt.want {
				t.Errorf("countLines() = %d, want %d", got, tt.want)
			}
		})
	}
}
