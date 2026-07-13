package extract

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

var defaultInspectTextFields = []string{"name", "type", "relativeBounds", "layout.mode", "layout.gap", "layout.paddingTop", "layout.paddingRight", "layout.paddingBottom", "layout.paddingLeft", "fills"}

// DefaultInspectTextFields returns the compact implementation-outline fields.
func DefaultInspectTextFields() []string {
	return append([]string(nil), defaultInspectTextFields...)
}

// ValidateInspectFields rejects fields outside the inspect output contract.
func ValidateInspectFields(fields []string) error {
	validFields := inspectFieldPaths()
	for _, field := range fields {
		if !validFields[field] {
			return fmt.Errorf("unknown inspect field %q", field)
		}
	}
	return nil
}

// FormatInspectText renders selected inspect properties as one indented line per node.
func FormatInspectText(nodes []InspectOutput, fields []string) (string, error) {
	if len(fields) == 0 {
		fields = defaultInspectTextFields
	}
	if err := ValidateInspectFields(fields); err != nil {
		return "", err
	}

	var output strings.Builder
	for _, node := range nodes {
		encoded, err := json.Marshal(node)
		if err != nil {
			return "", err
		}
		var values map[string]any
		if err := json.Unmarshal(encoded, &values); err != nil {
			return "", err
		}
		output.WriteString(strings.Repeat("  ", node.Depth))
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			formatted, ok := formatInspectNodeField(node, values, field)
			if !ok {
				continue
			}
			parts = append(parts, formatted)
		}
		output.WriteString(strings.Join(parts, " "))
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func inspectFieldPaths() map[string]bool {
	paths := map[string]bool{}
	collectJSONFieldPaths(reflect.TypeOf(InspectOutput{}), "", paths)
	return paths
}

func collectJSONFieldPaths(value reflect.Type, prefix string, paths map[string]bool) {
	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return
	}
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		paths[path] = true
		fieldType := field.Type
		for fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			collectJSONFieldPaths(fieldType, path, paths)
		}
	}
}

func inspectFieldValue(values map[string]any, path string) (any, bool) {
	var current any = values
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func formatInspectNodeField(node InspectOutput, values map[string]any, field string) (string, bool) {
	switch field {
	case "relativeBounds":
		if node.RelativeBounds == nil {
			return "", false
		}
		bounds := node.RelativeBounds
		return fmt.Sprintf("x:%v y:%v w:%v h:%v", bounds.X, bounds.Y, bounds.Width, bounds.Height), true
	case "bounds":
		return fmt.Sprintf("x:%v y:%v w:%v h:%v", node.Bounds.X, node.Bounds.Y, node.Bounds.Width, node.Bounds.Height), true
	}
	value, ok := inspectFieldValue(values, field)
	if !ok {
		return "", false
	}
	return formatInspectField(field, value), true
}

func formatInspectField(field string, value any) string {
	switch field {
	case "name":
		return fmt.Sprint(value)
	case "type":
		return "(" + fmt.Sprint(value) + ")"
	case "layout.mode":
		return "layout:" + fmt.Sprint(value)
	case "layout.gap":
		return "gap:" + fmt.Sprint(value)
	case "fills", "strokes":
		return field + ":" + formatInspectColors(value)
	default:
		return field + ":" + compactInspectValue(value)
	}
}

func formatInspectColors(value any) string {
	values, ok := value.([]any)
	if !ok || len(values) != 1 {
		return compactInspectValue(value)
	}
	return fmt.Sprint(values[0])
}

func compactInspectValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(encoded)
}
