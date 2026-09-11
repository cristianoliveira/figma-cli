package architecture

import (
	"os"
	"path/filepath"
	"strings"
)

// ScanModule walks root and returns a map of module-relative Go source
// file paths to their parsed direct import paths. It skips .git, .gwt,
// vendor, bin, node_modules, and unrelated nested modules.
//
// This function lives in a _test.go file (rather than checker.go)
// because it reads the filesystem, which the pure dependency-direction
// rule predicate must not depend on.
func ScanModule(root string) (map[string][]string, error) {
	files := map[string][]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		imports, err := parseImports(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = imports
		return nil
	})
	return files, err
}

// shouldSkipDir returns true for directories whose contents should
// never be inspected.
func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".gwt", "bin", "node_modules", "vendor":
		return true
	}
	return false
}

// ModuleRoot walks up from the current working directory until it
// finds the directory containing go.mod.
func ModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
