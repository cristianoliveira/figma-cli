package extract

import (
	"strings"
)

const (
	componentTypeComponent    = "COMPONENT"
	componentTypeComponentSet = "COMPONENT_SET"
	componentTypeInstance     = "INSTANCE"
)

// ComponentOutput is a curated summary of a node, used by `figma components`.
type ComponentOutput struct {
	ID                  string           `json:"id"`
	Name                string           `json:"name"`
	Type                string           `json:"type"`
	Text                string           `json:"text,omitempty"`
	ComponentID         string           `json:"componentId,omitempty"`
	ComponentSetID      string           `json:"componentSetId,omitempty"`
	VariantProperties   map[string]any   `json:"variantProperties,omitempty"`
	ComponentProperties map[string]any   `json:"componentProperties,omitempty"`
	PropertyDefinitions map[string]any   `json:"propertyDefinitions,omitempty"`
	Path                []string         `json:"path,omitempty"`
	Fills               []string         `json:"fills,omitempty"`
	Strokes             []string         `json:"strokes,omitempty"`
	Paints              paintsOutput     `json:"paints,omitempty"`
	StrokeWeight        float64          `json:"strokeWeight,omitempty"`
	StrokeAlign         string           `json:"strokeAlign,omitempty"`
	StrokeDashes        []float64        `json:"strokeDashes,omitempty"`
	Effects             []effectOutput   `json:"effects,omitempty"`
	Opacity             *float64         `json:"opacity,omitempty"`
	Bounds              boundsOutput     `json:"bounds,omitempty"`
	CornerRadius        *float64         `json:"cornerRadius,omitempty"`
	Layout              layoutOutput     `json:"layout,omitempty"`
	Typography          typographyOutput `json:"typography,omitempty"`
}

// ExtractComponentsFromDocuments extracts and combines multiple node subtrees.
func ExtractComponentsFromDocuments(documents []any) []ComponentOutput {
	components := make([]ComponentOutput, 0)
	for _, document := range documents {
		components = append(components, ExtractComponents(document)...)
	}
	return components
}

// ExtractRawComponentsFromDocuments extracts and combines raw node subtrees.
func ExtractRawComponentsFromDocuments(documents []any) []map[string]any {
	components := make([]map[string]any, 0)
	for _, document := range documents {
		components = append(components, ExtractRawComponents(document)...)
	}
	return components
}

// ExtractComponents walks a document and returns component-domain nodes only.
func ExtractComponents(value any) []ComponentOutput {
	return extractComponents(value, nil)
}

func extractComponents(value any, parentPath []string) []ComponentOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	name := StringValue(object["name"])
	path := append([]string(nil), parentPath...)
	if name != "" {
		path = append(path, name)
	}
	component := ComponentOutput{
		ID:                  StringValue(object["id"]),
		Name:                name,
		Type:                StringValue(object["type"]),
		Text:                StringValue(object["characters"]),
		ComponentID:         StringValue(object["componentId"]),
		ComponentSetID:      StringValue(object["componentSetId"]),
		VariantProperties:   mapValue(object["variantProperties"]),
		ComponentProperties: mapValue(object["componentProperties"]),
		PropertyDefinitions: mapValue(object["componentPropertyDefinitions"]),
		Path:                path,
		Fills:               colorsFromPaints(object["fills"]),
		Strokes:             colorsFromPaints(object["strokes"]),
		Paints:              paintsFromObject(object),
		StrokeWeight:        numberValue(object["strokeWeight"]),
		StrokeAlign:         StringValue(object["strokeAlign"]),
		StrokeDashes:        numberSlice(object["strokeDashes"]),
		Effects:             effectsFromValue(object["effects"]),
		Opacity:             optionalNumber(object["opacity"]),
		Bounds:              boundsFromValue(object["absoluteBoundingBox"]),
		CornerRadius:        optionalNumber(object["cornerRadius"]),
		Layout:              layoutFromObject(object),
		Typography:          typographyFromValue(object["style"]),
	}
	components := make([]ComponentOutput, 0)
	if isComponentType(component.Type) {
		components = append(components, component)
	}
	children, ok := object["children"].([]any)
	if !ok {
		return components
	}
	for _, child := range children {
		components = append(components, extractComponents(child, path)...)
	}
	return components
}

