package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSON_AlwaysMarshalsWithTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	if err := New(&buf, false).JSON(map[string]int{"a": 1}); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var got map[string]int
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, buf.String())
	}
	if got["a"] != 1 {
		t.Fatalf("got %v", got)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatalf("missing trailing newline: %q", buf.String())
	}
}

func TestJSON_UnaffectedByJSONFlag(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	// asJSON=true must not wrap a JSON result.
	if err := New(&buf, true).JSON(map[string]int{"a": 1}); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var got map[string]int
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("asJSON wrapped a JSON result: %v\n%s", err, buf.String())
	}
}

func TestText_RawByDefault(t *testing.T) {
	var buf bytes.Buffer
	if err := New(&buf, false).Text("css", ".a { color: red; }"); err != nil {
		t.Fatalf("Text: %v", err)
	}
	if buf.String() != ".a { color: red; }" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestText_WrappedUnderJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := New(&buf, true).Text("css", ".a { color: red; }"); err != nil {
		t.Fatalf("Text: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("not wrapped JSON: %v\n%s", err, buf.String())
	}
	if got["css"] != ".a { color: red; }" {
		t.Fatalf("got %v", got)
	}
}

func TestFile_RawPrintsPathLine(t *testing.T) {
	var buf bytes.Buffer
	if err := New(&buf, false).File("/tmp/a.svg", nil); err != nil {
		t.Fatalf("File: %v", err)
	}
	if buf.String() != "/tmp/a.svg\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFile_JSONMergesExtra(t *testing.T) {
	var buf bytes.Buffer
	extra := map[string]any{"format": "svg", "node": "1:2"}
	if err := New(&buf, true).File("/tmp/a.svg", extra); err != nil {
		t.Fatalf("File: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, buf.String())
	}
	if got["path"] != "/tmp/a.svg" || got["format"] != "svg" || got["node"] != "1:2" {
		t.Fatalf("got %v", got)
	}
}
