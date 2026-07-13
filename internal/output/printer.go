// Package output encapsulates command result rendering and filesystem artifact writes.
//
// Every command produces one of three kinds of result:
//
//   - JSON:  an already-structured value (inspect, colors, find, ...). It is
//     always rendered as JSON regardless of the --json flag, since the
//     value itself is JSON-shaped.
//   - Text:  raw text such as generated CSS or a tokens file. By default it is
//     printed verbatim; under --json it is wrapped as {"<key>": "..."}.
//   - File:  a path to an asset written to disk (export, or css/tokens --output).
//     By default the path is printed on its own line; under --json it is
//     emitted as {"path": "...", <extra...>}.
//
// Centralising this keeps the --json behaviour in exactly one place and lets
// cmd/ commands stay thin: build a value, hand it to a Printer.
package output

import (
	"encoding/json"
	"io"
	"strings"
)

// Printer renders command results to a writer. The asJSON flag (the global
// --json flag) decides how Text and File results are serialised; JSON results
// are unaffected because they are already structured.
type Printer struct {
	w      io.Writer
	asJSON bool
}

// New builds a Printer bound to w. Pass the resolved --json flag as asJSON.
func New(w io.Writer, asJSON bool) *Printer {
	return &Printer{w: w, asJSON: asJSON}
}

// JSON emits an already-structured value, always as indented JSON with a
// trailing newline. The --json flag is intentionally ignored: a JSON result is
// JSON either way.
func (p *Printer) JSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = p.w.Write(b)
	return err
}

// Text emits raw text verbatim by default. Under --json it wraps the value as
// {"<key>": value}, so text results stay consumable by JSON pipelines.
func (p *Printer) Text(key, value string) error {
	if p.asJSON {
		return p.JSON(map[string]string{key: value})
	}
	_, err := io.WriteString(p.w, value)
	return err
}

// File reports a path written to disk. By default it prints the path followed
// by a newline (matching export). Under --json it emits {"path": ..., <extra...>}
// so callers can carry format/node metadata alongside the path.
func (p *Printer) File(path string, extra map[string]any) error {
	if p.asJSON {
		obj := make(map[string]any, len(extra)+1)
		obj["path"] = path
		for k, v := range extra {
			obj[k] = v
		}
		return p.JSON(obj)
	}
	_, err := io.WriteString(p.w, path+"\n")
	return err
}

// Render emits a result that has both a machine-readable JSON form (v) and a
// human-readable text form (text). Under --json it emits v as indented JSON;
// otherwise it emits text verbatim (ensuring a single trailing newline). Use
// for structured results that can render a human view, so the --json switch
// stays in the printer instead of each command.
func (p *Printer) Render(v any, text string) error {
	if p.asJSON {
		return p.JSON(v)
	}
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	_, err := io.WriteString(p.w, text)
	return err
}
