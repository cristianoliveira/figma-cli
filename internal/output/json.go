package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Options configures JSON output formatting.
type Options struct {
	// Indent specifies the number of spaces to use for indentation.
	// If zero, output is compact (no whitespace).
	// If negative, defaults to 0.
	Indent int
	// Include specifies which fields to include in the output.
	// If empty, all fields are included.
	// Field names should match JSON tag names (e.g., "id", "name").
	Include []string
	// Exclude specifies which fields to exclude from the output.
	// Exclude takes precedence over Include.
	Exclude []string
}

// PrintJSON writes data as JSON to w according to opts.
// It filters fields based on Include/Exclude, and applies indentation.
// Returns an error if marshaling fails.
func PrintJSON(data interface{}, w io.Writer, opts Options) error {
	filtered, err := filterFields(data, opts.Include, opts.Exclude)
	if err != nil {
		return fmt.Errorf("filter fields: %w", err)
	}
	var bytes []byte
	if opts.Indent <= 0 {
		bytes, err = json.Marshal(filtered)
	} else {
		indent := strings.Repeat(" ", opts.Indent)
		bytes, err = json.MarshalIndent(filtered, "", indent)
	}
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = w.Write(bytes)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// PrintCompactJSON writes data as compact JSON (no whitespace) to w.
// If include is non-empty, only those fields are included.
func PrintCompactJSON(data interface{}, w io.Writer, include ...string) error {
	return PrintJSON(data, w, Options{Indent: 0, Include: include})
}

// PrintPrettyJSON writes data as pretty JSON with 2-space indent to w.
// If include is non-empty, only those fields are included.
func PrintPrettyJSON(data interface{}, w io.Writer, include ...string) error {
	return PrintJSON(data, w, Options{Indent: 2, Include: include})
}

// ValidateJSON checks whether data can be marshaled to valid JSON.
// It returns the JSON bytes if valid, otherwise an error.
func ValidateJSON(data interface{}) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if !json.Valid(bytes) {
		return nil, fmt.Errorf("produced invalid JSON")
	}
	return bytes, nil
}

// filterFields returns a filtered version of data containing only fields
// specified in include (or all if empty), excluding fields in exclude.
// It handles structs, maps, slices, and pointers recursively.
func filterFields(data interface{}, include, exclude []string) (interface{}, error) {
	if len(include) == 0 && len(exclude) == 0 {
		return data, nil
	}
	return filterValue(reflect.ValueOf(data), include, exclude), nil
}

// filterValue recursively filters a reflect.Value according to include/exclude.
// It returns a Go value suitable for JSON marshaling.
func filterValue(v reflect.Value, include, exclude []string) interface{} {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		return filterStruct(v, include, exclude)
	case reflect.Map:
		return filterMap(v, include, exclude)
	case reflect.Slice, reflect.Array:
		return filterSlice(v, include, exclude)
	default:
		return v.Interface()
	}
}

// filterStruct returns a map[string]interface{} with selected fields.
func filterStruct(v reflect.Value, include, exclude []string) map[string]interface{} {
	t := v.Type()
	result := make(map[string]interface{})
	fieldSet := make(map[string]bool)
	for _, f := range include {
		fieldSet[f] = true
	}
	excludeSet := make(map[string]bool)
	for _, f := range exclude {
		excludeSet[f] = true
	}
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		// Extract JSON field name (before first comma)
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}
		// Check exclusion
		if excludeSet[jsonName] {
			continue
		}
		// If include list is non-empty, only include specified fields
		if len(include) > 0 && !fieldSet[jsonName] {
			continue
		}
		// Recursively filter field value
		fieldValue := filterValue(v.Field(i), include, exclude)
		result[jsonName] = fieldValue
	}
	return result
}

// filterMap returns a filtered map suitable for JSON marshaling.
// For maps with string keys, include/exclude lists are applied to keys.
// For maps with other key types, include/exclude are ignored (since keys are not strings),
// but nested values are still filtered recursively.
func filterMap(v reflect.Value, include, exclude []string) interface{} {
	// Handle nil maps
	if v.IsNil() {
		return nil
	}
	keyKind := v.Type().Key().Kind()
	if keyKind != reflect.String {
		// Non‑string keys cannot be filtered by field name.
		// Convert keys to strings for JSON compatibility, include all keys.
		result := make(map[string]interface{})
		iter := v.MapRange()
		for iter.Next() {
			key := fmt.Sprint(iter.Key().Interface())
			// include/exclude ignored for non-string keys, so we always include
			filteredValue := filterValue(iter.Value(), include, exclude)
			result[key] = filteredValue
		}
		return result
	}
	// String keys: apply include/exclude filtering.
	result := make(map[string]interface{})
	fieldSet := make(map[string]bool)
	for _, f := range include {
		fieldSet[f] = true
	}
	excludeSet := make(map[string]bool)
	for _, f := range exclude {
		excludeSet[f] = true
	}
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		if excludeSet[key] {
			continue
		}
		if len(include) > 0 && !fieldSet[key] {
			continue
		}
		result[key] = filterValue(iter.Value(), include, exclude)
	}
	return result
}

// filterSlice returns a slice of filtered elements.
func filterSlice(v reflect.Value, include, exclude []string) []interface{} {
	length := v.Len()
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		elem := filterValue(v.Index(i), include, exclude)
		result = append(result, elem)
	}
	return result
}
