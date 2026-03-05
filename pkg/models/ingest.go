package models

// IngestRequest is the payload for POST /api/ingest.
type IngestRequest struct {
	URL    string `json:"url"`
	Branch string `json:"branch,omitempty"`
}

// IngestResponse is the response from POST /api/ingest.
type IngestResponse struct {
	RepoName     string `json:"repo_name"`
	Status       string `json:"status"` // "success" or "error"
	Message      string `json:"message,omitempty"`
	FilesIndexed int    `json:"files_indexed,omitempty"`
}

// RepoEntry represents a previously ingested repository.
type RepoEntry struct {
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	RepoName string `json:"repo_name"`
}
