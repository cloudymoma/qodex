package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"io"

	"github.com/go-git/go-git/v5"

	"qodex/internal/config"
	"qodex/internal/graph"
	"qodex/internal/indexer"
	"qodex/internal/parser"
	"qodex/internal/repository"
	"qodex/pkg/models"
)

var (
	safeNameRE   = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)
	safeBranchRE = regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`)
)

// IngestService orchestrates the clone → parse → index pipeline.
type IngestService struct {
	cfg     *config.Config
	repo    repository.Repository
	parser  *parser.Registry
	indexer indexer.Indexer
	builder *graph.Builder
	logger  *slog.Logger

	mu              sync.RWMutex
	data            *graph.Data
	repos           []models.RepoEntry
	currentRepoName string
}

func NewIngestService(
	cfg *config.Config,
	repo repository.Repository,
	psr *parser.Registry,
	idx indexer.Indexer,
	builder *graph.Builder,
	logger *slog.Logger,
) *IngestService {
	return &IngestService{
		cfg:     cfg,
		repo:    repo,
		parser:  psr,
		indexer: idx,
		builder: builder,
		logger:  logger,
		data:    graph.NewData(),
	}
}

// Ingest clones a repository, parses dependencies, builds the graph and index.
func (s *IngestService) Ingest(ctx context.Context, req *models.IngestRequest) (*models.IngestResponse, error) {
	repoName, err := ExtractRepoName(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %w", err)
	}

	s.logger.Debug("ingest started",
		"url", req.URL,
		"repo_name", repoName,
	)

	// Validate repo name to prevent path traversal
	if !safeNameRE.MatchString(repoName) || strings.Contains(repoName, "..") {
		return nil, fmt.Errorf("invalid repository name: %s", repoName)
	}

	branch := req.Branch
	if branch == "" {
		branch = "main"
	}

	// Validate branch name
	if !safeBranchRE.MatchString(branch) || strings.Contains(branch, "..") {
		return nil, fmt.Errorf("invalid branch name: %s", branch)
	}

	// Use structured data dirs: data/<repo>/code, data/<repo>/index
	codeDir := s.cfg.RepoCodeDir(repoName)
	indexDir := s.cfg.RepoIndexDir(repoName)

	// Verify the resolved path is still within the data directory
	cleanBase := filepath.Clean(s.cfg.Data.Dir)
	cleanCode := filepath.Clean(codeDir)
	if !strings.HasPrefix(cleanCode, cleanBase+string(filepath.Separator)) {
		return nil, fmt.Errorf("path traversal detected")
	}

	// Ensure repo subdirectories exist
	if err := os.MkdirAll(filepath.Dir(codeDir), 0o755); err != nil {
		return nil, fmt.Errorf("create repo data dir: %w", err)
	}

	s.logger.Debug("repo paths resolved",
		"code_dir", codeDir,
		"index_dir", indexDir,
		"branch", branch,
	)

	// Step 1: Clone with timeout
	s.logger.Debug("step 1: cloning repository", "url", req.URL, "branch", branch)
	cloneCtx, cloneCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cloneCancel()

	if err := s.repo.Clone(cloneCtx, req.URL, codeDir, branch); err != nil {
		return &models.IngestResponse{
			RepoName: repoName,
			Status:   "error",
			Message:  fmt.Sprintf("clone failed: %v", err),
		}, err
	}
	s.logger.Debug("step 1 complete: clone finished")

	// Step 2: Parse files and dependencies
	s.logger.Debug("step 2: parsing files and dependencies")
	files, deps, err := s.parser.Parse(ctx, codeDir)
	if err != nil {
		return &models.IngestResponse{
			RepoName: repoName,
			Status:   "error",
			Message:  fmt.Sprintf("parse failed: %v", err),
		}, err
	}
	s.logger.Debug("step 2 complete: parsing finished",
		"files", len(files),
		"dependencies", len(deps),
	)

	// Step 3: Build graph
	s.logger.Debug("step 3: building dependency graph")
	graphData := s.builder.Build(files, deps)
	s.logger.Debug("step 3 complete: graph built",
		"nodes", len(graphData.Nodes),
		"links", len(graphData.Links),
	)

	// Step 4: Build tree
	s.logger.Debug("step 4: building file tree")
	tree, err := s.parser.BuildTree(codeDir)
	if err != nil {
		return &models.IngestResponse{
			RepoName: repoName,
			Status:   "error",
			Message:  fmt.Sprintf("tree building failed: %v", err),
		}, err
	}
	s.logger.Debug("step 4 complete: tree built", "top_level_entries", len(tree))

	// Step 5: Index files for search
	s.logger.Debug("step 5: indexing files for search", "index_dir", indexDir)
	if err := s.indexer.Index(ctx, files, indexDir); err != nil {
		return &models.IngestResponse{
			RepoName: repoName,
			Status:   "error",
			Message:  fmt.Sprintf("indexing failed: %v", err),
		}, err
	}
	s.logger.Debug("step 5 complete: indexing finished")

	// Step 6: Store results in memory and track repo
	s.mu.Lock()
	s.data.Graph = graphData
	s.data.Tree = tree
	s.currentRepoName = repoName
	s.addRepoLocked(req.URL, branch, repoName)
	s.mu.Unlock()

	s.logger.Info("ingest complete",
		"repo", repoName,
		"files", len(files),
		"nodes", len(graphData.Nodes),
		"links", len(graphData.Links),
	)

	return &models.IngestResponse{
		RepoName:     repoName,
		Status:       "success",
		FilesIndexed: len(files),
	}, nil
}

// GraphData returns a copy of the current in-memory graph.
func (s *IngestService) GraphData() *models.GraphData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external mutation
	nodesCopy := make([]models.Node, len(s.data.Graph.Nodes))
	copy(nodesCopy, s.data.Graph.Nodes)

	linksCopy := make([]models.Link, len(s.data.Graph.Links))
	copy(linksCopy, s.data.Graph.Links)

	return &models.GraphData{
		Nodes: nodesCopy,
		Links: linksCopy,
	}
}

// TreeData returns a copy of the current file tree.
func (s *IngestService) TreeData() []*models.TreeNode {
	s.mu.RLock()
	defer s.mu.RUnlock()

	treeCopy := make([]*models.TreeNode, len(s.data.Tree))
	copy(treeCopy, s.data.Tree)
	return treeCopy
}

// ListRepos returns a copy of all previously ingested repos.
func (s *IngestService) ListRepos() []models.RepoEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.RepoEntry, len(s.repos))
	copy(out, s.repos)
	return out
}

// addRepoLocked upserts a repo entry. Must be called with mu held.
func (s *IngestService) addRepoLocked(url, branch, repoName string) {
	for i, r := range s.repos {
		if r.RepoName == repoName {
			s.repos[i].URL = url
			s.repos[i].Branch = branch
			return
		}
	}
	s.repos = append(s.repos, models.RepoEntry{
		URL:      url,
		Branch:   branch,
		RepoName: repoName,
	})
}

// ExtractRepoName derives a directory-safe name from a GitHub URL.
func ExtractRepoName(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	path := strings.TrimSuffix(u.Path, ".git")
	path = strings.Trim(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("URL must contain owner/repo: %s", rawURL)
	}

	owner := sanitizePart(parts[len(parts)-2])
	repo := sanitizePart(parts[len(parts)-1])

	if owner == "" || repo == "" {
		return "", fmt.Errorf("invalid owner/repo in URL: %s", rawURL)
	}

	return owner + "-" + repo, nil
}

// FileContent reads a file from the currently ingested repo's code directory.
// It validates against path traversal and caps file size at 1MB.
func (s *IngestService) FileContent(path string) (string, error) {
	s.mu.RLock()
	repoName := s.currentRepoName
	s.mu.RUnlock()

	if repoName == "" {
		return "", fmt.Errorf("no repository ingested")
	}

	// Validate path: no empty, no "..", must be relative
	if path == "" {
		return "", fmt.Errorf("invalid path: empty")
	}
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal detected")
	}
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("invalid path: must be relative")
	}

	codeDir := s.cfg.RepoCodeDir(repoName)
	fullPath := filepath.Join(codeDir, filepath.FromSlash(path))
	cleanFull := filepath.Clean(fullPath)
	cleanBase := filepath.Clean(codeDir)

	// Prefix check to prevent traversal
	if !strings.HasPrefix(cleanFull, cleanBase+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected")
	}

	// Check file exists and size
	info, err := os.Stat(cleanFull)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", path)
		}
		return "", fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", path)
	}

	const maxFileSize = 1 << 20 // 1MB
	if info.Size() > maxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileSize)
	}

	data, err := os.ReadFile(cleanFull)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	return string(data), nil
}

// CurrentRepoURL returns the URL of the currently ingested repo.
func (s *IngestService) CurrentRepoURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.repos {
		if r.RepoName == s.currentRepoName {
			return r.URL
		}
	}
	return ""
}

// CommitHistory returns the git commit log for the currently ingested repo.
func (s *IngestService) CommitHistory(limit int) ([]models.CommitEntry, error) {
	s.mu.RLock()
	repoName := s.currentRepoName
	s.mu.RUnlock()

	if repoName == "" {
		return nil, fmt.Errorf("no repository ingested")
	}

	codeDir := s.cfg.RepoCodeDir(repoName)
	repo, err := git.PlainOpen(codeDir)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	logIter, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	defer logIter.Close()

	var commits []models.CommitEntry
	for i := 0; i < limit; i++ {
		c, err := logIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iterate commits: %w", err)
		}
		hash := c.Hash.String()
		commits = append(commits, models.CommitEntry{
			Hash:    hash,
			Short:   hash[:7],
			Message: strings.TrimSpace(c.Message),
			Author:  c.Author.Name,
			Date:    c.Author.When.Format("2006-01-02 15:04"),
		})
	}

	return commits, nil
}

// sanitizePart removes unsafe characters from URL path segments.
func sanitizePart(s string) string {
	s = strings.ReplaceAll(s, "..", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "\\", "")
	return strings.TrimSpace(s)
}
