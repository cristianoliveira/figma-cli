package extract

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

// LayoutComparison contains variants in caller order and adjacent transitions.
type LayoutComparison struct {
	Variants    []LayoutVariant    `json:"variants"`
	Transitions []LayoutTransition `json:"transitions"`
}

// LayoutVariant is one explicit responsive frame and its CSS-ready layout intent.
type LayoutVariant struct {
	ID                 string        `json:"id"`
	Name               string        `json:"name"`
	Width              float64       `json:"width,omitempty"`
	Height             float64       `json:"height,omitempty"`
	Mode               string        `json:"mode,omitempty"`
	Wrap               string        `json:"wrap,omitempty"`
	Gap                float64       `json:"gap,omitempty"`
	CounterAxisSpacing float64       `json:"counterAxisSpacing,omitempty"`
	PrimaryAxisAlign   string        `json:"primaryAxisAlign,omitempty"`
	CounterAxisAlign   string        `json:"counterAxisAlign,omitempty"`
	Padding            LayoutPadding `json:"padding"`
	CSS                LayoutCSS     `json:"css"`
}

// LayoutCSS maps explicit Figma auto-layout values to CSS declarations.
type LayoutCSS struct {
	Display        string  `json:"display,omitempty"`
	FlexDirection  string  `json:"flexDirection,omitempty"`
	FlexWrap       string  `json:"flexWrap,omitempty"`
	JustifyContent string  `json:"justifyContent,omitempty"`
	AlignItems     string  `json:"alignItems,omitempty"`
	Gap            float64 `json:"gap,omitempty"`
	RowGap         float64 `json:"rowGap,omitempty"`
}

// LayoutTransition compares adjacent variants in the caller's explicit order.
type LayoutTransition struct {
	FromID  string                 `json:"fromId"`
	ToID    string                 `json:"toId"`
	Changes []LayoutPropertyChange `json:"changes"`
}

// LayoutPropertyChange is one responsive layout difference.
type LayoutPropertyChange struct {
	Property string `json:"property"`
	From     any    `json:"from,omitempty"`
	To       any    `json:"to,omitempty"`
}

// FindLayoutVariantsByName resolves exact case-insensitive names and rejects ambiguity.
func FindLayoutVariantsByName(document any, names []string) ([]any, error) {
	variants := make([]any, 0, len(names))
	for _, name := range names {
		matches := make([]any, 0)
		findLayoutVariantsByName(document, name, &matches)
		if len(matches) == 0 {
			return nil, fmt.Errorf("layout variant name %q was not found", name)
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("layout variant name %q matched %d nodes; use explicit --id", name, len(matches))
		}
		variants = append(variants, matches[0])
	}
	return variants, nil
}

func findLayoutVariantsByName(value any, name string, matches *[]any) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}
	if strings.EqualFold(document.StringValue(object["name"]), name) {
		*matches = append(*matches, object)
	}
	children, _ := object["children"].([]any)
	for _, child := range children {
		findLayoutVariantsByName(child, name, matches)
	}
}

// CompareLayouts compares explicitly ordered frames without inferring breakpoints.
func CompareLayouts(documents []any) LayoutComparison {
	comparison := LayoutComparison{
		Variants:    make([]LayoutVariant, 0, len(documents)),
		Transitions: make([]LayoutTransition, 0, max(0, len(documents)-1)),
	}
	for _, document := range documents {
		comparison.Variants = append(comparison.Variants, layoutVariantFromValue(document))
	}
	for index := 1; index < len(comparison.Variants); index++ {
		from := comparison.Variants[index-1]
		to := comparison.Variants[index]
		comparison.Transitions = append(comparison.Transitions, LayoutTransition{
			FromID: from.ID, ToID: to.ID, Changes: changedLayoutProperties(from, to),
		})
	}
	return comparison
}

func layoutVariantFromValue(value any) LayoutVariant {
	object, _ := value.(map[string]any)
	bounds, _ := object["absoluteBoundingBox"].(map[string]any)
	variant := LayoutVariant{
		ID:                 document.StringValue(object["id"]),
		Name:               document.StringValue(object["name"]),
		Width:              document.NumberValue(bounds["width"]),
		Height:             document.NumberValue(bounds["height"]),
		Mode:               document.StringValue(object["layoutMode"]),
		Wrap:               document.StringValue(object["layoutWrap"]),
		Gap:                document.NumberValue(object["itemSpacing"]),
		CounterAxisSpacing: document.NumberValue(object["counterAxisSpacing"]),
		PrimaryAxisAlign:   document.StringValue(object["primaryAxisAlignItems"]),
		CounterAxisAlign:   document.StringValue(object["counterAxisAlignItems"]),
		Padding: LayoutPadding{
			Top: document.NumberValue(object["paddingTop"]), Right: document.NumberValue(object["paddingRight"]),
			Bottom: document.NumberValue(object["paddingBottom"]), Left: document.NumberValue(object["paddingLeft"]),
		},
	}
	variant.CSS = layoutCSS(variant)
	return variant
}

func layoutCSS(variant LayoutVariant) LayoutCSS {
	css := LayoutCSS{}
	switch variant.Mode {
	case layoutModeHorizontal:
		css.Display = "flex"
		css.FlexDirection = "row"
	case layoutModeVertical:
		css.Display = "flex"
		css.FlexDirection = "column"
	}
	if variant.Wrap == "WRAP" {
		css.FlexWrap = "wrap"
	} else if variant.Wrap != "" {
		css.FlexWrap = "nowrap"
	}
	css.JustifyContent = cssAlignment(variant.PrimaryAxisAlign)
	css.AlignItems = cssAlignment(variant.CounterAxisAlign)
	css.Gap = variant.Gap
	if variant.Wrap == "WRAP" {
		css.RowGap = variant.CounterAxisSpacing
	}
	return css
}

func cssAlignment(value string) string {
	return map[string]string{
		"MIN": "flex-start", "CENTER": "center", "MAX": "flex-end", "SPACE_BETWEEN": "space-between",
	}[value]
}

func changedLayoutProperties(from, to LayoutVariant) []LayoutPropertyChange {
	properties := []struct {
		name string
		from any
		to   any
	}{
		{"width", from.Width, to.Width}, {"height", from.Height, to.Height},
		{"mode", from.Mode, to.Mode}, {"wrap", from.Wrap, to.Wrap},
		{"gap", from.Gap, to.Gap}, {"counterAxisSpacing", from.CounterAxisSpacing, to.CounterAxisSpacing},
		{"primaryAxisAlign", from.PrimaryAxisAlign, to.PrimaryAxisAlign},
		{"counterAxisAlign", from.CounterAxisAlign, to.CounterAxisAlign},
		{"padding", from.Padding, to.Padding},
		{"css.flexDirection", from.CSS.FlexDirection, to.CSS.FlexDirection},
		{"css.flexWrap", from.CSS.FlexWrap, to.CSS.FlexWrap},
		{"css.justifyContent", from.CSS.JustifyContent, to.CSS.JustifyContent},
		{"css.alignItems", from.CSS.AlignItems, to.CSS.AlignItems},
	}
	changes := make([]LayoutPropertyChange, 0)
	for _, property := range properties {
		if !reflect.DeepEqual(property.from, property.to) {
			changes = append(changes, LayoutPropertyChange{Property: property.name, From: property.from, To: property.to})
		}
	}
	return changes
}
