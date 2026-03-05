package handler

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"qodex/internal/service"
	"qodex/pkg/models"
)

// File handles GET /api/file?path=src/main.go.
func File(svc *service.IngestService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "query parameter 'path' is required",
			})
			return
		}

		content, err := svc.FileContent(filePath)
		if err != nil {
			logger.Error("file content failed", "error", err, "path", filePath)
			status := http.StatusInternalServerError
			if strings.Contains(err.Error(), "no repository") ||
				strings.Contains(err.Error(), "not found") {
				status = http.StatusNotFound
			}
			if strings.Contains(err.Error(), "path traversal") ||
				strings.Contains(err.Error(), "invalid path") {
				status = http.StatusBadRequest
			}
			writeJSON(w, status, map[string]string{
				"error": err.Error(),
			})
			return
		}

		lang := languageFromExt(filepath.Ext(filePath))
		resp := models.FileResponse{
			Path:     filePath,
			Content:  content,
			Language: lang,
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

// languageFromExt maps file extensions to syntax highlighter language names.
func languageFromExt(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".go":
		return "go"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".mts", ".cts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "jsx"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".css":
		return "css"
	case ".html", ".htm":
		return "html"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".md":
		return "markdown"
	case ".sh", ".bash":
		return "bash"
	case ".sql":
		return "sql"
	case ".toml":
		return "toml"
	case ".dockerfile":
		return "dockerfile"
	default:
		return "text"
	}
}
