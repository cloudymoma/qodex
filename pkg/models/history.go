package models

// HistoryResponse is the response from GET /api/history.
type HistoryResponse struct {
	RepoURL string        `json:"repo_url"`
	Commits []CommitEntry `json:"commits"`
}

// CommitEntry represents a single git commit.
type CommitEntry struct {
	Hash         string   `json:"hash"`
	Short        string   `json:"short"`
	Message      string   `json:"message"`
	Author       string   `json:"author"`
	Date         string   `json:"date"`
	FilesChanged []string `json:"files_changed,omitempty"`
}
