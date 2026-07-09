package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func propsMap(rule CSSRule) map[string]string {
	m := map[string]string{}
	for _, p := range rule.Props {
		m[p.Key] = p.Value
	}
	return m
}

func findRule(rules []CSSRule, sel string) (CSSRule, bool) {
	for _, r := range rules {
		if r.Selector == sel {
			return r, true
		}
	}
	return CSSRule{}, false
}

func TestExtractCSSRules_VerticalLayout(t *testing.T) {
	doc := map[string]any{
		"name":                   "Customization page",
		"layoutMode":             "VERTICAL",
		"itemSpacing":            float64(27.5),
		"paddingTop":             float64(40),
		"paddingRight":           float64(32),
		"paddingBottom":          float64(16),
		"paddingLeft":            float64(32),
		"counterAxisAlignItems":  "CENTER",
		"primaryAxisAlignItems":  "MIN",
		"layoutSizingHorizontal": "FIXED",
		"absoluteBoundingBox":    map[string]any{"width": float64(575), "height": float64(400)},
	}
	rules := ExtractCSSRules(doc)
	r, ok := findRule(rules, ".customization-page")
	if !ok {
		t.Fatalf("expected rule .customization-page, got %+v", rules)
	}
	p := propsMap(r)
	expectations := map[string]string{
		"display":         "flex",
		"flex-direction":  "column",
		"gap":             "27.5px",
		"padding":         "40px 32px 16px 32px",
		"align-items":     "center",
		"justify-content": "flex-start",
		"width":           "575px",
	}
	for k, want := range expectations {
		assert.Equal(t, want, p[k], "prop %s", k)
	}
}

func TestExtractCSSRules_HorizontalLayoutOmitsDirection(t *testing.T) {
	doc := map[string]any{"name": "Row", "layoutMode": "HORIZONTAL"}
	rules := ExtractCSSRules(doc)
	r, ok := findRule(rules, ".row")
	if !ok {
		t.Fatalf("expected rule .row")
	}
	p := propsMap(r)
	assert.Equal(t, "flex", p["display"], "display")
	_, has := p["flex-direction"]
	assert.False(t, has, "flex-direction should be omitted for row")
}

func TestExtractCSSRules_PaddingShorthand(t *testing.T) {
	cases := []struct {
		name           string
		pt, pr, pb, pl float64
		want           string
	}{
		{"all equal", 16, 16, 16, 16, "16px"},
		{"pair", 16, 32, 16, 32, "16px 32px"},
		{"all different", 40, 32, 16, 32, "40px 32px 16px 32px"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := map[string]any{
				"name": "P", "layoutMode": "VERTICAL",
				"paddingTop": c.pt, "paddingRight": c.pr,
				"paddingBottom": c.pb, "paddingLeft": c.pl,
			}
			p := propsMap(mustFindRule(t, ExtractCSSRules(doc), ".p"))
			assert.Equal(t, c.want, p["padding"], "padding")
		})
	}
}

func mustFindRule(t *testing.T, rules []CSSRule, sel string) CSSRule {
	t.Helper()
	r, ok := findRule(rules, sel)
	if !ok {
		t.Fatalf("missing rule %s in %+v", sel, rules)
	}
	return r
}

func TestExtractCSSRules_BackgroundAndRadius(t *testing.T) {
	doc := map[string]any{
		"name": "Card",
		"fills": []any{
			map[string]any{"type": "SOLID", "visible": true, "color": map[string]any{"r": float64(1), "g": float64(1), "b": float64(1), "a": float64(1)}},
		},
		"cornerRadius": float64(8),
	}
	p := propsMap(mustFindRule(t, ExtractCSSRules(doc), ".card"))
	assert.Equal(t, "#FFFFFF", p["background"], "background")
	assert.Equal(t, "8px", p["border-radius"], "border-radius")
}

func TestExtractCSSRules_HiddenFillSkipped(t *testing.T) {
	doc := map[string]any{
		"name": "Hidden",
		"fills": []any{
			map[string]any{"type": "SOLID", "visible": false, "color": map[string]any{"r": 1, "g": 0, "b": 0, "a": 1}},
		},
	}
	rules := ExtractCSSRules(doc)
	assert.Empty(t, rules, "expected no rules for hidden-only fill")
}

func TestExtractCSSRules_TextNode(t *testing.T) {
	doc := map[string]any{
		"name": "Title",
		"type": "TEXT",
		"style": map[string]any{
			"fontFamily":    "Inter",
			"fontSize":      float64(14),
			"fontWeight":    float64(600),
			"lineHeightPx":  float64(20),
			"letterSpacing": float64(0.1),
		},
		"fills": []any{
			map[string]any{"type": "SOLID", "visible": true, "color": map[string]any{"r": float64(0), "g": float64(0), "b": float64(0), "a": float64(1)}},
		},
		"textAlignHorizontal": "CENTER",
	}
	p := propsMap(mustFindRule(t, ExtractCSSRules(doc), ".title"))
	expectations := map[string]string{
		"font-family":    "\"Inter\"",
		"font-size":      "14px",
		"font-weight":    "600",
		"line-height":    "20px",
		"letter-spacing": "0.1px",
		"color":          "#000000",
		"text-align":     "center",
	}
	for k, want := range expectations {
		assert.Equal(t, want, p[k], k)
	}
}

func TestExtractCSSRules_DedupClassName(t *testing.T) {
	doc := map[string]any{
		"name":       "Item",
		"layoutMode": "HORIZONTAL",
		"children": []any{
			map[string]any{"name": "Item", "layoutMode": "VERTICAL"},
		},
	}
	rules := ExtractCSSRules(doc)
	selectors := map[string]bool{}
	for _, r := range rules {
		selectors[r.Selector] = true
	}
	assert.True(t, selectors[".item"], "expected .item")
	assert.True(t, selectors[".item-2"], "expected .item-2")
}

func TestExtractCSSRules_SkipsEmpty(t *testing.T) {
	doc := map[string]any{
		"name":       "Frame",
		"layoutMode": "VERTICAL",
		"children": []any{
			map[string]any{"name": "Empty group"},
		},
	}
	rules := ExtractCSSRules(doc)
	assert.Len(t, rules, 1, "expected 1 rule (Frame only)")
}

func TestFormatCSSRules(t *testing.T) {
	rules := []CSSRule{
		{Selector: ".b", Props: []CSSProp{{Key: "z-index", Value: "1"}, {Key: "display", Value: "flex"}}},
		{Selector: ".a", Props: []CSSProp{{Key: "color", Value: "red"}}},
	}
	out := FormatCSSRules(rules)
	assert.Contains(t, out, ".a {\n  color: red;\n}")
	assert.Contains(t, out, ".b {\n  display: flex;\n  z-index: 1;\n}")
}
