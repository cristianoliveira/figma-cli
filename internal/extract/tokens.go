// Tokens are pure transforms: Figma variables/styles -> a normalized Token list,
// and Token list -> css/tailwind/json text. No network, no IO — everything here
// is testable with plain map[string]any fixtures.
package extract

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

// Token is a single design token: a name path, the namespace it belongs to
// (Category, drives tailwind grouping + unit hints), the rendered CSS value,
// and a style-dictionary style "type" for JSON output.
type Token struct {
	Path     []string `json:"-"`
	Category string   `json:"-"`
	Value    string   `json:"value"`
	Type     string   `json:"type"`
}

// Figma variable resolvedType values (see /variables/local).
const (
	varTypeColor   = "COLOR"
	varTypeString  = "STRING"
	varTypeBoolean = "BOOLEAN"
	varTypeFloat   = "FLOAT"
)

// Token categories — internal namespaces used for unit hints, tailwind
// grouping, and JSON output. Distinct from Figma field names.
const (
	catColor         = "color"
	catString        = "string"
	catBoolean       = "boolean"
	catSpace         = "space"
	catRadius        = "radius"
	catFontSize      = "fontSize"
	catFontWeight    = "fontWeight"
	catFontFamily    = "fontFamily"
	catLineHeight    = "lineHeight"
	catLetterSpacing = "letterSpacing"
	catShadow        = "shadow"
	catOther         = "other"
)

// ExtractTokensFromVariables builds tokens from a /variables/local `meta` object.
// mode selects a named mode ("") for the default mode of each collection.
func ExtractTokensFromVariables(meta map[string]any, mode string) []Token {
	tokens, _ := ExtractTokensFromVariablesE(meta, mode)
	return tokens
}

// ExtractTokensFromVariablesE is like ExtractTokensFromVariables but returns an
// error when a requested mode does not exist.
func ExtractTokensFromVariablesE(meta map[string]any, mode string) ([]Token, error) {
	collections, _ := meta["variableCollections"].(map[string]any)
	variables, _ := meta["variables"].(map[string]any)
	if collections == nil {
		collections = map[string]any{}
	}
	if variables == nil {
		variables = map[string]any{}
	}

	// id -> variable name, used to resolve VARIABLE_ALIAS references.
	idToName := map[string]string{}
	for id, raw := range variables {
		v, _ := raw.(map[string]any)
		if name, _ := v["name"].(string); name != "" {
			idToName[id] = name
		}
	}

	// Resolve the active mode id per collection once, keyed by the collection id
	// (the same id variables reference via variableCollectionId).
	collectionMode := map[string]string{}
	for id, raw := range collections {
		col, _ := raw.(map[string]any)
		if col == nil {
			continue
		}
		modeID, err := resolveModeID(col, mode)
		if err != nil {
			return nil, err
		}
		collectionMode[id] = modeID
	}

	var tokens []Token
	for _, raw := range variables {
		v, _ := raw.(map[string]any)
		if v == nil {
			continue
		}
		name, _ := v["name"].(string)
		if name == "" {
			continue
		}
		vtype, _ := v["resolvedType"].(string)
		colID, _ := v["variableCollectionId"].(string)
		modeID := collectionMode[colID]
		values, _ := v["valuesByMode"].(map[string]any)
		value := values[modeID]
		path := strings.Split(name, "/")

		// codeSyntax.WEB is the designer-authored value — it wins over rendering.
		if cs, ok := v["codeSyntax"].(map[string]any); ok {
			if web, _ := cs["WEB"].(string); web != "" {
				tokens = append(tokens, Token{
					Path:     path,
					Category: categoryForVariable(vtype, scopesOf(v)),
					Value:    web,
					Type:     typeFor(vtype),
				})
				continue
			}
		}

		// Aliases resolve to a reference of the target variable's CSS name.
		if isAlias(value) {
			if tname, ok := idToName[aliasTarget(value)]; ok {
				tokens = append(tokens, Token{
					Path:     path,
					Category: categoryForVariable(vtype, scopesOf(v)),
					Value:    "var(--" + dashPath(strings.Split(tname, "/")) + ")",
					Type:     typeFor(vtype),
				})
			}
			continue
		}

		rendered, ok := renderVariableValue(value, vtype)
		if !ok {
			continue
		}
		category := categoryForVariable(vtype, scopesOf(v))
		tokens = append(tokens, Token{
			Path:     path,
			Category: category,
			Value:    applyUnit(rendered, category),
			Type:     typeFor(vtype),
		})
	}

	sortTokens(tokens)
	return tokens, nil
}

