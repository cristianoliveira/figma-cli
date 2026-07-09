package extract

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// CSSRule is one CSS selector and its declared properties.
type CSSRule struct {
	Selector string
	Props    []CSSProp
}

// CSSProp is a single CSS declaration.
type CSSProp struct {
	Key   string
	Value string
}

// ExtractCSSRules walks a Figma document subtree and returns one CSS rule per
// node that contributes at least one meaningful visual/layout property.
// Output is deterministic: properties are sorted within each rule.
func ExtractCSSRules(doc any) []CSSRule {
	seen := map[string]int{}
	var rules []CSSRule
	walkCSS(doc, seen, &rules)
	return rules
}

func walkCSS(value any, seen map[string]int, rules *[]CSSRule) {
	node, ok := value.(map[string]any)
	if !ok {
		return
	}
	props := cssPropsFor(node)
	if len(props) > 0 {
		name := StringValue(node["name"])
		*rules = append(*rules, CSSRule{
			Selector: "." + cssClassName(name, seen),
			Props:    sortCSSProps(props),
		})
	}
	if children, ok := node["children"].([]any); ok {
		for _, child := range children {
			walkCSS(child, seen, rules)
		}
	}
}

// cssClassName builds a slug from a node name, de-duplicating repeats.
func cssClassName(name string, seen map[string]int) string {
	slug := dashPath([]string{name})
	if slug == "" {
		slug = "node"
	}
	n := seen[slug]
	seen[slug] = n + 1
	if n == 0 {
		return slug
	}
	return fmt.Sprintf("%s-%d", slug, n+1)
}

func cssPropsFor(node map[string]any) []CSSProp {
	var props []CSSProp
	props = append(props, layoutProps(node)...)
	props = append(props, backgroundProp(node)...)
	props = append(props, borderProps(node)...)
	props = append(props, radiusProp(node)...)
	if StringValue(node["type"]) == "TEXT" {
		props = append(props, textProps(node)...)
	}
	return props
}

func layoutProps(node map[string]any) []CSSProp {
	var props []CSSProp
	switch StringValue(node["layoutMode"]) {
	case "VERTICAL":
		props = append(props, CSSProp{"display", "flex"}, CSSProp{"flex-direction", "column"})
	case "HORIZONTAL":
		props = append(props, CSSProp{"display", "flex"})
	}
	if gap := numberValue(node["itemSpacing"]); gap > 0 {
		props = append(props, CSSProp{"gap", px(gap)})
	}
	if p := paddingProp(node); p != nil {
		props = append(props, *p)
	}
	if v := alignValue(StringValue(node["counterAxisAlignItems"])); v != "" {
		props = append(props, CSSProp{"align-items", v})
	}
	if v := alignValue(StringValue(node["primaryAxisAlignItems"])); v != "" {
		props = append(props, CSSProp{"justify-content", v})
	}
	if w := sizingProp(node, "layoutSizingHorizontal", "width"); w != nil {
		props = append(props, *w)
	}
	if h := sizingProp(node, "layoutSizingVertical", "height"); h != nil {
		props = append(props, *h)
	}
	if numberValue(node["layoutGrow"]) == 1 {
		props = append(props, CSSProp{"flex-grow", "1"})
	}
	return props
}

// paddingProp collapses four paddings into the shortest valid CSS shorthand.
func paddingProp(node map[string]any) *CSSProp {
	top := numberValue(node["paddingTop"])
	right := numberValue(node["paddingRight"])
	bottom := numberValue(node["paddingBottom"])
	left := numberValue(node["paddingLeft"])
	if top == 0 && right == 0 && bottom == 0 && left == 0 {
		return nil
	}
	switch {
	case top == bottom && left == right && top == left:
		return &CSSProp{"padding", px(top)}
	case top == bottom && left == right:
		return &CSSProp{"padding", px(top) + " " + px(left)}
	default:
		return &CSSProp{"padding", px(top) + " " + px(right) + " " + px(bottom) + " " + px(left)}
	}
}