func isComponentType(nodeType string) bool {
	return nodeType == componentTypeComponent || nodeType == componentTypeComponentSet || nodeType == componentTypeInstance
}

// FilterComponentsByKind filters component-domain output using user-facing kind names.
func FilterComponentsByKind(components []ComponentOutput, kind string) []ComponentOutput {
	if kind == "" {
		return components
	}
	wantedType := map[string]string{"component": componentTypeComponent, "set": componentTypeComponentSet, "instance": componentTypeInstance}[kind]
	filtered := make([]ComponentOutput, 0)
	for _, component := range components {
		if component.Type == wantedType {
			filtered = append(filtered, component)
		}
	}
	return filtered
}

// FilterRawComponentsByKind filters raw output when an explicit kind is requested.
func FilterRawComponentsByKind(nodes []map[string]any, kind string) []map[string]any {
	if kind == "" {
		return nodes
	}
	wantedType := map[string]string{"component": componentTypeComponent, "set": componentTypeComponentSet, "instance": componentTypeInstance}[kind]
	filtered := make([]map[string]any, 0)
	for _, node := range nodes {
		if StringValue(node["type"]) == wantedType {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// ExtractRawComponents returns each node as its raw map, for `--raw` output.
func ExtractRawComponents(value any) []map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	components := []map[string]any{object}
	children, ok := object["children"].([]any)
	if !ok {
		return components
	}
	for _, child := range children {
		components = append(components, ExtractRawComponents(child)...)
	}
	return components
}

// FilterByName filters extracted components by a case-insensitive name substring.
// It accepts the two shapes produced by the components command (curated or raw).
func FilterByName(value any, nameFilter string) any {
	lower := strings.ToLower(nameFilter)
	switch v := value.(type) {
	case []ComponentOutput:
		var filtered []ComponentOutput
		for _, c := range v {
			if strings.Contains(strings.ToLower(c.Name), lower) {
				filtered = append(filtered, c)
			}
		}
		return filtered
	case []map[string]any:
		var filtered []map[string]any
		for _, c := range v {
			name, _ := c["name"].(string)
			if strings.Contains(strings.ToLower(name), lower) {
				filtered = append(filtered, c)
			}
		}
		return filtered
	}
	return value
}

func paintsFromObject(object map[string]any) paintsOutput {
	return paintsOutput{Fills: paintOutputsFromValue(object["fills"]), Strokes: paintOutputsFromValue(object["strokes"])}
}

func paintOutputsFromValue(value any) []paintOutput {
	paints, ok := value.([]any)
	if !ok {
		return nil
	}
	outputs := make([]paintOutput, 0, len(paints))
	for _, paint := range paints {
		paintObject, ok := paint.(map[string]any)
		if !ok {
			continue
		}
		outputs = append(outputs, paintOutput{
			Type:          StringValue(paintObject["type"]),
			Color:         colorHexFromPaint(paintObject),
			Opacity:       numberValue(paintObject["opacity"]),
			Visible:       paintObject["visible"] != false,
			ImageRef:      StringValue(paintObject["imageRef"]),
			ScaleMode:     StringValue(paintObject["scaleMode"]),
			GradientStops: gradientStopsFromValue(paintObject["gradientStops"]),
		})
	}
	return outputs
}

func gradientStopsFromValue(value any) []gradientStopOutput {
	stops, _ := value.([]any)
	outputs := make([]gradientStopOutput, 0, len(stops))
	for _, stop := range stops {
		stopObject, ok := stop.(map[string]any)
		if !ok {
			continue
		}
		color, ok := stopObject["color"].(map[string]any)
		if !ok {
			continue
		}
		opacity := 1.0
		if alpha, exists := color["a"]; exists {
			opacity = numberValue(alpha)
		}
		outputs = append(outputs, gradientStopOutput{
			Position: numberValue(stopObject["position"]),
			Color:    colorHexFromPaint(stopObject),
			Opacity:  opacity,
		})
	}
	return outputs
}

func effectsFromValue(value any) []effectOutput {
	effects, ok := value.([]any)
	if !ok {
		return nil
	}
	outputs := make([]effectOutput, 0, len(effects))
	for _, effect := range effects {
		effectObject, ok := effect.(map[string]any)
		if !ok {
			continue
		}
		offset, _ := effectObject["offset"].(map[string]any)
		outputs = append(outputs, effectOutput{
			Type:      StringValue(effectObject["type"]),
			Color:     colorHexFromPaint(effectObject),
			Radius:    numberValue(effectObject["radius"]),
			Spread:    numberValue(effectObject["spread"]),
			OffsetX:   numberValue(offset["x"]),
			OffsetY:   numberValue(offset["y"]),
			BlendMode: StringValue(effectObject["blendMode"]),
			Visible:   effectObject["visible"] != false,
		})
	}
	return outputs
}

func boundsFromValue(value any) boundsOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return boundsOutput{}
	}
	return boundsOutput{X: numberValue(object["x"]), Y: numberValue(object["y"]), Width: numberValue(object["width"]), Height: numberValue(object["height"])}
}

