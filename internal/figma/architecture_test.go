package figma

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generatedAPIPath is the package import path whose leakage we forbid.
// Generated OpenAPI types live here; upstream schema regeneration must not
// force changes outside the Figma adapter.
const generatedAPIPath = "github.com/cristianoliveira/figma-cli/internal/figma/api"

// TestNoGeneratedAPIImportsOutsideBoundary fails the build when any
// production (non-test) Go file outside internal/figma/** imports the
// generated OpenAPI types. Generated code is regenerated from the OpenAPI
// schema, so it must not leak past the adapter boundary.
//
// Test files anywhere (and production files inside internal/figma/**) are
// exempt — tests routinely construct API responses directly, and the
// adapter itself is the legitimate consumer.
func TestNoGeneratedAPIImportsOutsideBoundary(t *testing.T) {
	root := moduleRoot(t)
	offenders := scanGeneratedAPIImports(root)

	if len(offenders) == 0 {
		return
	}

	relatives := make([]string, 0, len(offenders))
	for _, file := range offenders {
		rel, err := filepath.Rel(root, file)
		if err != nil {
			rel = file
		}
		relatives = append(relatives, rel)
	}
	t.Fatalf(
		"generated API package %q must not be imported from production code outside internal/figma/**; offenders:\n  %s",
		generatedAPIPath,
		strings.Join(relatives, "\n  "),
	)
}

// scanGeneratedAPIImports walks the repository and returns the absolute
// paths of every production Go file outside internal/figma/** that imports
// the generated API package. Test files and generated/vendored data are
// skipped.
func scanGeneratedAPIImports(root string) []string {
	var offenders []string
	boundaryPrefix := filepath.Join("internal", "figma") + string(filepath.Separator)

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shouldSkipDir(info.Name()) && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if isInsideBoundary(path, root, boundaryPrefix) {
			return nil
		}
		if fileImportsPackage(path, generatedAPIPath) {
			offenders = append(offenders, path)
		}
		return nil
	})

	return offenders
}

// isInsideBoundary reports whether the file path lives under the boundary
// directory, relative to the scan root. Paths produced by filepath.Walk can
// be absolute or relative depending on the root; both are handled.
func isInsideBoundary(path, root, boundaryPrefix string) bool {
	rel := path
	if filepath.IsAbs(path) && filepath.IsAbs(root) {
		r, err := filepath.Rel(root, path)
		if err == nil {
			rel = r
		}
	}
	rel = filepath.ToSlash(rel)
	if !strings.HasSuffix(boundaryPrefix, "/") {
		boundaryPrefix = boundaryPrefix + "/"
	}
	normalizedBoundary := filepath.ToSlash(boundaryPrefix)
	if strings.HasPrefix(rel, normalizedBoundary) {
		return true
	}
	// filepath.Rel returns "./internal/figma/x.go" on some platforms; the
	// leading dot-segment must not break the prefix check.
	return strings.HasPrefix(strings.TrimPrefix(rel, "./"), normalizedBoundary)
}

// moduleRoot walks up from the test file's directory until it finds the
// directory that contains go.mod. The architecture check scopes its scan to
// the current module, so sub-modules or vendor trees are ignored.
func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("locating module root: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no go.mod found above %s", wd)
		}
		dir = parent
	}
}

// shouldSkipDir returns true for directories whose contents should never
// be inspected for forbidden imports.
func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".gwt", "bin", "node_modules", "vendor":
		return true
	}
	return false
}

// fileImportsPackage parses a single Go source file and reports whether it
// imports the given package path. The file must be parseable on its own
// (no type resolution), which is sufficient for import-graph checks.
func fileImportsPackage(path, importPath string) bool {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return false
	}
	for _, imp := range file.Imports {
		if imp.Path == nil {
			continue
		}
		value := strings.Trim(imp.Path.Value, `"`)
		if value == importPath {
			return true
		}
	}
	return false
}
