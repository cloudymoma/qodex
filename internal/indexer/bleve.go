package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"

	"qodex/internal/config"
	"qodex/internal/parser"
	"qodex/pkg/models"
)

// fileDocument is the bleve document structure for indexing.
type fileDocument struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

// BleveIndexer implements Indexer using the bleve search library.
type BleveIndexer struct {
	cfg    config.IndexerConfig
	logger *slog.Logger
	index  bleve.Index
}

func NewBleveIndexer(cfg config.IndexerConfig, logger *slog.Logger) (*BleveIndexer, error) {
	return &BleveIndexer{cfg: cfg, logger: logger}, nil
}

func (b *BleveIndexer) Index(ctx context.Context, files []parser.FileInfo, indexPath string) error {
	// Close previous index if any
	if b.index != nil {
		if err := b.index.Close(); err != nil {
			b.logger.Warn("failed to close previous index", "error", err)
		}
		b.index = nil
	}

	// Remove old index directory
	if err := os.RemoveAll(indexPath); err != nil {
		return fmt.Errorf("remove old index: %w", err)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return fmt.Errorf("create index parent dir: %w", err)
	}

	// Create new bleve index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		return fmt.Errorf("create bleve index: %w", err)
	}

	// Track whether indexing succeeded to decide cleanup
	success := false
	defer func() {
		if !success {
			index.Close()
		}
	}()

	// Batch index files
	batch := index.NewBatch()
	batchCount := 0
	batchSize := b.cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for _, f := range files {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		doc := fileDocument{
			Path:     f.Path,
			Name:     f.Name,
			Language: f.Language,
			Content:  string(f.Content),
		}

		if err := batch.Index(f.Path, doc); err != nil {
			b.logger.Warn("failed to index file", "path", f.Path, "error", err)
			continue
		}

		batchCount++
		if batchCount >= batchSize {
			if err := index.Batch(batch); err != nil {
				return fmt.Errorf("batch index: %w", err)
			}
			batch = index.NewBatch()
			batchCount = 0
		}
	}

	// Flush remaining batch
	if batchCount > 0 {
		if err := index.Batch(batch); err != nil {
			return fmt.Errorf("batch index final: %w", err)
		}
	}

	success = true
	b.index = index
	b.logger.Info("indexing complete", "files", len(files), "path", indexPath)
	return nil
}

func (b *BleveIndexer) Search(ctx context.Context, query string, limit int) ([]models.SearchResult, error) {
	if b.index == nil {
		return nil, nil
	}

	if limit <= 0 {
		limit = 20
	}

	searchRequest := bleve.NewSearchRequestOptions(
		bleve.NewMatchQuery(query),
		limit, 0, false,
	)
	searchRequest.Fields = []string{"path", "name", "language", "content"}
	searchRequest.Highlight = bleve.NewHighlightWithStyle("html")

	searchResult, err := b.index.SearchInContext(ctx, searchRequest)
	if err != nil {
		return nil, fmt.Errorf("bleve search: %w", err)
	}

	var results []models.SearchResult
	for _, hit := range searchResult.Hits {
		filePath := hit.ID
		fileName := ""
		if name, ok := hit.Fields["name"].(string); ok {
			fileName = name
		}

		var matches []models.MatchFragment
		if fragments, ok := hit.Fragments["content"]; ok {
			for i, frag := range fragments {
				matches = append(matches, models.MatchFragment{
					LineNumber: i + 1,
					Line:       stripHTMLTags(frag),
				})
			}
		}

		results = append(results, models.SearchResult{
			FilePath: filePath,
			FileName: fileName,
			Score:    hit.Score,
			Matches:  matches,
		})
	}

	return results, nil
}

func (b *BleveIndexer) Close() error {
	if b.index != nil {
		err := b.index.Close()
		b.index = nil
		return err
	}
	return nil
}

// stripHTMLTags removes simple HTML tags from bleve highlight output.
func stripHTMLTags(s string) string {
	s = strings.ReplaceAll(s, "<mark>", "")
	s = strings.ReplaceAll(s, "</mark>", "")
	return s
}
