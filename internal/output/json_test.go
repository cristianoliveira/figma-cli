package output

import (
	"bytes"
	"testing"
)

func TestPrintJSON_MapWithIntKeys(t *testing.T) {
	// Map with integer keys should not panic when filtering
	data := map[int]interface{}{
		1: "value1",
		2: "value2",
		3: map[string]string{"nested": "val"},
	}

	var buf bytes.Buffer
	// Include list should be ignored for non-string keys
	err := PrintJSON(data, &buf, Options{Include: []string{"1"}})
	if err != nil {
		t.Fatalf("PrintJSON with int keys failed: %v", err)
	}
	// Expect full map (since include ignored)
	// Not checking exact output, just ensure no panic
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestPrintJSON_MapWithStringKeysInclude(t *testing.T) {
	data := map[string]interface{}{
		"name":  "test",
		"id":    123,
		"extra": "skip",
	}
	var buf bytes.Buffer
	err := PrintJSON(data, &buf, Options{Include: []string{"name", "id"}})
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}
	// Output should contain only name and id fields
	output := buf.String()
	if !contains(output, "name") || !contains(output, "id") {
		t.Errorf("expected output to contain 'name' and 'id', got: %s", output)
	}
	if contains(output, "extra") {
		t.Errorf("expected output to exclude 'extra', got: %s", output)
	}
}

func TestPrintJSON_MapWithIntKeysNestedStruct(t *testing.T) {
	type nested struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2"`
	}
	data := map[int]interface{}{
		1: nested{Field1: "val1", Field2: "val2"},
		2: "plain",
	}
	var buf bytes.Buffer
	// Include list ignored, but nested struct should still be filtered (field selection?)
	// Since include is empty, all fields included.
	err := PrintJSON(data, &buf, Options{})
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}
	// Ensure no panic
}

func TestPrintJSON_NilMap(t *testing.T) {
	var nilMap map[string]interface{}
	var buf bytes.Buffer
	err := PrintJSON(nilMap, &buf, Options{})
	if err != nil {
		t.Fatalf("PrintJSON with nil map failed: %v", err)
	}
	// Expect "null"
	if buf.String() != "null" {
		t.Errorf("expected 'null' for nil map, got: %s", buf.String())
	}
}

func TestPrintJSON_MapWithIntKeysIncludeExcludeIgnored(t *testing.T) {
	data := map[int]string{
		1: "val1",
		2: "val2",
		3: "val3",
	}
	var buf bytes.Buffer
	// Include only "2" but should be ignored, all keys present
	err := PrintJSON(data, &buf, Options{Include: []string{"2"}})
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}
	output := buf.String()
	// All keys should appear as strings "1","2","3"
	if !contains(output, "1") || !contains(output, "2") || !contains(output, "3") {
		t.Errorf("expected all integer keys in output, got: %s", output)
	}
	// Exclude "2" should also be ignored
	var buf2 bytes.Buffer
	err = PrintJSON(data, &buf2, Options{Exclude: []string{"2"}})
	if err != nil {
		t.Fatalf("PrintJSON with exclude failed: %v", err)
	}
	output2 := buf2.String()
	if !contains(output2, "2") {
		t.Errorf("exclude ignored: key '2' should still appear, got: %s", output2)
	}
}

func TestPrintJSON_MapWithFloatKeys(t *testing.T) {
	data := map[float64]interface{}{
		1.5: "value1",
		2.7: "value2",
	}
	var buf bytes.Buffer
	// Include a dummy field to trigger filtering (include/exclude ignored for non-string keys)
	err := PrintJSON(data, &buf, Options{Include: []string{"dummy"}})
	if err != nil {
		t.Fatalf("PrintJSON with float keys failed: %v", err)
	}
	// Should not panic, and all keys should appear (since include ignored)
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
	// Ensure both keys appear as strings "1.5" and "2.7" (fmt.Sprint representation)
	output := buf.String()
	if !contains(output, "1.5") || !contains(output, "2.7") {
		t.Errorf("expected float keys in output, got: %s", output)
	}
}

func TestPrintJSON_MapWithBoolKeys(t *testing.T) {
	data := map[bool]interface{}{
		true:  "yes",
		false: "no",
	}
	var buf bytes.Buffer
	// Include a dummy field to trigger filtering (include/exclude ignored for non-string keys)
	err := PrintJSON(data, &buf, Options{Include: []string{"dummy"}})
	if err != nil {
		t.Fatalf("PrintJSON with bool keys failed: %v", err)
	}
	// Should not panic, and all keys should appear (since include ignored)
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
	// Ensure both keys appear as strings "true" and "false"
	output := buf.String()
	if !contains(output, "true") || !contains(output, "false") {
		t.Errorf("expected bool keys in output, got: %s", output)
	}
}

func TestPrintJSON_NestedMapWithIntKeys(t *testing.T) {
	// Outer map has string keys, inner map has int keys
	data := map[string]interface{}{
		"outer": map[int]string{
			1: "inner1",
			2: "inner2",
		},
	}
	var buf bytes.Buffer
	err := PrintJSON(data, &buf, Options{})
	if err != nil {
		t.Fatalf("PrintJSON with nested int-key map failed: %v", err)
	}
	// Should not panic
}

func TestPrintJSON_ExcludePrecedenceOverInclude(t *testing.T) {
	data := map[string]interface{}{
		"name":  "test",
		"id":    123,
		"extra": "skip",
	}
	var buf bytes.Buffer
	// Include name and id, but exclude id -> id should be excluded
	err := PrintJSON(data, &buf, Options{
		Include: []string{"name", "id"},
		Exclude: []string{"id"},
	})
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}
	output := buf.String()
	if !contains(output, "name") {
		t.Errorf("expected 'name' in output, got: %s", output)
	}
	if contains(output, "id") {
		t.Errorf("expected 'id' excluded, got: %s", output)
	}
	if contains(output, "extra") {
		t.Errorf("expected 'extra' excluded, got: %s", output)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
