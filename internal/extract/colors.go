package extract

import (
	"sort"
)

// blackColor is skipped when building a palette: it is the default "no color"
// and would otherwise dominate every result.
const blackColor = "#000000"

// ColorEntry is one color and where it appears, used by `figma colors`.
type ColorEntry struct {
	Color string   `json:"color"`
	Count int      `json:"count"`
	Usage []string `json:"usage"`
}

// CollectColors walks a document and returns its color palette, sorted by usage.
func CollectColors(value any) []ColorEntry {
	colorMap := map[string]*ColorEntry{}
	walkColors(value, colorMap)

	entries := make([]ColorEntry, 0, len(colorMap))
	for _, e := range colorMap {
		entries = append(entries, *e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if entries == nil {
		entries = []ColorEntry{}
	}
	return entries
}

func walkColors(value any, colorMap map[string]*ColorEntry) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}

	name := StringValue(object["name"])

	collectFills := func(paints any) {
		items, ok := paints.([]any)
		if !ok {
			return
		}
		for _, p := range items {
			paint, ok := p.(map[string]any)
			if !ok {
				continue
			}
			if paint["visible"] == false {
				continue
			}
			c := colorHexFromPaint(paint)
			if c == "" || c == blackColor {
				continue
			}
			entry, exists := colorMap[c]
			if !exists {
				entry = &ColorEntry{Color: c, Usage: []string{}}
				colorMap[c] = entry
			}
			entry.Count++
			if len(entry.Usage) < 3 {
				entry.Usage = append(entry.Usage, name)
			}
		}
	}

	collectFills(object["fills"])
	collectFills(object["strokes"])

	if bg, ok := object["backgroundColor"].(map[string]any); ok {
		c := colorHexFromPaint(bg)
		if c != "" && c != blackColor {
			entry, exists := colorMap[c]
			if !exists {
				entry = &ColorEntry{Color: c, Usage: []string{}}
				colorMap[c] = entry
			}
			entry.Count++
			if len(entry.Usage) < 3 {
				entry.Usage = append(entry.Usage, name+" (bg)")
			}
		}
	}

	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkColors(child, colorMap)
	}
}
