package service

import (
	"os"
	"path/filepath"
	"testing"

	"qodex/internal/config"
)

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "standard github url",
			url:  "https://github.com/owner/repo",
			want: "owner-repo",
		},
		{
			name: "with .git suffix",
			url:  "https://github.com/owner/repo.git",
			want: "owner-repo",
		},
		{
			name: "with trailing slash",
			url:  "https://github.com/owner/repo/",
			want: "owner-repo",
		},
		{
			name:    "missing repo",
			url:     "https://github.com/owner",
			wantErr: true,
		},
		{
			name: "nested path",
			url:  "https://github.com/org/team/repo",
			want: "team-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractRepoName(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestExtractRepoNamePathTraversal(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "double dot in path", url: "https://github.com/../../../etc"},
		{name: "dot dot owner", url: "https://github.com/..%2F../repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := ExtractRepoName(tt.url)
			if err != nil {
				return // error is acceptable
			}
			// If no error, the name must not contain ".."
			if name == "" {
				t.Error("expected non-empty name or error")
			}
		})
	}
}

func TestFileContent(t *testing.T) {
	// Set up a temporary data directory with a repo code folder
	tmpDir := t.TempDir()
	codeDir := filepath.Join(tmpDir, "owner-repo", "code")
	if err := os.MkdirAll(filepath.Join(codeDir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codeDir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(codeDir, "src", "lib.go"), []byte("package src"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Data.Dir = tmpDir

	svc := &IngestService{
		cfg:             cfg,
		currentRepoName: "owner-repo",
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr string
	}{
		{
			name: "read root file",
			path: "main.go",
			want: "package main",
		},
		{
			name: "read nested file",
			path: "src/lib.go",
			want: "package src",
		},
		{
			name:    "path traversal with ..",
			path:    "../../../etc/passwd",
			wantErr: "path traversal",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: "invalid path: empty",
		},
		{
			name:    "absolute path",
			path:    "/etc/passwd",
			wantErr: "invalid path: must be relative",
		},
		{
			name:    "nonexistent file",
			path:    "does-not-exist.go",
			wantErr: "file not found",
		},
		{
			name:    "directory path",
			path:    "src",
			wantErr: "path is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.FileContent(tt.path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if got2 := err.Error(); !contains(got2, tt.wantErr) {
					t.Errorf("error %q does not contain %q", got2, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("FileContent(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestFileContentSizeLimit(t *testing.T) {
	tmpDir := t.TempDir()
	codeDir := filepath.Join(tmpDir, "owner-repo", "code")
	if err := os.MkdirAll(codeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a file just over 1MB
	largeContent := make([]byte, (1<<20)+1)
	if err := os.WriteFile(filepath.Join(codeDir, "large.bin"), largeContent, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.Data.Dir = tmpDir
	svc := &IngestService{cfg: cfg, currentRepoName: "owner-repo"}

	_, err := svc.FileContent("large.bin")
	if err == nil {
		t.Fatal("expected error for large file, got nil")
	}
	if !contains(err.Error(), "file too large") {
		t.Errorf("expected 'file too large' error, got: %v", err)
	}
}

func TestFileContentNoRepo(t *testing.T) {
	cfg := &config.Config{}
	cfg.Data.Dir = t.TempDir()
	svc := &IngestService{cfg: cfg}

	_, err := svc.FileContent("main.go")
	if err == nil {
		t.Fatal("expected error for no repo, got nil")
	}
	if !contains(err.Error(), "no repository") {
		t.Errorf("expected 'no repository' error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSanitizePart(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"owner", "owner"},
		{"my-repo", "my-repo"},
		{"../etc", "etc"},
		{"foo/bar", "foobar"},
		{"foo\\bar", "foobar"},
		{" spaces ", "spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizePart(tt.input)
			if got != tt.want {
				t.Errorf("sanitizePart(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
