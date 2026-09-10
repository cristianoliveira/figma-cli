// Package document owns the stable document model consumed by extract
// capabilities. The model is intentionally narrow: it represents only
// the fields required by the find pilot (and any future capability that
// can be served from the same shape). The model does not import
// generated Figma types, HTTP, Cobra, filesystem, or output packages.
// Mapping from Figma API values to this model happens inside
// internal/figma.
package document

import "fmt"

// Node is the stable representation of a Figma document node.
//
// Fields are limited to what the find pilot (and any capability built on
// top of it) needs. Children preserve source order. An empty Children
// slice is non-nil so JSON callers can rely on the array being present.
type Node struct {
	ID       string
	Name     string
	Type     string
	Text     string // populated only when Type is a text node
	Children []*Node
}

// MalformedError is returned by the adapter mapper when a required field
// is missing, has the wrong type, or contains malformed data. It carries
// enough node/path context for callers to surface a precise diagnostic.
//
// Path uses slash-separated segment indices ("0/2/1") so logs and CLI
// output stay stable and easy to copy/paste.
type MalformedError struct {
	Path   string
	Field  string
	Reason string
}

func (e *MalformedError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("malformed document: %s %s (%s)", e.Field, "required", e.Reason)
	}
	return fmt.Sprintf("malformed document at %s: %s %s (%s)", e.Path, e.Field, "required", e.Reason)
}
