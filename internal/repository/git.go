package repository

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GitRepository implements Repository using go-git.
type GitRepository struct {
	logger *slog.Logger
}

func NewGitRepository(logger *slog.Logger) *GitRepository {
	return &GitRepository{logger: logger}
}

func (r *GitRepository) Clone(ctx context.Context, rawURL, dest, branch string) error {
	// Validate URL scheme and host to prevent SSRF
	if err := validateCloneURL(rawURL); err != nil {
		return fmt.Errorf("URL validation failed: %w", err)
	}

	// If the directory already exists and is a git repo, pull instead of re-cloning
	if _, err := os.Stat(dest); err == nil {
		// If the existing clone is shallow, remove and re-clone fully
		shallowFile := filepath.Join(dest, ".git", "shallow")
		if _, err := os.Stat(shallowFile); err == nil {
			r.logger.Info("removing shallow clone for full re-clone", "path", dest)
			if rmErr := os.RemoveAll(dest); rmErr != nil {
				return fmt.Errorf("remove shallow repo dir: %w", rmErr)
			}
			return r.clone(ctx, rawURL, dest, branch)
		}

		if pullErr := r.pull(ctx, dest, branch); pullErr != nil {
			// Pull failed (stale/corrupted clone) — remove and re-clone fresh
			r.logger.Warn("pull failed, removing stale clone and re-cloning",
				"path", dest, "error", pullErr)
			if rmErr := os.RemoveAll(dest); rmErr != nil {
				return fmt.Errorf("remove stale repo dir: %w", rmErr)
			}
			return r.clone(ctx, rawURL, dest, branch)
		}
		return nil
	}

	return r.clone(ctx, rawURL, dest, branch)
}

func (r *GitRepository) clone(ctx context.Context, rawURL, dest, branch string) error {
	r.logger.Info("cloning repository", "url", rawURL, "dest", dest, "branch", branch)

	opts := &git.CloneOptions{
		URL:      rawURL,
		Progress: nil, // silent
	}

	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		opts.SingleBranch = true
	}

	_, err := git.PlainCloneContext(ctx, dest, false, opts)
	if err != nil {
		return fmt.Errorf("git clone %s: %w", rawURL, err)
	}

	r.logger.Info("clone complete", "dest", dest)
	return nil
}

func (r *GitRepository) pull(ctx context.Context, dest, branch string) error {
	repo, err := git.PlainOpen(dest)
	if err != nil {
		// Not a valid git repo — remove and let caller re-clone
		r.logger.Warn("directory exists but is not a git repo, removing", "path", dest, "error", err)
		if rmErr := os.RemoveAll(dest); rmErr != nil {
			return fmt.Errorf("remove invalid repo dir: %w", rmErr)
		}
		return fmt.Errorf("removed invalid repo dir, retry needed: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("get worktree: %w", err)
	}

	// Checkout the requested branch
	refName := plumbing.NewBranchReferenceName(branch)
	r.logger.Debug("checking out branch", "branch", branch, "ref", refName)

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: refName,
		Force:  true,
	})
	if err != nil {
		r.logger.Debug("checkout failed, branch may not exist locally", "error", err)
	}

	// Pull latest changes
	r.logger.Info("pulling latest changes", "path", dest, "branch", branch)

	pullOpts := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: refName,
		SingleBranch:  true,
		Force:         true,
	}

	err = wt.PullContext(ctx, pullOpts)
	if err == git.NoErrAlreadyUpToDate {
		r.logger.Info("repository already up to date", "path", dest)
		return nil
	}
	if err != nil {
		return fmt.Errorf("git pull: %w", err)
	}

	r.logger.Info("pull complete", "path", dest)
	return nil
}

// validateCloneURL ensures the URL is safe to clone from (prevents SSRF).
func validateCloneURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow HTTPS and HTTP
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("unsupported scheme: %s (only http/https allowed)", u.Scheme)
	}

	host := u.Hostname()

	// Block localhost
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0.0.0.0" {
		return fmt.Errorf("localhost URLs not allowed")
	}

	// Block private IP ranges
	if ip := net.ParseIP(host); ip != nil {
		if isPrivateIP(ip) {
			return fmt.Errorf("private IP addresses not allowed")
		}
	}

	// Allow only well-known git hosting domains
	allowedDomains := []string{
		"github.com",
		"gitlab.com",
		"bitbucket.org",
		"codeberg.org",
		"sr.ht",
	}

	if !isAllowedDomain(host, allowedDomains) {
		return fmt.Errorf("domain not in allowlist: %s", host)
	}

	return nil
}

func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"fc00::/7",
	}
	for _, cidr := range privateRanges {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func isAllowedDomain(host string, allowed []string) bool {
	for _, d := range allowed {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}
