package graph

import (
	"testing"

	"qodex/internal/parser"
	"qodex/pkg/models"
)

func TestBuilderBuild(t *testing.T) {
	b := NewBuilder()

	files := []parser.FileInfo{
		{Path: "main.go", Name: "main.go", Size: 50, Language: "go"},
		{Path: "util.go", Name: "util.go", Size: 30, Language: "go"},
		{Path: "app.js", Name: "app.js", Size: 100, Language: "javascript"},
	}

	deps := []parser.Dependency{
		{From: "main.go", To: []string{"util.go"}},
		{From: "app.js", To: []string{"util.go"}},
	}

	g := b.Build(files, deps)

	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Links) != 2 {
		t.Errorf("expected 2 links, got %d", len(g.Links))
	}

	for _, node := range g.Nodes {
		switch node.Name {
		case "main.go", "util.go":
			if node.Group != models.GroupGo {
				t.Errorf("node %s group = %d, want %d", node.Name, node.Group, models.GroupGo)
			}
		case "app.js":
			if node.Group != models.GroupJS {
				t.Errorf("node %s group = %d, want %d", node.Name, node.Group, models.GroupJS)
			}
		}
	}
}

func TestBuilderBuildEmpty(t *testing.T) {
	b := NewBuilder()
	g := b.Build(nil, nil)

	if len(g.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.Nodes))
	}
	if len(g.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(g.Links))
	}
}

func TestBuilderFiltersDanglingLinks(t *testing.T) {
	b := NewBuilder()

	files := []parser.FileInfo{
		{Path: "src/main.rs", Name: "main.rs", Size: 10, Language: "rust"},
		{Path: "src/utils.rs", Name: "utils.rs", Size: 10, Language: "rust"},
	}

	deps := []parser.Dependency{
		{From: "src/main.rs", To: []string{
			"utils",         // should resolve to src/utils.rs
			"nonexistent",   // should be filtered out
			"std::io",       // should be filtered out
		}},
	}

	g := b.Build(files, deps)

	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(g.Nodes))
	}

	// Only "utils" -> "src/utils.rs" should resolve
	if len(g.Links) != 1 {
		t.Errorf("expected 1 link (dangling filtered), got %d", len(g.Links))
		for _, l := range g.Links {
			t.Logf("  link: %s -> %s", l.Source, l.Target)
		}
	}
}

func TestBuilderResolvesModulePaths(t *testing.T) {
	b := NewBuilder()

	files := []parser.FileInfo{
		{Path: "src/main.rs", Name: "main.rs", Size: 10, Language: "rust"},
		{Path: "src/error.rs", Name: "error.rs", Size: 10, Language: "rust"},
		{Path: "src/utils/mod.rs", Name: "mod.rs", Size: 10, Language: "rust"},
		{Path: "src/algorithms/sort.rs", Name: "sort.rs", Size: 10, Language: "rust"},
	}

	deps := []parser.Dependency{
		{From: "src/main.rs", To: []string{
			"error/Result",      // should strip "/Result" and resolve to src/error.rs
			"utils",             // should resolve to src/utils/mod.rs via module index
			"algorithms/sort",   // should resolve to src/algorithms/sort.rs
		}},
	}

	g := b.Build(files, deps)

	if len(g.Links) != 3 {
		t.Errorf("expected 3 resolved links, got %d", len(g.Links))
		for _, l := range g.Links {
			t.Logf("  link: %s -> %s", l.Source, l.Target)
		}
	}

	// Verify specific targets
	targets := make(map[string]bool)
	for _, l := range g.Links {
		targets[l.Target] = true
	}
	for _, want := range []string{"src/error.rs", "src/utils/mod.rs", "src/algorithms/sort.rs"} {
		if !targets[want] {
			t.Errorf("expected link target %q not found", want)
		}
	}
}

func TestBuilderRelativeImports(t *testing.T) {
	b := NewBuilder()

	files := []parser.FileInfo{
		{Path: "src/components/App.tsx", Name: "App.tsx", Size: 10, Language: "typescript"},
		{Path: "src/components/Button.tsx", Name: "Button.tsx", Size: 10, Language: "typescript"},
		{Path: "src/utils/helpers.ts", Name: "helpers.ts", Size: 10, Language: "typescript"},
	}

	deps := []parser.Dependency{
		{From: "src/components/App.tsx", To: []string{
			"./Button",          // relative to src/components/
			"../utils/helpers",  // relative, up one level
		}},
	}

	g := b.Build(files, deps)

	if len(g.Links) < 1 {
		t.Errorf("expected at least 1 resolved link, got %d", len(g.Links))
		for _, l := range g.Links {
			t.Logf("  link: %s -> %s", l.Source, l.Target)
		}
	}
}

func TestBuilderSkipsSelfReferences(t *testing.T) {
	b := NewBuilder()

	files := []parser.FileInfo{
		{Path: "main.go", Name: "main.go", Size: 10, Language: "go"},
	}

	deps := []parser.Dependency{
		{From: "main.go", To: []string{"main.go"}},
	}

	g := b.Build(files, deps)

	if len(g.Links) != 0 {
		t.Errorf("expected 0 links (self-ref filtered), got %d", len(g.Links))
	}
}

func TestBuildModuleIndex(t *testing.T) {
	files := []parser.FileInfo{
		{Path: "src/utils.go", Name: "utils.go"},
		{Path: "src/error/mod.rs", Name: "mod.rs"},
		{Path: "lib/__init__.py", Name: "__init__.py"},
		{Path: "src/components/index.ts", Name: "index.ts"},
	}

	idx := buildModuleIndex(files)

	tests := []struct {
		key  string
		want string
	}{
		{"src/utils", "src/utils.go"},
		{"utils", "src/utils.go"},
		{"src/error/mod", "src/error/mod.rs"},
		{"src/error", "src/error/mod.rs"},
		{"lib/__init__", "lib/__init__.py"},
		{"lib", "lib/__init__.py"},
		{"src/components/index", "src/components/index.ts"},
		{"src/components", "src/components/index.ts"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := idx[tt.key]
			if !ok {
				t.Errorf("key %q not found in module index", tt.key)
				return
			}
			if got != tt.want {
				t.Errorf("moduleIndex[%q] = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestLanguageGroup(t *testing.T) {
	tests := []struct {
		lang string
		want int
	}{
		{"go", models.GroupGo},
		{"rust", models.GroupRust},
		{"javascript", models.GroupJS},
		{"typescript", models.GroupTS},
		{"python", models.GroupPython},
		{"java", models.GroupJava},
		{"unknown", models.GroupOther},
		{"", models.GroupOther},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := languageGroup(tt.lang)
			if got != tt.want {
				t.Errorf("languageGroup(%q) = %d, want %d", tt.lang, got, tt.want)
			}
		})
	}
}