func layoutFromObject(object map[string]any) layoutOutput {
	return layoutOutput{
		Mode:                   StringValue(object["layoutMode"]),
		Gap:                    numberValue(object["itemSpacing"]),
		PaddingTop:             numberValue(object["paddingTop"]),
		PaddingRight:           numberValue(object["paddingRight"]),
		PaddingBottom:          numberValue(object["paddingBottom"]),
		PaddingLeft:            numberValue(object["paddingLeft"]),
		LayoutAlign:            StringValue(object["layoutAlign"]),
		LayoutGrow:             numberValue(object["layoutGrow"]),
		LayoutSizingHorizontal: StringValue(object["layoutSizingHorizontal"]),
		LayoutSizingVertical:   StringValue(object["layoutSizingVertical"]),
		PrimaryAxisSizingMode:  StringValue(object["primaryAxisSizingMode"]),
		CounterAxisSizingMode:  StringValue(object["counterAxisSizingMode"]),
		PrimaryAxisAlignItems:  StringValue(object["primaryAxisAlignItems"]),
		CounterAxisAlignItems:  StringValue(object["counterAxisAlignItems"]),
		Wrap:                   StringValue(object["layoutWrap"]),
		CounterAxisSpacing:     numberValue(object["counterAxisSpacing"]),
		Positioning:            StringValue(object["layoutPositioning"]),
		MinWidth:               numberValue(object["minWidth"]),
		MaxWidth:               numberValue(object["maxWidth"]),
		MinHeight:              numberValue(object["minHeight"]),
		MaxHeight:              numberValue(object["maxHeight"]),
		Constraints:            constraintsFromValue(object["constraints"]),
	}
}

func constraintsFromValue(value any) *constraintsOutput {
	constraints, _ := value.(map[string]any)
	horizontal := StringValue(constraints["horizontal"])
	vertical := StringValue(constraints["vertical"])
	if horizontal == "" && vertical == "" {
		return nil
	}
	return &constraintsOutput{Horizontal: horizontal, Vertical: vertical}
}

func typographyFromValue(value any) typographyOutput {
	style, ok := value.(map[string]any)
	if !ok {
		return typographyOutput{}
	}
	return typographyOutput{
		FontFamily:          StringValue(style["fontFamily"]),
		FontSize:            numberValue(style["fontSize"]),
		FontWeight:          numberValue(style["fontWeight"]),
		LineHeight:          numberValue(style["lineHeightPx"]),
		LetterSpacing:       numberValue(style["letterSpacing"]),
		ParagraphSpacing:    numberValue(style["paragraphSpacing"]),
		TextCase:            StringValue(style["textCase"]),
		TextDecoration:      StringValue(style["textDecoration"]),
		TextAlignHorizontal: StringValue(style["textAlignHorizontal"]),
		TextAlignVertical:   StringValue(style["textAlignVertical"]),
	}
}
