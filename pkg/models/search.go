package models

// SearchResponse contains search results for GET /api/search.
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
}

// SearchResult represents a single matched file.
type SearchResult struct {
	FilePath string          `json:"file_path"`
	FileName string          `json:"file_name"`
	Score    float64         `json:"score"`
	Matches  []MatchFragment `json:"matches"`
}

// MatchFragment represents a highlighted code snippet.
type MatchFragment struct {
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
}
