package handler

import (
	"testing"
)

func TestLanguageFromExt(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "go"},
		{".js", "javascript"},
		{".ts", "typescript"},
		{".tsx", "tsx"},
		{".jsx", "jsx"},
		{".py", "python"},
		{".rs", "rust"},
		{".java", "java"},
		{".rb", "ruby"},
		{".c", "c"},
		{".h", "c"},
		{".cpp", "cpp"},
		{".cs", "csharp"},
		{".css", "css"},
		{".html", "html"},
		{".json", "json"},
		{".yaml", "yaml"},
		{".yml", "yaml"},
		{".xml", "xml"},
		{".md", "markdown"},
		{".sh", "bash"},
		{".sql", "sql"},
		{".toml", "toml"},
		{".unknown", "text"},
		{"", "text"},
		{".GO", "go"},      // case insensitive
		{".Py", "python"},  // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := languageFromExt(tt.ext)
			if got != tt.want {
				t.Errorf("languageFromExt(%q) = %q, want %q", tt.ext, got, tt.want)
			}
		})
	}
}
