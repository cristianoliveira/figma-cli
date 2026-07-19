// Package output owns structured result rendering and filesystem artifact output.
package output

import (
	"encoding/json"
	"io"

	toon "github.com/toon-format/toon-go"
)

// Printer renders structured values in one selected format. Text and file
// artifacts remain raw by default and retain existing JSON compatibility envelopes.
type Printer struct {
	w      io.Writer
	format Format
}

func New(w io.Writer, format Format) *Printer {
	return &Printer{w: w, format: format}
}

// Structured emits a JSON-shaped domain value as TOON by default or JSON when
// explicitly selected. JSON normalization preserves existing json tags and
// omitempty behavior before TOON encoding.
func (p *Printer) Structured(value any) error {
	if p.format == FormatJSON {
		return p.JSON(value)
	}

	normalized, err := normalizeJSON(value)
	if err != nil {
		return err
	}
	encoded, err := toon.Marshal(normalized)
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	_, err = p.w.Write(encoded)
	return err
}

// JSON emits compatibility JSON with the established indentation and newline.
func (p *Printer) JSON(value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	_, err = p.w.Write(encoded)
	return err
}

func normalizeJSON(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var normalized any
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

func (p *Printer) Text(key, value string) error {
	if p.format == FormatJSON {
		return p.JSON(map[string]string{key: value})
	}
	_, err := io.WriteString(p.w, value)
	return err
}

func (p *Printer) File(path string, extra map[string]any) error {
	if p.format == FormatJSON {
		value := make(map[string]any, len(extra)+1)
		value["path"] = path
		for key, item := range extra {
			value[key] = item
		}
		return p.JSON(value)
	}
	_, err := io.WriteString(p.w, path+"\n")
	return err
}

// Render now uses the structured value for AXI output. The retained text
// argument keeps command construction independent from output migration.
func (p *Printer) Render(value any, _ string) error {
	return p.Structured(value)
}
