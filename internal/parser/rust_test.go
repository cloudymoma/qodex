package parser

import (
	"sort"
	"testing"
)

func TestParseRustImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDeps []string
	}{
		{
			name:     "use crate",
			content:  "use crate::config::Settings;",
			wantDeps: []string{"config/Settings"},
		},
		{
			name:     "mod declaration",
			content:  "mod utils;",
			wantDeps: []string{"utils"},
		},
		{
			name: "mixed",
			content: `use crate::handler::api;
mod config;
use std::io;`,
			wantDeps: []string{"config", "handler/api"},
		},
		{
			name:     "external crate ignored",
			content:  `use std::collections::HashMap;`,
			wantDeps: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRustImports([]byte(tt.content), "main.rs")

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
