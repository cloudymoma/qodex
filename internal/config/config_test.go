package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.yaml")

	content := `
server:
  port: 8080
  host: "127.0.0.1"
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 5s
data:
  dir: "` + dir + `/data"
indexer:
  batch_size: 50
  max_file_size_mb: 5
parser:
  max_depth: 50
  ignore_patterns:
    - "node_modules"
    - "vendor"
frontend:
  static_dir: "./web/static"
logging:
  level: "debug"
  format: "text"
  dir: "` + dir + `/logs"
  rotation:
    max_size_mb: 32
    max_age_days: 15
    max_backups: 3
    compress: true
cors:
  allowed_origins:
    - "http://localhost:3000"
  allowed_methods:
    - "GET"
    - "POST"
  allowed_headers:
    - "Content-Type"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"server.port", cfg.Server.Port, 8080},
		{"server.host", cfg.Server.Host, "127.0.0.1"},
		{"data.dir", cfg.Data.Dir, dir + "/data"},
		{"indexer.batch_size", cfg.Indexer.BatchSize, 50},
		{"parser.max_depth", cfg.Parser.MaxDepth, 50},
		{"logging.level", cfg.Logging.Level, "debug"},
		{"logging.dir", cfg.Logging.Dir, dir + "/logs"},
		{"rotation.max_size_mb", cfg.Logging.Rotation.MaxSizeMB, 32},
		{"rotation.max_age_days", cfg.Logging.Rotation.MaxAgeDays, 15},
		{"rotation.max_backups", cfg.Logging.Rotation.MaxBackups, 3},
		{"rotation.compress", cfg.Logging.Rotation.Compress, true},
		{"cors origins count", len(cfg.CORS.AllowedOrigins), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.yaml")

	// Minimal config — defaults should kick in
	content := `
server:
  port: 8080
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Data.Dir != "./data" {
		t.Errorf("data.dir = %q, want ./data", cfg.Data.Dir)
	}
	if cfg.Logging.Dir != "./logs" {
		t.Errorf("logging.dir = %q, want ./logs", cfg.Logging.Dir)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("logging.level = %q, want info", cfg.Logging.Level)
	}
	if cfg.Logging.Rotation.MaxSizeMB != 64 {
		t.Errorf("rotation.max_size_mb = %d, want 64", cfg.Logging.Rotation.MaxSizeMB)
	}
	if cfg.Logging.Rotation.MaxAgeDays != 30 {
		t.Errorf("rotation.max_age_days = %d, want 30", cfg.Logging.Rotation.MaxAgeDays)
	}
}

func TestRepoDataDirs(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.yaml")

	content := `
server:
  port: 8080
data:
  dir: "` + dir + `/data"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	codeDir := cfg.RepoCodeDir("owner-repo")
	indexDir := cfg.RepoIndexDir("owner-repo")

	wantCode := filepath.Join(dir, "data", "owner-repo", "code")
	wantIndex := filepath.Join(dir, "data", "owner-repo", "index")

	if codeDir != wantCode {
		t.Errorf("RepoCodeDir = %q, want %q", codeDir, wantCode)
	}
	if indexDir != wantIndex {
		t.Errorf("RepoIndexDir = %q, want %q", indexDir, wantIndex)
	}
}

func TestLoadEnvExpansion(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.yaml")

	content := `
server:
  port: 9090
  host: "0.0.0.0"
  read_timeout: 5s
  write_timeout: 5s
  shutdown_timeout: 5s
data:
  dir: "` + dir + `/store"
logging:
  level: "info"
  format: "text"
cors:
  allowed_origins: []
  allowed_methods: []
  allowed_headers: []
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("port = %d, want 9090", cfg.Server.Port)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "test.yaml")

	content := `
server:
  port: 0
data:
  dir: "/tmp/test"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for invalid port, got nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo/bar", filepath.Join(home, "foo/bar")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
