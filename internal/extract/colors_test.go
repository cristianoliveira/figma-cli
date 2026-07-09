package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func paint(typ string, r, g, b float64, visible bool) map[string]any {
	return map[string]any{
		"type":    typ,
		"visible": visible,
		"color":   map[string]any{"r": r, "g": g, "b": b, "a": 1.0},
	}
}

func TestCollectColors(t *testing.T) {
	doc := map[string]any{
		"name":  "root",
		"fills": []any{paint("SOLID", 1, 0, 0, true)},
		"children": []any{
			map[string]any{"name": "a", "fills": []any{paint("SOLID", 1, 0, 0, true), paint("SOLID", 0, 1, 0, true)}},
			map[string]any{"name": "b", "fills": []any{paint("SOLID", 0, 0, 0, true)}},  // black -> filtered
			map[string]any{"name": "c", "fills": []any{paint("SOLID", 1, 0, 0, false)}}, // hidden -> filtered
			map[string]any{"name": "d", "backgroundColor": map[string]any{"color": map[string]any{"r": 0.0, "g": 0.0, "b": 1.0, "a": 1.0}}},
		},
	}

	palette := CollectColors(doc)

	require.Len(t, palette, 3)
	assert.Equal(t, "#FF0000", palette[0].Color)
	assert.Equal(t, 2, palette[0].Count)

	// background color collected with "(bg)" suffix
	var bg *ColorEntry
	for i := range palette {
		if palette[i].Color == "#0000FF" {
			bg = &palette[i]
		}
	}
	require.NotNil(t, bg)
	require.NotEmpty(t, bg.Usage)
	assert.Equal(t, "d (bg)", bg.Usage[0])
}

func TestCollectColorsEmpty(t *testing.T) {
	palette := CollectColors(map[string]any{"id": "0:0", "name": "empty", "type": "FRAME"})

	assert.NotNil(t, palette)
	assert.Empty(t, palette)
}
