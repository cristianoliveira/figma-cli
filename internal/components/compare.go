// Package components compares published Figma components with frontend source directories.
package components

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// FigmaComponent identifies a published component in a Figma file.
type FigmaComponent struct {
	Name   string `json:"name"`
	NodeID string `json:"nodeId"`
}

// CodeComponent identifies a component conventionally represented by a source file or directory.
type CodeComponent struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Match connects a Figma component to its code counterpart.
type Match struct {
	Figma FigmaComponent `json:"figma"`
	Code  CodeComponent  `json:"code"`
}

// Comparison is the stable, machine-readable result of a component parity check.
type Comparison struct {
	Matched []Match          `json:"matched"`
	Missing []FigmaComponent `json:"missing"`
	Extra   []CodeComponent  `json:"extra"`
}

// Compare matches each Figma component's base name (before a variant separator)
// against code component names. Matching ignores case, whitespace, and punctuation.
func Compare(figmaComponents []FigmaComponent, codeComponents []CodeComponent) Comparison {
	byName := make(map[string]CodeComponent, len(codeComponents))
	for _, component := range codeComponents {
		byName[canonicalName(component.Name)] = component
	}

	comparison := Comparison{
		Matched: make([]Match, 0),
		Missing: make([]FigmaComponent, 0),
		Extra:   make([]CodeComponent, 0),
	}
	matchedCode := make(map[string]bool)
	for _, component := range figmaComponents {
		name := canonicalName(baseName(component.Name))
		code, exists := byName[name]
		if !exists {
			comparison.Missing = append(comparison.Missing, component)
			continue
		}
		comparison.Matched = append(comparison.Matched, Match{Figma: component, Code: code})
		matchedCode[name] = true
	}
	for _, component := range codeComponents {
		if !matchedCode[canonicalName(component.Name)] {
			comparison.Extra = append(comparison.Extra, component)
		}
	}
	return comparison
}

// DiscoverCodeComponents finds immediate component directories and source files.
// A directory represents one component; source files at the codebase root represent
// standalone components. Nested files are intentionally ignored because their parent
// directory already supplies the component name.
func DiscoverCodeComponents(codebase string) ([]CodeComponent, error) {
	entries, err := os.ReadDir(codebase)
	if err != nil {
		return nil, fmt.Errorf("reading codebase %q: %w", codebase, err)
	}

	components := make([]CodeComponent, 0)
	for _, entry := range entries {
		path := filepath.Join(codebase, entry.Name())
		if entry.IsDir() {
			components = append(components, CodeComponent{Name: entry.Name(), Path: path})
			continue
		}
		if sourceExtension(filepath.Ext(entry.Name())) {
			components = append(components, CodeComponent{Name: strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())), Path: path})
		}
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	return components, nil
}

func baseName(name string) string {
	return strings.TrimSpace(strings.SplitN(name, "/", 2)[0])
}

func canonicalName(name string) string {
	var normalized strings.Builder
	for _, character := range strings.ToLower(name) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			normalized.WriteRune(character)
		}
	}
	return normalized.String()
}

func sourceExtension(extension string) bool {
	switch extension {
	case ".js", ".jsx", ".svelte", ".ts", ".tsx", ".vue":
		return true
	default:
		return false
	}
}
