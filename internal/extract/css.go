package extract

import (
	"fmt"
	"math"
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

// ExtractCSSRules returns CSS for the root node and, when recursive is true,
// every descendant that contributes meaningful visual/layout properties.
// Recursive defaults to true for callers that omit it.
func ExtractCSSRules(doc any, recursive ...bool) []CSSRule {
	includeDescendants := len(recursive) == 0 || recursive[0]
	seen := map[string]int{}
	var rules []CSSRule
	walkCSS(doc, includeDescendants, seen, &rules)
	return rules
}

func walkCSS(value any, recursive bool, seen map[string]int, rules *[]CSSRule) {
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
	if !recursive {
		return
	}
	if children, ok := node["children"].([]any); ok {
		for _, child := range children {
			walkCSS(child, true, seen, rules)
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
	props = append(props, opacityProp(node)...)
	props = append(props, clippingProp(node)...)
	props = append(props, effectProps(node)...)
	if StringValue(node["type"]) == "TEXT" {
		props = append(props, textProps(node)...)
	}
	return props
}

func layoutProps(node map[string]any) []CSSProp {
	var props []CSSProp
	switch StringValue(node["layoutMode"]) {
	case layoutModeVertical:
		props = append(props, CSSProp{"display", "flex"}, CSSProp{"flex-direction", "column"})
	case layoutModeHorizontal:
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
	fills, _ := node["fills"].([]any)
	for _, value := range fills {
		paint, _ := value.(map[string]any)
		if paint == nil || paint["visible"] == false {
			continue
		}
		switch StringValue(paint["type"]) {
		case paintTypeSolid:
			if color, ok := paint["color"].(map[string]any); ok {
				return []CSSProp{{"background", cssPaintColor(color, paintOpacity(paint))}}
			}
		case "GRADIENT_LINEAR":
			if gradient := linearGradient(paint); gradient != "" {
				return []CSSProp{{"background", gradient}}
			}
		}
	}
	return nil
}

func linearGradient(paint map[string]any) string {
	stops, _ := paint["gradientStops"].([]any)
	if len(stops) == 0 {
		return ""
	}
	formattedStops := make([]string, 0, len(stops))
	for _, value := range stops {
		stop, _ := value.(map[string]any)
		color, _ := stop["color"].(map[string]any)
		if color == nil {
			continue
		}
		position := numStr(roundTo(numberValue(stop["position"])*100, 2)) + "%"
		formattedStops = append(formattedStops, cssPaintColor(color, paintOpacity(paint))+" "+position)
	}
	if len(formattedStops) == 0 {
		return ""
	}
	return "linear-gradient(" + gradientAngle(paint["gradientHandlePositions"]) + "deg, " + strings.Join(formattedStops, ", ") + ")"
}

func gradientAngle(value any) string {
	handles, _ := value.([]any)
	if len(handles) < 2 {
		return "180"
	}
	start, _ := handles[0].(map[string]any)
	end, _ := handles[1].(map[string]any)
	dx := numberValue(end["x"]) - numberValue(start["x"])
	dy := numberValue(end["y"]) - numberValue(start["y"])
	angle := math.Atan2(dx, -dy) * 180 / math.Pi
	if angle < 0 {
		angle += 360
	}
	return numStr(roundTo(angle, 2))
}

func cssPaintColor(color map[string]any, paintAlpha float64) string {
	colorAlpha := 1.0
	if alpha, exists := color["a"]; exists {
		colorAlpha = numberValue(alpha)
	}
	alpha := roundTo(colorAlpha*paintAlpha, 4)
	if alpha >= 1 {
		return fmt.Sprintf("#%02X%02X%02X", colorChannel(color["r"]), colorChannel(color["g"]), colorChannel(color["b"]))
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %s)", colorChannel(color["r"]), colorChannel(color["g"]), colorChannel(color["b"]), numStr(alpha))
}

func paintOpacity(paint map[string]any) float64 {
	if opacity, exists := paint["opacity"]; exists {
		return numberValue(opacity)
	}
	return 1
}

func firstCSSSolidColor(value any) string {
	paints, _ := value.([]any)
	for _, value := range paints {
		paint, _ := value.(map[string]any)
		if paint == nil || paint["visible"] == false || StringValue(paint["type"]) != paintTypeSolid {
			continue
		}
		if color, ok := paint["color"].(map[string]any); ok {
			return cssPaintColor(color, paintOpacity(paint))
		}
	}
	return ""
}

func borderProps(node map[string]any) []CSSProp {
	c := firstCSSSolidColor(node["strokes"])
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

func effectProps(node map[string]any) []CSSProp {
	effects, _ := node["effects"].([]any)
	shadows := make([]string, 0)
	var layerBlur string
	var backgroundBlur string
	for _, value := range effects {
		effect, _ := value.(map[string]any)
		if effect == nil || effect["visible"] == false {
			continue
		}
		switch StringValue(effect["type"]) {
		case "DROP_SHADOW":
			if shadow := cssShadow(effect, false); shadow != "" {
				shadows = append(shadows, shadow)
			}
		case "INNER_SHADOW":
			if shadow := cssShadow(effect, true); shadow != "" {
				shadows = append(shadows, shadow)
			}
		case "LAYER_BLUR":
			if radius := numberValue(effect["radius"]); radius > 0 && layerBlur == "" {
				layerBlur = "blur(" + px(radius) + ")"
			}
		case "BACKGROUND_BLUR":
			if radius := numberValue(effect["radius"]); radius > 0 && backgroundBlur == "" {
				backgroundBlur = "blur(" + px(radius) + ")"
			}
		}
	}
	props := make([]CSSProp, 0, 3)
	if len(shadows) > 0 {
		props = append(props, CSSProp{"box-shadow", strings.Join(shadows, ", ")})
	}
	if layerBlur != "" {
		props = append(props, CSSProp{"filter", layerBlur})
	}
	if backgroundBlur != "" {
		props = append(props, CSSProp{"backdrop-filter", backgroundBlur})
	}
	return props
}

func cssShadow(effect map[string]any, inset bool) string {
	color, _ := effect["color"].(map[string]any)
	if color == nil {
		return ""
	}
	offset, _ := effect["offset"].(map[string]any)
	parts := []string{
		px(numberValue(offset["x"])),
		px(numberValue(offset["y"])),
		px(numberValue(effect["radius"])),
		px(numberValue(effect["spread"])),
		cssPaintColor(color, 1),
	}
	shadow := strings.Join(parts, " ")
	if inset {
		return "inset " + shadow
	}
	return shadow
}

func opacityProp(node map[string]any) []CSSProp {
	opacity := optionalNumber(node["opacity"])
	if opacity == nil || *opacity == 1 {
		return nil
	}
	return []CSSProp{{"opacity", strconv.FormatFloat(*opacity, 'f', -1, 64)}}
}

func clippingProp(node map[string]any) []CSSProp {
	if clips, _ := node["clipsContent"].(bool); clips {
		return []CSSProp{{"overflow", "hidden"}}
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
	if c := firstCSSSolidColor(node["fills"]); c != "" {
		props = append(props, CSSProp{"color", c})
	}
	if v := textAlign(StringValue(node["textAlignHorizontal"])); v != "" {
		props = append(props, CSSProp{"text-align", v})
	}
	textCase := StringValue(style["textCase"])
	if textCase == "" {
		textCase = StringValue(node["textCase"])
	}
	if v := textTransform(textCase); v != "" {
		props = append(props, CSSProp{"text-transform", v})
	}
	if v := textDecoration(StringValue(style["textDecoration"])); v != "" {
		props = append(props, CSSProp{"text-decoration", v})
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

func textDecoration(figma string) string {
	switch figma {
	case "UNDERLINE":
		return "underline"
	case "STRIKETHROUGH":
		return "line-through"
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
