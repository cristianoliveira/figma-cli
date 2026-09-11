// Package architecture enforces package dependency direction. It is a
// test-time guardrail: production code never imports this package; the
// tests in this directory scan the repository and fail when a
// documented dependency edge is violated.
//
// The rules here are intentionally narrow and direction-focused. They
// do not impose a rigid layer taxonomy; capability-oriented packages
// remain free to depend on their documented collaborators.
package architecture

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
)

// modulePath is the Go module import path of this repository.
const modulePath = "github.com/cristianoliveira/figma-cli"

// Rule describes one forbidden dependency direction. Forbidden lists
// exact import paths. ExemptPackageDirs lists module-relative directory
// prefixes whose production files may import the forbidden paths (the
// documented edge owners and generated-code boundary).
type Rule struct {
	Name              string
	Forbidden         []string
	ExemptPackageDirs []string
}

// Rules returns the documented dependency-direction rules. These are
// mirrored in root AGENTS.md; keep the two in sync.
func Rules() []Rule {
	return []Rule{
		{
			// Generated OpenAPI types must never leak past the Figma
			// adapter. Generated code itself lives under
			// internal/figma/api.
			Name:      "generated-api-boundary",
			Forbidden: []string{modulePath + "/internal/figma/api"},
			ExemptPackageDirs: []string{
				"internal/figma",
			},
		},
		{
			// Application/domain packages own policy, not transport or
			// rendering mechanics. Framework, HTTP, filesystem,
			// environment, TOON, and generated Figma API are owned by
			// the documented edge owners listed below. Applies to
			// internal/** only; the composition root (cmd) and repo
			// scripts are outside the application layer.
			Name: "domain-infra",
			Forbidden: []string{
				"github.com/spf13/cobra",
				"net/http",
				"os",
				"io/fs",
				modulePath + "/internal/env",
				"github.com/toon-format/toon-go",
				modulePath + "/internal/figma/api",
			},
			ExemptPackageDirs: []string{
				"internal/cli",        // CLI runtime: Cobra, env, HTTP, filesystem
				"internal/figma",      // Figma transport + generated API adapter
				"internal/output",     // TOON rendering + filesystem artifacts
				"internal/assetsedge", // asset export HTTP/filesystem adapters
				"internal/artifact",   // artifact filesystem persistence adapter
				"internal/env",        // environment access owner
				"internal/components", // legacy codebase filesystem discovery
			},
		},
		{
			// The composition root (cmd) must never be imported by any
			// capability package; commands depend on internals, never
			// the reverse.
			Name:              "internal-cmd",
			Forbidden:         []string{modulePath + "/cmd", modulePath + "/cmd/"},
			ExemptPackageDirs: nil,
		},
	}
}

// Offender is a single violated dependency direction. File is a
// module-relative path; Import is the forbidden import path.
type Offender struct {
	Rule   string
	File   string
	Import string
}

// Check applies the dependency-direction rules to a map of
// module-relative source files to their direct import paths. It returns
// offenders sorted by (File, Import, Rule) for deterministic output.
func Check(files map[string][]string) []Offender {
	var offenders []Offender
	rules := Rules()
	for file, imports := range files {
		dir := packageDir(file)
		for _, rule := range rules {
			if ruleExemptsPackage(rule, dir, file) {
				continue
			}
			if !ruleAppliesToFile(rule, file) {
				continue
			}
			for _, imp := range imports {
				if ruleForbids(rule, imp) {
					offenders = append(offenders, Offender{Rule: rule.Name, File: file, Import: imp})
				}
			}
		}
	}
	sort.Slice(offenders, func(i, j int) bool {
		if offenders[i].File != offenders[j].File {
			return offenders[i].File < offenders[j].File
		}
		if offenders[i].Import != offenders[j].Import {
			return offenders[i].Import < offenders[j].Import
		}
		return offenders[i].Rule < offenders[j].Rule
	})
	return offenders
}

// packageDir returns the module-relative directory of a file, e.g.
// "internal/assets/app.go" -> "internal/assets".
func packageDir(file string) string {
	dir := filepath.ToSlash(filepath.Dir(file))
	if dir == "." {
		return ""
	}
	return dir
}

// ruleExemptsPackage reports whether the file's package directory is an
// exempt edge owner or the file is test/generated code.
func ruleExemptsPackage(rule Rule, dir, file string) bool {
	if isTestFile(file) || isGeneratedFile(file) {
		return true
	}
	for _, prefix := range rule.ExemptPackageDirs {
		if dir == prefix || strings.HasPrefix(dir, prefix+"/") {
			return true
		}
	}
	return false
}

// ruleAppliesToFile scopes a rule to the intended files. The
// internal-cmd rule only applies to files under internal/**; the
// domain-infra rule applies to internal/** as well; generated-api-boundary
// applies everywhere.
func ruleAppliesToFile(rule Rule, file string) bool {
	switch rule.Name {
	case "internal-cmd", "domain-infra":
		return strings.HasPrefix(file, "internal/")
	default:
		return true
	}
}

// ruleForbids reports whether an import path violates the rule.
func ruleForbids(rule Rule, importPath string) bool {
	for _, forbidden := range rule.Forbidden {
		if importPath == forbidden {
			return true
		}
		// The internal-cmd rule forbids cmd and its subpackages.
		if strings.HasSuffix(forbidden, "/cmd/") && strings.HasPrefix(importPath, forbidden) {
			return true
		}
	}
	return false
}

// isTestFile reports whether the path is a Go test file. Test fixtures
// may import anything; the production rule is enforced separately.
func isTestFile(file string) bool {
	return strings.HasSuffix(file, "_test.go")
}

// isGeneratedFile reports whether the path lives under the generated
// OpenAPI package. Generated code is produced by oapi-codegen and must
// not be hand-edited; it is exempt from the dependency-direction scan.
func isGeneratedFile(file string) bool {
	return strings.HasPrefix(file, "internal/figma/api/")
}

// parseImports returns the direct import paths of a Go source file,
// using the Go parser so aliases and grouped imports are handled.
func parseImports(path string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	imports := make([]string, 0, len(file.Imports))
	for _, imp := range file.Imports {
		if imp.Path != nil {
			imports = append(imports, strings.Trim(imp.Path.Value, `"`))
		}
	}
	return imports, nil
}
