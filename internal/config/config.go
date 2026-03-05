package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Data     DataConfig     `yaml:"data"`
	Indexer  IndexerConfig  `yaml:"indexer"`
	Parser   ParserConfig   `yaml:"parser"`
	Frontend FrontendConfig `yaml:"frontend"`
	Logging  LoggingConfig  `yaml:"logging"`
	CORS     CORSConfig     `yaml:"cors"`
}

type ServerConfig struct {
	Port            int           `yaml:"port"`
	Host            string        `yaml:"host"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// DataConfig defines the root data directory for all persistent data.
// Each repo gets subfolders: data/<repo>/code/ and data/<repo>/index/
type DataConfig struct {
	Dir string `yaml:"dir"`
}

type IndexerConfig struct {
	BatchSize     int `yaml:"batch_size"`
	MaxFileSizeMB int `yaml:"max_file_size_mb"`
}

type ParserConfig struct {
	MaxDepth       int      `yaml:"max_depth"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
}

type FrontendConfig struct {
	StaticDir string `yaml:"static_dir"`
}

type LoggingConfig struct {
	Level    string         `yaml:"level"`
	Format   string         `yaml:"format"`
	Dir      string         `yaml:"dir"`
	Rotation RotationConfig `yaml:"rotation"`
}

type RotationConfig struct {
	MaxSizeMB  int  `yaml:"max_size_mb"`
	MaxAgeDays int  `yaml:"max_age_days"`
	MaxBackups int  `yaml:"max_backups"`
	Compress   bool `yaml:"compress"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// RepoCodeDir returns the path for cloned repo source code.
func (c *Config) RepoCodeDir(repoName string) string {
	return filepath.Join(c.Data.Dir, repoName, "code")
}

// RepoIndexDir returns the path for repo's bleve search index.
func (c *Config) RepoIndexDir(repoName string) string {
	return filepath.Join(c.Data.Dir, repoName, "index")
}

// Load reads and parses the config file with environment variable expansion.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.setDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.Data.Dir == "" {
		c.Data.Dir = "./data"
	}
	if c.Logging.Dir == "" {
		c.Logging.Dir = "./logs"
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
	if c.Logging.Rotation.MaxSizeMB == 0 {
		c.Logging.Rotation.MaxSizeMB = 64
	}
	if c.Logging.Rotation.MaxAgeDays == 0 {
		c.Logging.Rotation.MaxAgeDays = 30
	}
	if c.Logging.Rotation.MaxBackups == 0 {
		c.Logging.Rotation.MaxBackups = 5
	}
}

func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Server.Port)
	}

	c.Data.Dir = expandPath(c.Data.Dir)
	c.Logging.Dir = expandPath(c.Logging.Dir)

	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
