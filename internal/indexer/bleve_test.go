package indexer

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"qodex/internal/config"
	"qodex/internal/parser"
)

func TestBleveIndexerIndexAndSearch(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "test-index")

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := config.IndexerConfig{BatchSize: 10, MaxFileSizeMB: 10}

	idx, err := NewBleveIndexer(cfg, logger)
	if err != nil {
		t.Fatalf("NewBleveIndexer() error: %v", err)
	}
	defer idx.Close()

	files := []parser.FileInfo{
		{
			Path:     "main.go",
			Name:     "main.go",
			Language: "go",
			Content:  []byte(`package main\n\nfunc main() {\n\tfmt.Println("hello world")\n}`),
		},
		{
			Path:     "util.go",
			Name:     "util.go",
			Language: "go",
			Content:  []byte(`package util\n\nfunc FormatName(name string) string {\n\treturn "Hello, " + name\n}`),
		},
		{
			Path:     "app.js",
			Name:     "app.js",
			Language: "javascript",
			Content:  []byte(`function greetUser(user) {\n\tconsole.log("hello " + user)\n}`),
		},
	}

	ctx := context.Background()

	// Test indexing
	if err := idx.Index(ctx, files, indexPath); err != nil {
		t.Fatalf("Index() error: %v", err)
	}

	// Test search for "hello"
	results, err := idx.Search(ctx, "hello", 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected search results for 'hello', got none")
	}

	// All files contain "hello", so we should get multiple results
	foundMainGo := false
	for _, r := range results {
		if r.FilePath == "main.go" {
			foundMainGo = true
		}
	}
	if !foundMainGo {
		t.Error("expected main.go in search results")
	}
}

func TestBleveIndexerSearchBeforeIndex(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := config.IndexerConfig{BatchSize: 10}

	idx, err := NewBleveIndexer(cfg, logger)
	if err != nil {
		t.Fatalf("NewBleveIndexer() error: %v", err)
	}
	defer idx.Close()

	// Search without indexing should return nil
	results, err := idx.Search(context.Background(), "hello", 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results, got %d", len(results))
	}
}

func TestBleveIndexerReindex(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "reindex-test")

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := config.IndexerConfig{BatchSize: 10}

	idx, err := NewBleveIndexer(cfg, logger)
	if err != nil {
		t.Fatalf("NewBleveIndexer() error: %v", err)
	}
	defer idx.Close()

	ctx := context.Background()

	// First index
	files1 := []parser.FileInfo{
		{Path: "old.go", Name: "old.go", Language: "go", Content: []byte("package old\nfunc OldFunc() {}")},
	}
	if err := idx.Index(ctx, files1, indexPath); err != nil {
		t.Fatalf("first Index() error: %v", err)
	}

	// Re-index with different content
	files2 := []parser.FileInfo{
		{Path: "new.go", Name: "new.go", Language: "go", Content: []byte("package new\nfunc NewFunc() {}")},
	}
	if err := idx.Index(ctx, files2, indexPath); err != nil {
		t.Fatalf("second Index() error: %v", err)
	}

	// Search for new content
	results, err := idx.Search(ctx, "NewFunc", 10)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) == 0 {
		t.Error("expected results for 'NewFunc' after reindex")
	}
}
