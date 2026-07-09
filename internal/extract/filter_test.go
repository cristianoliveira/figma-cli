package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterByNameComponents(t *testing.T) {
	components := []ComponentOutput{
		{Name: "Login Button"},
		{Name: "Signup Button"},
		{Name: "Footer"},
	}

	got := FilterByName(components, "button").([]ComponentOutput)

	assert.Len(t, got, 2)
}

func TestFilterByNameRaw(t *testing.T) {
	raw := []map[string]any{
		{"name": "Login"},
		{"name": "Logout"},
		{"name": "Footer"},
	}

	got := FilterByName(raw, "log").([]map[string]any)

	assert.Len(t, got, 2)
}

func TestFilterByNamePassthrough(t *testing.T) {
	// Unknown shapes pass through unchanged.
	assert.Equal(t, 42, FilterByName(42, "x"))
}
