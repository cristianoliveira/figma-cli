package extract

import "testing"

func TestFilterByNameComponents(t *testing.T) {
	components := []ComponentOutput{
		{Name: "Login Button"},
		{Name: "Signup Button"},
		{Name: "Footer"},
	}

	got := FilterByName(components, "button").([]ComponentOutput)

	if len(got) != 2 {
		t.Fatalf("got %d, want 2: %#v", len(got), got)
	}
}

func TestFilterByNameRaw(t *testing.T) {
	raw := []map[string]any{
		{"name": "Login"},
		{"name": "Logout"},
		{"name": "Footer"},
	}

	got := FilterByName(raw, "log").([]map[string]any)

	if len(got) != 2 {
		t.Fatalf("got %d, want 2: %#v", len(got), got)
	}
}

func TestFilterByNamePassthrough(t *testing.T) {
	// Unknown shapes pass through unchanged.
	if got := FilterByName(42, "x"); got != 42 {
		t.Errorf("passthrough = %v, want 42", got)
	}
}