// ExtractTokensFromStyles builds tokens from Figma styles + their resolved nodes.
// styles is the list from /files/{key}/styles; nodes maps node_id -> node (either
// a raw node or {document: node} as returned by /files/{key}/nodes).
func ExtractTokensFromStyles(styles []map[string]any, nodes map[string]any) []Token {
	var tokens []Token
	for _, st := range styles {
		name, _ := st["name"].(string)
		if name == "" {
			continue
		}
		nodeID, _ := st["node_id"].(string)
		node := resolveStyleNode(nodes, nodeID)
		if node == nil {
			continue
		}
		styleType, _ := st["style_type"].(string)
		nameParts := strings.Split(name, "/")

		switch styleType {
		case "FILL":
			if hex := firstSolidColor(node["fills"]); hex != "" {
				tokens = append(tokens, Token{
					Path: append([]string{catColor}, nameParts...), Category: catColor, Value: hex, Type: catColor,
				})
			}
		case "TEXT":
			tokens = append(tokens, typographyTokens(nameParts, node["style"])...)
		case "EFFECT":
			if sh := firstShadow(node["effects"]); sh != "" {
				tokens = append(tokens, Token{
					Path: append([]string{catShadow}, nameParts...), Category: catShadow, Value: sh, Type: catShadow,
				})
			}
		}
	}
	sortTokens(tokens)
	return tokens
}

// ExtractTokensFromDocument scans a document tree and emits a token for every
// distinct color, shadow, and text style it finds. This is the fallback for
// files that use raw fills instead of named Styles or Variables — tokens are
// named by value (e.g. --color-edeef0) since no semantic name exists.
func ExtractTokensFromDocument(value any) []Token {
	var tokens []Token

	// Colors: reuse the same palette walker as `figma colors`.
	for _, e := range CollectColors(value) {
		tokens = append(tokens, Token{
			Path:     []string{catColor, hexSlug(e.Color)},
			Category: catColor, Value: e.Color, Type: catColor,
		})
	}

	// Shadows: distinct DROP_SHADOW renderings, sorted for stable indexing.
	shadowSet := map[string]struct{}{}
	walkShadows(value, func(sh string) {
		if _, ok := shadowSet[sh]; ok {
			return
		}
		shadowSet[sh] = struct{}{}
	})
	shadows := sortedKeys(shadowSet)
	for i, sh := range shadows {
		tokens = append(tokens, Token{
			Path: []string{catShadow, strconv.Itoa(i + 1)}, Category: catShadow, Value: sh, Type: catShadow,
		})
	}

	// Fonts: distinct families and sizes across text nodes.
	famSet := map[string]struct{}{}
	sizeSet := map[string]struct{}{}
	walkTextStyles(value, func(style map[string]any) {
		if fam, _ := style["fontFamily"].(string); fam != "" {
			famSet[fam] = struct{}{}
		}
		if sz := document.NumberValue(style["fontSize"]); sz > 0 {
			sizeSet[numStr(sz)] = struct{}{}
		}
	})
	for _, f := range sortedKeys(famSet) {
		tokens = append(tokens, Token{
			Path: []string{"font-family", dashPath([]string{f})}, Category: catFontFamily, Value: strconv.Quote(f), Type: "typography",
		})
	}
	for _, s := range sortedKeys(sizeSet) {
		tokens = append(tokens, Token{
			Path: []string{"font-size", s}, Category: catFontSize, Value: s + "px", Type: "typography",
		})
	}

	sortTokens(tokens)
	return tokens
}

func walkShadows(value any, emit func(string)) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}
	for _, sh := range shadowsFromEffects(object["effects"]) {
		emit(sh)
	}
	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkShadows(child, emit)
	}
}

func walkTextStyles(value any, visit func(map[string]any)) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}
	if style, ok := object["style"].(map[string]any); ok {
		if _, hasFam := style["fontFamily"]; hasFam {
			visit(style)
		}
	}
	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkTextStyles(child, visit)
	}
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// hexSlug turns "#EDEFF0" into "edeef0" for value-based token names.
func hexSlug(hex string) string {
	return strings.ToLower(strings.TrimPrefix(hex, "#"))
}

