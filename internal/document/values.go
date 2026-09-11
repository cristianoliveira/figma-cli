// Package document owns the stable document tree model and the pure
// value-coercion helpers used by every extract capability while
// walking that tree.
//
// No Figma-shape coupling: only `any` -> primitive conversions used
// while walking the raw document tree.
package document

// StringValue extracts a string from an any value, or returns "".
func StringValue(value any) string {
	text, _ := value.(string)
	return text
}

// NumberValue coerces a numeric field to float64, defaulting to 0.
func NumberValue(value any) float64 {
	number, ok := value.(float64)
	if !ok {
		return 0
	}
	return number
}

// OptionalNumber returns a pointer to a numeric field, or nil if absent.
func OptionalNumber(value any) *float64 {
	number, ok := value.(float64)
	if !ok {
		return nil
	}
	return &number
}

// NumberSlice returns the numeric contents of an any-array, or nil.
func NumberSlice(value any) []float64 {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	numbers := make([]float64, 0, len(items))
	for _, item := range items {
		numbers = append(numbers, NumberValue(item))
	}
	return numbers
}

// MapValue returns the underlying map or nil.
func MapValue(value any) map[string]any {
	object, _ := value.(map[string]any)
	return object
}
