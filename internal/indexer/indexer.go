package indexer

import (
	"context"

	"qodex/internal/parser"
	"qodex/pkg/models"
)

// Indexer indexes file contents for full-text search.
type Indexer interface {
	Index(ctx context.Context, files []parser.FileInfo, indexPath string) error
	Search(ctx context.Context, query string, limit int) ([]models.SearchResult, error)
	Close() error
}
