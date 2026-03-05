package parser

import (
	"log/slog"
	"os"
	"sort"
	"testing"
)

func TestGoParserParseImports(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	p := NewGoParser(logger)

	tests := []struct {
		name     string
		content  string
		wantDeps []string
	}{
		{
			name: "single import",
			content: `package main
import "github.com/foo/bar"
`,
			wantDeps: []string{"bar"},
		},
		{
			name: "multi import",
			content: `package main
import (
	"fmt"
	"os"
	"github.com/foo/bar"
	"github.com/baz/qux/pkg"
)
`,
			wantDeps: []string{"bar", "pkg"},
		},
		{
			name: "stdlib only",
			content: `package main
import (
	"fmt"
	"os"
	"strings"
)
`,
			wantDeps: nil,
		},
		{
			name: "no imports",
			content: `package main

func main() {}
`,
			wantDeps: nil,
		},
		{
			name: "mixed imports",
			content: `package main
import "net/http"
import (
	"context"
	"github.com/example/mylib"
)
`,
			wantDeps: []string{"mylib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ParseImports([]byte(tt.content), "main.go")

			if len(got) != len(tt.wantDeps) {
				t.Errorf("got %d deps %v, want %d deps %v", len(got), got, len(tt.wantDeps), tt.wantDeps)
				return
			}

			sort.Strings(got)
			sort.Strings(tt.wantDeps)
			for i := range got {
				if got[i] != tt.wantDeps[i] {
					t.Errorf("dep[%d] = %q, want %q", i, got[i], tt.wantDeps[i])
				}
			}
		})
	}
}

func TestIsStdLib(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"fmt", true},
		{"net/http", true},
		{"os", true},
		{"context", true},
		{"github.com/foo/bar", false},
		{"golang.org/x/tools", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isStdLib(tt.path); got != tt.want {
				t.Errorf("isStdLib(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
