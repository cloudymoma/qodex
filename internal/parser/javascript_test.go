package parser

import (
	"sort"
	"testing"
)

func TestParseJSImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDeps []string
	}{
		{
			name:     "es6 import",
			content:  `import { foo } from './utils/helper'`,
			wantDeps: []string{"utils/helper"},
		},
		{
			name:     "default import",
			content:  `import React from 'react'`,
			wantDeps: nil, // non-relative, skipped
		},
		{
			name: "require",
			content: `const fs = require('fs')
const helper = require('./helper')`,
			wantDeps: []string{"helper"},
		},
		{
			name: "mixed",
			content: `import { a } from './moduleA'
const b = require('./moduleB')
import c from 'external-lib'`,
			wantDeps: []string{"moduleA", "moduleB"},
		},
		{
			name:     "parent import",
			content:  `import { x } from '../shared/utils'`,
			wantDeps: []string{"shared/utils"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseJSImports([]byte(tt.content), "index.js")

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