// sizingProp maps a Figma sizing mode to a width/height declaration.
func sizingProp(node map[string]any, modeKey, cssKey string) *CSSProp {
	switch StringValue(node[modeKey]) {
	case "FIXED":
		box := mapValue(node["absoluteBoundingBox"])
		if cssKey == "width" {
			if w := numberValue(box["width"]); w > 0 {
				return &CSSProp{cssKey, px(w)}
			}
		} else {
			if h := numberValue(box["height"]); h > 0 {
				return &CSSProp{cssKey, px(h)}
			}
		}
	case "FILL":
		return &CSSProp{cssKey, "100%"}
	}
	return nil
}

func backgroundProp(node map[string]any) []CSSProp {
	if c := firstSolidColor(node["fills"]); c != "" {
		return []CSSProp{{"background", c}}
	}
	return nil
}

func borderProps(node map[string]any) []CSSProp {
	c := firstSolidColor(node["strokes"])
	if c == "" {
		return nil
	}
	weight := numberValue(node["strokeWeight"])
	if weight == 0 {
		weight = 1
	}
	return []CSSProp{{"border", px(weight) + " solid " + c}}
}

func radiusProp(node map[string]any) []CSSProp {
	if r := numberValue(node["cornerRadius"]); r > 0 {
		return []CSSProp{{"border-radius", px(r)}}
	}
	if radii := numberSlice(node["rectangleCornerRadii"]); len(radii) == 4 {
		return []CSSProp{{"border-radius", px(radii[0]) + " " + px(radii[1]) + " " + px(radii[2]) + " " + px(radii[3])}}
	}
	return nil
}

func textProps(node map[string]any) []CSSProp {
	style := mapValue(node["style"])
	var props []CSSProp
	if fam := StringValue(style["fontFamily"]); fam != "" {
		props = append(props, CSSProp{"font-family", strconv.Quote(fam)})
	}
	if sz := numberValue(style["fontSize"]); sz > 0 {
		props = append(props, CSSProp{"font-size", px(sz)})
	}
	if w := numberValue(style["fontWeight"]); w > 0 {
		props = append(props, CSSProp{"font-weight", strconv.FormatFloat(w, 'f', -1, 64)})
	}
	if lh := numberValue(style["lineHeightPx"]); lh > 0 {
		props = append(props, CSSProp{"line-height", px(lh)})
	}
	if ls := numberValue(style["letterSpacing"]); ls != 0 {
		props = append(props, CSSProp{"letter-spacing", px(ls)})
	}
	if c := firstSolidColor(node["fills"]); c != "" {
		props = append(props, CSSProp{"color", c})
	}
	if v := textAlign(StringValue(node["textAlignHorizontal"])); v != "" {
		props = append(props, CSSProp{"text-align", v})
	}
	if v := textTransform(StringValue(node["textCase"])); v != "" {
		props = append(props, CSSProp{"text-transform", v})
	}
	return props
}

// alignValue maps Figma axis-alignment enums to CSS flex values.
func alignValue(figma string) string {
	switch figma {
	case "MIN":
		return "flex-start"
	case "CENTER":
		return "center"
	case "MAX":
		return "flex-end"
	case "SPACE_BETWEEN":
		return "space-between"
	}
	return ""
}

func textAlign(figma string) string {
	switch figma {
	case "LEFT":
		return "left"
	case "CENTER":
		return "center"
	case "RIGHT":
		return "right"
	case "JUSTIFIED":
		return "justify"
	}
	return ""
}

func textTransform(figma string) string {
	switch figma {
	case "UPPER":
		return "uppercase"
	case "LOWER":
		return "lowercase"
	case "TITLE":
		return "capitalize"
	}
	return ""
}

func sortCSSProps(props []CSSProp) []CSSProp {
	out := make([]CSSProp, len(props))
	copy(out, props)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

// FormatCSSRules renders rules as a stylesheet. Rules keep document order;
// properties within each rule are sorted for deterministic output.
func FormatCSSRules(rules []CSSRule) string {
	var b strings.Builder
	for _, r := range rules {
		props := sortCSSProps(r.Props)
		b.WriteString(r.Selector)
		b.WriteString(" {\n")
		for _, p := range props {
			b.WriteString("  ")
			b.WriteString(p.Key)
			b.WriteString(": ")
			b.WriteString(p.Value)
			b.WriteString(";\n")
		}
		b.WriteString("}\n")
	}
	return b.String()
}