// FormatTokens renders tokens in the requested format. Output is sorted, so the
// same input always produces the same bytes (CI-safe, idempotent).
func FormatTokens(tokens []Token, format, prefix string) (string, error) {
	sorted := make([]Token, len(tokens))
	copy(sorted, tokens)
	sortTokens(sorted)
	switch format {
	case "css":
		return formatCSS(sorted, prefix), nil
	case "json":
		return formatJSON(sorted), nil
	case "tailwind":
		return formatTailwind(sorted), nil
	}
	return "", fmt.Errorf("unknown format %q (want css, tailwind, or json)", format)
}

// colorString renders a Figma RGBA object (0..1 channels) to CSS: #RRGGBB when
// fully opaque, rgba(...) when it has transparency.
func colorString(c map[string]any) string {
	a := document.NumberValue(c["a"])
	if a >= 1 {
		return fmt.Sprintf("#%02X%02X%02X", colorChannel(c["r"]), colorChannel(c["g"]), colorChannel(c["b"]))
	}
	return fmt.Sprintf("rgba(%d, %d, %d, %s)", colorChannel(c["r"]), colorChannel(c["g"]), colorChannel(c["b"]), numStr(roundTo(a, 4)))
}

// roundTo limits precision to tame float noise from Figma (e.g. 0.16 stored as
// 0.1599999964...), keeping generated CSS clean and diff-stable.
func roundTo(n float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(n*p) / p
}

// ---- value helpers ----

