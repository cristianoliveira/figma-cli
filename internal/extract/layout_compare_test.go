package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareLayoutsReportsExplicitVariantTransitions(t *testing.T) {
	desktop := map[string]any{
		"id": "1:1", "name": "Desktop", "type": "FRAME", "layoutMode": "HORIZONTAL", "layoutWrap": "WRAP",
		"itemSpacing": 24.0, "counterAxisSpacing": 32.0, "primaryAxisAlignItems": "SPACE_BETWEEN", "counterAxisAlignItems": "CENTER",
		"paddingTop": 24.0, "paddingRight": 24.0, "paddingBottom": 24.0, "paddingLeft": 24.0,
		"absoluteBoundingBox": map[string]any{"width": 1440.0, "height": 900.0},
	}
	mobile := map[string]any{
		"id": "2:1", "name": "Mobile", "type": "FRAME", "layoutMode": "VERTICAL", "layoutWrap": "NO_WRAP",
		"itemSpacing": 12.0, "primaryAxisAlignItems": "MIN", "counterAxisAlignItems": "MIN",
		"paddingTop": 16.0, "paddingRight": 16.0, "paddingBottom": 16.0, "paddingLeft": 16.0,
		"absoluteBoundingBox": map[string]any{"width": 375.0, "height": 800.0},
	}

	comparison := CompareLayouts([]any{desktop, mobile})

	require.Len(t, comparison.Variants, 2)
	assert.Equal(t, LayoutVariant{
		ID: "1:1", Name: "Desktop", Width: 1440, Height: 900, Mode: "HORIZONTAL", Wrap: "WRAP", Gap: 24,
		CounterAxisSpacing: 32, PrimaryAxisAlign: "SPACE_BETWEEN", CounterAxisAlign: "CENTER",
		Padding: LayoutPadding{Top: 24, Right: 24, Bottom: 24, Left: 24},
		CSS:     LayoutCSS{Display: "flex", FlexDirection: "row", FlexWrap: "wrap", JustifyContent: "space-between", AlignItems: "center", Gap: 24, RowGap: 32},
	}, comparison.Variants[0])
	require.Len(t, comparison.Transitions, 1)
	assert.Equal(t, "1:1", comparison.Transitions[0].FromID)
	assert.Equal(t, "2:1", comparison.Transitions[0].ToID)
	assert.Contains(t, comparison.Transitions[0].Changes, LayoutPropertyChange{Property: "mode", From: "HORIZONTAL", To: "VERTICAL"})
	assert.Contains(t, comparison.Transitions[0].Changes, LayoutPropertyChange{Property: "width", From: float64(1440), To: float64(375)})
	assert.Contains(t, comparison.Transitions[0].Changes, LayoutPropertyChange{Property: "css.flexDirection", From: "row", To: "column"})
}

func TestCompareLayoutsPreservesCallerOrder(t *testing.T) {
	comparison := CompareLayouts([]any{
		map[string]any{"id": "tablet", "name": "Tablet"},
		map[string]any{"id": "mobile", "name": "Mobile"},
		map[string]any{"id": "desktop", "name": "Desktop"},
	})

	assert.Equal(t, []string{"tablet", "mobile", "desktop"}, []string{
		comparison.Variants[0].ID, comparison.Variants[1].ID, comparison.Variants[2].ID,
	})
	assert.Len(t, comparison.Transitions, 2)
}
