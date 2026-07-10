package extract

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ---- formatters ----

func formatCSS(tokens []Token, prefix string) string {
	var b strings.Builder
	b.WriteString(":root {\n")
	for _, tk := range tokens {
		fmt.Fprintf(&b, "  --%s%s: %s;\n", prefix, dashPath(tk.Path), tk.Value)
	}
	b.WriteString("}\n")
	return b.String()
}

func formatJSON(tokens []Token) string {
	root := map[string]any{}
	for _, tk := range tokens {
		insertNested(root, tk.Path, map[string]any{"value": tk.Value, "type": tk.Type})
	}
	out, _ := json.MarshalIndent(root, "", "  ")
	return string(out) + "\n"
}

func insertNested(root map[string]any, path []string, leaf map[string]any) {
	current := root
	for i, seg := range path {
		key := dashPath([]string{seg})
		if i == len(path)-1 {
			current[key] = leaf
			return
		}
		next, _ := current[key].(map[string]any)
		if next == nil {
			next = map[string]any{}
			current[key] = next
		}
		current = next
	}
}

var tailwindOrder = []string{
	"colors", "fontFamily", "fontSize", "fontWeight", "lineHeight", "letterSpacing",
	"borderRadius", "spacing", "boxShadow",
}

func formatTailwind(tokens []Token) string {
	sections := map[string]map[string]string{}
	for _, tk := range tokens {
		sec := tailwindSection(tk.Category)
		if sec == "" {
			continue
		}
		if sections[sec] == nil {
			sections[sec] = map[string]string{}
		}
		sections[sec][tailwindKey(tk)] = tailwindValue(tk)
	}

	var b strings.Builder
	b.WriteString("module.exports = {\n  theme: {\n    extend: {\n")
	for _, sec := range tailwindOrder {
		items := sections[sec]
		if len(items) == 0 {
			continue
		}
		fmt.Fprintf(&b, "      %s: {\n", sec)
		keys := make([]string, 0, len(items))
		for k := range items {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(&b, "        %s: %s,\n", k, items[k])
		}
		b.WriteString("      },\n")
	}
	b.WriteString("    },\n  },\n};\n")
	return b.String()
}

func tailwindSection(category string) string {
	switch category {
	case catColor:
		return "colors"
	case catFontFamily:
		return "fontFamily"
	case catFontSize:
		return "fontSize"
	case catFontWeight:
		return "fontWeight"
	case catLineHeight:
		return "lineHeight"
	case catLetterSpacing:
		return "letterSpacing"
	case catRadius:
		return "borderRadius"
	case catSpace:
		return "spacing"
	case catShadow:
		return "boxShadow"
	}
	return ""
}

func tailwindKey(tk Token) string {
	p := tk.Path
	if len(p) > 1 {
		p = p[1:]
	}
	switch tk.Category {
	case catFontFamily, catFontSize, catFontWeight, catLineHeight, catLetterSpacing:
		if len(p) > 1 {
			p = p[:len(p)-1]
		}
	}
	key := dashPath(p)
	if key == "" {
		key = dashPath(tk.Path)
	}
	return key
}

func tailwindValue(tk Token) string {
	if tk.Category == catFontFamily {
		name := strings.Trim(tk.Value, `"`)
		return fmt.Sprintf(`["%s"]`, name)
	}
	return strconv.Quote(tk.Value)
}
