package parser

import (
	"regexp"
	"strings"
)

var (
	// import module or import module1, module2
	pyImportRE = regexp.MustCompile(`(?m)^import\s+([a-zA-Z_][\w.]*)`)
	// from module import ...
	pyFromImportRE = regexp.MustCompile(`(?m)^from\s+([a-zA-Z_][\w.]*)\s+import`)
	// from .relative import ...
	pyRelativeImportRE = regexp.MustCompile(`(?m)^from\s+(\.+[a-zA-Z_][\w.]*)\s+import`)
)

// parsePythonImports extracts import references from Python source files.
func parsePythonImports(content []byte, filePath string) []string {
	src := string(content)
	seen := make(map[string]bool)

	for _, match := range pyRelativeImportRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 {
			mod := strings.TrimLeft(match[1], ".")
			if mod != "" {
				seen[mod] = true
			}
		}
	}

	for _, match := range pyFromImportRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 && !isStdPythonLib(match[1]) {
			seen[match[1]] = true
		}
	}

	for _, match := range pyImportRE.FindAllStringSubmatch(src, -1) {
		if len(match) > 1 && !isStdPythonLib(match[1]) {
			seen[match[1]] = true
		}
	}

	var deps []string
	for dep := range seen {
		// Convert dotted module to path-like reference
		deps = append(deps, strings.ReplaceAll(dep, ".", "/"))
	}
	return deps
}

func isStdPythonLib(mod string) bool {
	std := map[string]bool{
		"os": true, "sys": true, "re": true, "io": true, "json": true,
		"math": true, "time": true, "datetime": true, "collections": true,
		"functools": true, "itertools": true, "pathlib": true, "typing": true,
		"abc": true, "copy": true, "logging": true, "unittest": true,
		"subprocess": true, "threading": true, "multiprocessing": true,
		"http": true, "urllib": true, "socket": true, "email": true,
		"argparse": true, "hashlib": true, "hmac": true, "secrets": true,
		"dataclasses": true, "enum": true, "contextlib": true, "textwrap": true,
	}
	first := strings.SplitN(mod, ".", 2)[0]
	return std[first]
}
