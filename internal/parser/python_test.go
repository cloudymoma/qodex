package parser

import (
	"sort"
	"testing"
)

func TestParsePythonImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDeps []string
	}{
		{
			name:     "simple import",
			content:  "import mypackage",
			wantDeps: []string{"mypackage"},
		},
		{
			name:     "from import",
			content:  "from mypackage.submodule import foo",
			wantDeps: []string{"mypackage/submodule"},
		},
		{
			name:     "relative import",
			content:  "from .utils import helper",
			wantDeps: []string{"utils"},
		},
		{
			name:     "stdlib filtered",
			content:  "import os\nimport json\nimport sys",
			wantDeps: nil,
		},
		{
			name: "mixed",
			content: `import os
import mylib
from mylib.sub import thing
from .local import util`,
			wantDeps: []string{"local", "mylib", "mylib/sub"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePythonImports([]byte(tt.content), "main.py")

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
