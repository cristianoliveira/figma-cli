package extract

import (
	"strings"
	"testing"
)

// ---- Variables extraction ----

func variablesMeta(t *testing.T) map[string]any {
	t.Helper()
	return map[string]any{
		"variableCollections": map[string]any{
			"col-1": map[string]any{
				"name":          "Collection",
				"defaultModeId": "m-light",
				"variableIds":   []any{"v-color", "v-space", "v-string", "v-alias", "v-web"},
				"modes": []any{
					map[string]any{"modeId": "m-light", "name": "Light"},
					map[string]any{"modeId": "m-dark", "name": "Dark"},
				},
			},
		},
		"variables": map[string]any{
			"v-color": map[string]any{
				"name":                 "Color/Primary/500",
				"resolvedType":         "COLOR",
				"variableCollectionId": "col-1",
				"valuesByMode": map[string]any{
					"m-light": map[string]any{"r": 0.024, "g": 0.4, "b": 0.784, "a": 1.0},
					"m-dark":  map[string]any{"r": 0.2, "g": 0.6, "b": 1.0, "a": 1.0},
				},
			},
			"v-space": map[string]any{
				"name":                 "Space/4",
				"resolvedType":         "FLOAT",
				"variableCollectionId": "col-1",
				"scopes":               []any{"GAP"},
				"valuesByMode":         map[string]any{"m-light": 16.0, "m-dark": 16.0},
			},
			"v-string": map[string]any{
				"name":                 "Font Family/Body",
				"resolvedType":         "STRING",
				"variableCollectionId": "col-1",
				"valuesByMode":         map[string]any{"m-light": "Inter"},
			},
			"v-web": map[string]any{
				"name":                 "Color/Brand",
				"resolvedType":         "COLOR",
				"variableCollectionId": "col-1",
				"codeSyntax":           map[string]any{"WEB": "var(--brand)"},
				"valuesByMode":         map[string]any{"m-light": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0, "a": 1.0}},
			},
			"v-alias": map[string]any{
				"name":                 "Color/Primary",
				"resolvedType":         "COLOR",
				"variableCollectionId": "col-1",
				"valuesByMode":         map[string]any{"m-light": map[string]any{"type": "VARIABLE_ALIAS", "id": "v-color"}},
			},
		},
	}
}

func TestExtractTokensFromVariablesDefaultMode(t *testing.T) {
	tokens := ExtractTokensFromVariables(variablesMeta(t), "")

	byCSS := map[string]string{}
	for _, tk := range tokens {
		byCSS[tokenCSSName(tk, "")] = tk.Value
	}

	cases := map[string]string{
		"--color-primary-500": "#0666C8",
		"--space-4":           "16px",
		"--font-family-body":  `"Inter"`,
		"--color-brand":       "var(--brand)", // codeSyntax.WEB overrides
		"--color-primary":     "var(--color-primary-500)",
	}
	for name, want := range cases {
		got, ok := byCSS[name]
		if !ok {
			t.Errorf("missing token %s; have %v", name, byCSS)
			continue
		}
		if got != want {
			t.Errorf("token %s = %q, want %q", name, got, want)
		}
	}
}

func TestExtractTokensFromVariablesSpecificMode(t *testing.T) {
	tokens := ExtractTokensFromVariables(variablesMeta(t), "Dark")

	want := "#3399FF"
	for _, tk := range tokens {
		if tokenCSSName(tk, "") == "--color-primary-500" && tk.Value != want {
			t.Errorf("Dark mode color = %q, want %q", tk.Value, want)
		}
	}
}

func TestExtractTokensFromVariablesUnknownModeErrors(t *testing.T) {
	_, err := ExtractTokensFromVariablesE(variablesMeta(t), "Nope")
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
}

// ---- Styles extraction ----

func TestExtractTokensFromStyles(t *testing.T) {
	styles := []map[string]any{
		{"key": "s1", "name": "Background/Primary", "style_type": "FILL", "node_id": "1:1"},
		{"key": "s2", "name": "Body", "style_type": "TEXT", "node_id": "1:2"},
		{"key": "s3", "name": "Card", "style_type": "EFFECT", "node_id": "1:3"},
	}
	nodes := map[string]any{
		"1:1": map[string]any{
			"type":  "RECTANGLE",
			"fills": []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 1.0, "g": 0.4, "b": 0.0, "a": 1.0}}},
		},
		"1:2": map[string]any{
			"type":  "TEXT",
			"style": map[string]any{"fontFamily": "Inter", "fontSize": 14.0, "fontWeight": 400.0, "lineHeightPx": 20.0},
		},
		"1:3": map[string]any{
			"type": "FRAME",
			"effects": []any{map[string]any{
				"type": "DROP_SHADOW", "radius": 8.0, "spread": 0.0,
				"offset": map[string]any{"x": 0.0, "y": 2.0},
				"color":  map[string]any{"r": 0.0, "g": 0.0, "b": 0.0, "a": 0.08},
			}},
		},
	}

	tokens := ExtractTokensFromStyles(styles, nodes)
	byCSS := map[string]string{}
	for _, tk := range tokens {
		byCSS[tokenCSSName(tk, "")] = tk.Value
	}

	cases := map[string]string{
		"--color-background-primary": "#FF6600",
		"--font-body-family":         `"Inter"`,
		"--font-body-size":           "14px",
		"--font-body-weight":         "400",
		"--font-body-line-height":    "20px",
		"--shadow-card":              "0px 2px 8px 0px rgba(0, 0, 0, 0.08)",
	}
	for name, want := range cases {
		got, ok := byCSS[name]
		if !ok {
			t.Errorf("missing token %s; have %v", name, byCSS)
			continue
		}
		if got != want {
			t.Errorf("token %s = %q, want %q", name, got, want)
		}
	}
}