func resolveModeID(col map[string]any, mode string) (string, error) {
	modes, _ := col["modes"].([]any)
	if mode == "" {
		if def, _ := col["defaultModeId"].(string); def != "" {
			return def, nil
		}
		if len(modes) > 0 {
			if m, ok := modes[0].(map[string]any); ok {
				if id, _ := m["modeId"].(string); id != "" {
					return id, nil
				}
			}
		}
		return "", nil
	}
	for _, raw := range modes {
		m, _ := raw.(map[string]any)
		if name, _ := m["name"].(string); name == mode {
			if id, _ := m["modeId"].(string); id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("unknown mode %q in collection %q", mode, col["name"])
}

func scopesOf(v map[string]any) []string {
	raw, _ := v["scopes"].([]any)
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		if str, ok := s.(string); ok {
			out = append(out, str)
		}
	}
	return out
}

func categoryForVariable(vtype string, scopes []string) string {
	switch vtype {
	case varTypeColor:
		return catColor
	case varTypeString:
		return catString
	case varTypeBoolean:
		return catBoolean
	case varTypeFloat:
		for _, s := range scopes {
			switch s {
			case "CORNER_RADIUS":
				return catRadius
			case "GAP", "WIDTH_HEIGHT":
				return catSpace
			case "FONT_SIZE":
				return catFontSize
			case "FONT_WEIGHT", "FONT_STYLE":
				return catFontWeight
			case "LINE_HEIGHT":
				return catLineHeight
			case "LETTER_SPACING":
				return catLetterSpacing
			}
		}
		return catSpace
	}
	return catOther
}

func typeFor(vtype string) string {
	switch vtype {
	case varTypeColor:
		return "color"
	case varTypeFloat:
		return "number"
	case varTypeString:
		return "string"
	case varTypeBoolean:
		return "boolean"
	}
	return "other"
}

func renderVariableValue(value any, vtype string) (string, bool) {
	switch vtype {
	case varTypeColor:
		obj, _ := value.(map[string]any)
		if obj == nil {
			return "", false
		}
		return colorString(obj), true
	case varTypeString:
		s, _ := value.(string)
		return strconv.Quote(s), true
	case varTypeBoolean:
		b, _ := value.(bool)
		return strconv.FormatBool(b), true
	case varTypeFloat:
		n, ok := value.(float64)
		if !ok {
			return "", false
		}
		return strconv.FormatFloat(n, 'f', -1, 64), true
	}
	return "", false
}

func applyUnit(value, category string) string {
	switch category {
	case catSpace, catRadius, catFontSize, catLineHeight, catLetterSpacing:
		if isNumeric(value) {
			return value + "px"
		}
	}
	return value
}

// ---- style-node helpers ----

func resolveStyleNode(nodes map[string]any, nodeID string) map[string]any {
	raw, ok := nodes[nodeID]
	if !ok {
		return nil
	}
	node, _ := raw.(map[string]any)
	if doc, ok := node["document"].(map[string]any); ok {
		return doc
	}
	return node
}

func firstSolidColor(fills any) string {
	arr, _ := fills.([]any)
	for _, f := range arr {
		paint, _ := f.(map[string]any)
		if paint == nil || paint["visible"] == false {
			continue
		}
		if t, _ := paint["type"].(string); t != paintTypeSolid {
			continue
		}
		if c, ok := paint["color"].(map[string]any); ok {
			return colorString(c)
		}
	}
	return ""
}

func firstShadow(effects any) string {
	for _, sh := range shadowsFromEffects(effects) {
		return sh
	}
	return ""
}

// shadowsFromEffects renders every visible DROP_SHADOW in an effects array.
func shadowsFromEffects(effects any) []string {
	arr, _ := effects.([]any)
	var out []string
	for _, e := range arr {
		eff, _ := e.(map[string]any)
		if eff == nil || eff["visible"] == false {
			continue
		}
		if t, _ := eff["type"].(string); t != "DROP_SHADOW" {
			continue
		}
		if sh := shadowFromEffect(eff); sh != "" {
			out = append(out, sh)
		}
	}
	return out
}

func shadowFromEffect(eff map[string]any) string {
	off, _ := eff["offset"].(map[string]any)
	x := document.NumberValue(off["x"])
	y := document.NumberValue(off["y"])
	radius := document.NumberValue(eff["radius"])
	spread := document.NumberValue(eff["spread"])
	col := ""
	if c, ok := eff["color"].(map[string]any); ok {
		col = colorString(c)
	}
	return fmt.Sprintf("%spx %spx %spx %spx %s", numStr(x), numStr(y), numStr(radius), numStr(spread), col)
}

func typographyTokens(nameParts []string, styleRaw any) []Token {
	style, _ := styleRaw.(map[string]any)
	if style == nil {
		return nil
	}
	base := append([]string{"font"}, nameParts...)
	var tokens []Token
	add := func(leaf, val, cat string) {
		if val == "" {
			return
		}
		path := make([]string, 0, len(base)+1)
		path = append(path, base...)
		path = append(path, leaf)
		tokens = append(tokens, Token{Path: path, Category: cat, Value: val, Type: "typography"})
	}
	if fam, _ := style["fontFamily"].(string); fam != "" {
		add("family", strconv.Quote(fam), catFontFamily)
	}
	if sz := document.NumberValue(style["fontSize"]); sz > 0 {
		add("size", px(sz), catFontSize)
	}
	if w := document.NumberValue(style["fontWeight"]); w > 0 {
		add("weight", numStr(w), catFontWeight)
	}
	if lh := document.NumberValue(style["lineHeightPx"]); lh > 0 {
		add("line-height", px(lh), catLineHeight)
	}
	if ls := style["letterSpacing"]; ls != nil {
		add("letter-spacing", px(document.NumberValue(ls)), catLetterSpacing)
	}
	return tokens
}

// ---- alias helpers ----

func isAlias(value any) bool {
	obj, _ := value.(map[string]any)
	t, _ := obj["type"].(string)
	return t == "VARIABLE_ALIAS"
}

func aliasTarget(value any) string {
	obj, _ := value.(map[string]any)
	id, _ := obj["id"].(string)
	return id
}

// ---- shared helpers ----

func numStr(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func px(n float64) string {
	// Round to 3 decimals to avoid float noise (e.g. 27.50250244140625 -> 27.503),
	// matching Figma Dev Mode's display precision.
	return numStr(math.Round(n*1000)/1000) + "px"
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func sortTokens(tokens []Token) {
	sort.SliceStable(tokens, func(i, j int) bool {
		ki, kj := dashPath(tokens[i].Path), dashPath(tokens[j].Path)
		if ki != kj {
			return ki < kj
		}
		return tokens[i].Value < tokens[j].Value
	})
}

// dashPath turns name segments into a lowercase, dash-separated CSS-safe slug.
// It collapses "/", "_", ".", spaces, and any non-alphanumeric rune to a single "-".
func dashPath(parts []string) string {
	joined := strings.ToLower(strings.Join(parts, "/"))
	var b strings.Builder
	prevDash := false
	for _, r := range joined {
		switch {
		case r == ' ' || r == '/' || r == '_' || r == '.' || r == '-':
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// tokenCSSName renders the CSS custom property name for a token (test helper).
func tokenCSSName(t Token, prefix string) string {
	return "--" + prefix + dashPath(t.Path)
}
