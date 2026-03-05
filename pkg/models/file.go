package models

// FileResponse is the response for GET /api/file.
type FileResponse struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Language string `json:"language"`
}