// ---- Formatters ----

func TestFormatCSS(t *testing.T) {
	tokens := []Token{
		{Path: []string{"color", "primary"}, Category: "color", Value: "#0666C8", Type: "color"},
		{Path: []string{"color", "bg"}, Category: "color", Value: "#FFFFFF", Type: "color"},
	}
	out, err := FormatTokens(tokens, "css", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, ":root {") || !strings.Contains(out, "--color-primary: #0666C8;") {
		t.Fatalf("css output = %q", out)
	}
}

func TestFormatCSSPrefix(t *testing.T) {
	tokens := []Token{{Path: []string{"color", "primary"}, Category: "color", Value: "#0666C8", Type: "color"}}
	out, _ := FormatTokens(tokens, "css", "fig-")
	if !strings.Contains(out, "--fig-color-primary: #0666C8;") {
		t.Fatalf("prefixed css = %q", out)
	}
}

func TestFormatCSSIdempotent(t *testing.T) {
	tokens := []Token{
		{Path: []string{"b"}, Value: "1", Type: "number"},
		{Path: []string{"a"}, Value: "2", Type: "number"},
	}
	first, _ := FormatTokens(tokens, "css", "")
	second, _ := FormatTokens(tokens, "css", "")
	if first != second {
		t.Fatal("css output not idempotent")
	}
}

func TestFormatJSON(t *testing.T) {
	tokens := []Token{
		{Path: []string{"color", "primary"}, Category: "color", Value: "#0666C8", Type: "color"},
	}
	out, err := FormatTokens(tokens, "json", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"primary": {`) || !strings.Contains(out, `"value": "#0666C8"`) || !strings.Contains(out, `"type": "color"`) {
		t.Fatalf("json output = %q", out)
	}
}

func TestFormatTailwind(t *testing.T) {
	tokens := []Token{
		{Path: []string{"color", "primary"}, Category: "color", Value: "#0666C8", Type: "color"},
		{Path: []string{"font", "body", "family"}, Category: "fontFamily", Value: `"Inter"`, Type: "typography"},
		{Path: []string{"font", "body", "size"}, Category: "fontSize", Value: "14px", Type: "typography"},
		{Path: []string{"shadow", "card"}, Category: "shadow", Value: "0px 2px 8px rgba(0,0,0,0.08)", Type: "shadow"},
	}
	out, err := FormatTokens(tokens, "tailwind", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `primary: "#0666C8"`) {
		t.Errorf("tailwind colors missing; got %q", out)
	}
	if !strings.Contains(out, `body: ["Inter"]`) && !strings.Contains(out, `body: "Inter"`) {
		t.Errorf("tailwind fontFamily missing; got %q", out)
	}
	if !strings.Contains(out, `body: "14px"`) {
		t.Errorf("tailwind fontSize missing; got %q", out)
	}
}

func TestFormatTokensUnknownFormat(t *testing.T) {
	_, err := FormatTokens(nil, "yaml", "")
	if err == nil {
		t.Fatal("expected error for unknown format")
	}
}

func TestDashPath(t *testing.T) {
	got := dashPath([]string{"Color", "Primary/500", "Sub"})
	want := "color-primary-500-sub"
	if got != want {
		t.Fatalf("dashPath = %q, want %q", got, want)
	}
}

// ---- Document scan extraction ----

func TestExtractTokensFromDocument(t *testing.T) {
	doc := map[string]any{
		"id": "root", "name": "Root", "type": "FRAME",
		"fills": []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 0.024, "g": 0.4, "b": 0.784, "a": 1.0}}},
		"effects": []any{map[string]any{
			"type": "DROP_SHADOW", "radius": 8.0, "spread": 0.0,
			"offset": map[string]any{"x": 0.0, "y": 2.0},
			"color":  map[string]any{"r": 0.0, "g": 0.0, "b": 0.0, "a": 0.08},
		}},
		"children": []any{
			map[string]any{"type": "TEXT", "style": map[string]any{"fontFamily": "Inter", "fontSize": 14.0}},
			map[string]any{"type": "TEXT", "style": map[string]any{"fontFamily": "Inter", "fontSize": 16.0}},
			map[string]any{"type": "RECTANGLE", "fills": []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 1.0, "g": 1.0, "b": 1.0, "a": 1.0}}}},
		},
	}

	byCSS := map[string]string{}
	for _, tk := range ExtractTokensFromDocument(doc) {
		byCSS[tokenCSSName(tk, "")] = tk.Value
	}

	cases := map[string]string{
		"--color-0666c8":      "#0666C8",
		"--color-ffffff":      "#FFFFFF",
		"--shadow-1":          "0px 2px 8px 0px rgba(0, 0, 0, 0.08)",
		"--font-family-inter": `"Inter"`,
		"--font-size-14":      "14px",
		"--font-size-16":      "16px",
	}
	for name, want := range cases {
		got, ok := byCSS[name]
		if !ok {
			t.Errorf("missing token %s; have %v", name, byCSS)
			continue
		}
		if got != want {
			t.Errorf("token %s = %q, want %q", name, got, want)
		}
	}
}

func TestExtractTokensFromDocumentIdempotent(t *testing.T) {
	doc := map[string]any{"fills": []any{
		map[string]any{"type": "SOLID", "color": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0, "a": 1.0}},
		map[string]any{"type": "SOLID", "color": map[string]any{"r": 0.0, "g": 1.0, "b": 0.0, "a": 1.0}},
	}}
	first, _ := FormatTokens(ExtractTokensFromDocument(doc), "css", "")
	second, _ := FormatTokens(ExtractTokensFromDocument(doc), "css", "")
	if first != second {
		t.Fatal("scan output not idempotent")
	}
}
