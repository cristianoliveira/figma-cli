package extract

import (
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

const (
	componentTypeComponent    = "COMPONENT"
	componentTypeComponentSet = "COMPONENT_SET"
	componentTypeInstance     = "INSTANCE"
)

// ComponentOutput is a curated summary of a node, used by `figma components`.
// ComponentInstanceUsage describes one occurrence of a component instance.
type ComponentInstanceUsage struct {
	ID                  string         `json:"id"`
	Name                string         `json:"name"`
	Path                []string       `json:"path,omitempty"`
	VariantProperties   map[string]any `json:"variantProperties,omitempty"`
	ComponentProperties map[string]any `json:"componentProperties,omitempty"`
}

// ComponentUsageOutput groups instances by exact Figma component ID.
type ComponentUsageOutput struct {
	ComponentID string                   `json:"componentId"`
	Name        string                   `json:"name"`
	Count       int                      `json:"count"`
	Instances   []ComponentInstanceUsage `json:"instances"`
}

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
	name := document.StringValue(object["name"])
	path := append([]string(nil), parentPath...)
	if name != "" {
		path = append(path, name)
	}
	component := ComponentOutput{
		ID:                  document.StringValue(object["id"]),
		Name:                name,
		Type:                document.StringValue(object["type"]),
		Text:                document.StringValue(object["characters"]),
		ComponentID:         document.StringValue(object["componentId"]),
		ComponentSetID:      document.StringValue(object["componentSetId"]),
		VariantProperties:   document.MapValue(object["variantProperties"]),
		ComponentProperties: document.MapValue(object["componentProperties"]),
		PropertyDefinitions: document.MapValue(object["componentPropertyDefinitions"]),
		Path:                path,
		Fills:               colorsFromPaints(object["fills"]),
		Strokes:             colorsFromPaints(object["strokes"]),
		Paints:              paintsFromObject(object),
		StrokeWeight:        document.NumberValue(object["strokeWeight"]),
		StrokeAlign:         document.StringValue(object["strokeAlign"]),
		StrokeDashes:        document.NumberSlice(object["strokeDashes"]),
		Effects:             effectsFromValue(object["effects"]),
		Opacity:             document.OptionalNumber(object["opacity"]),
		Bounds:              boundsFromValue(object["absoluteBoundingBox"]),
		CornerRadius:        document.OptionalNumber(object["cornerRadius"]),
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

// AggregateComponentUsage groups instances by exact component ID in first-seen order.
// Detached instances without a component ID are omitted rather than guessed by name.
func AggregateComponentUsage(components []ComponentOutput) []ComponentUsageOutput {
	usage := make([]ComponentUsageOutput, 0)
	indexes := make(map[string]int)
	for _, component := range components {
		if component.Type != componentTypeInstance || component.ComponentID == "" {
			continue
		}
		index, exists := indexes[component.ComponentID]
		if !exists {
			index = len(usage)
			indexes[component.ComponentID] = index
			usage = append(usage, ComponentUsageOutput{
				ComponentID: component.ComponentID,
				Name:        component.Name,
				Instances:   make([]ComponentInstanceUsage, 0),
			})
		}
		usage[index].Count++
		usage[index].Instances = append(usage[index].Instances, ComponentInstanceUsage{
			ID:                  component.ID,
			Name:                component.Name,
			Path:                component.Path,
			VariantProperties:   component.VariantProperties,
			ComponentProperties: component.ComponentProperties,
		})
	}
	return usage
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
		if document.StringValue(node["type"]) == wantedType {
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
			Type:          document.StringValue(paintObject["type"]),
			Color:         colorHexFromPaint(paintObject),
			Opacity:       document.NumberValue(paintObject["opacity"]),
			Visible:       paintObject["visible"] != false,
			ImageRef:      document.StringValue(paintObject["imageRef"]),
			ScaleMode:     document.StringValue(paintObject["scaleMode"]),
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
			opacity = document.NumberValue(alpha)
		}
		outputs = append(outputs, gradientStopOutput{
			Position: document.NumberValue(stopObject["position"]),
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
			Type:      document.StringValue(effectObject["type"]),
			Color:     colorHexFromPaint(effectObject),
			Radius:    document.NumberValue(effectObject["radius"]),
			Spread:    document.NumberValue(effectObject["spread"]),
			OffsetX:   document.NumberValue(offset["x"]),
			OffsetY:   document.NumberValue(offset["y"]),
			BlendMode: document.StringValue(effectObject["blendMode"]),
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
	return boundsOutput{X: document.NumberValue(object["x"]), Y: document.NumberValue(object["y"]), Width: document.NumberValue(object["width"]), Height: document.NumberValue(object["height"])}
}

func layoutFromObject(object map[string]any) layoutOutput {
	return layoutOutput{
		Mode:                   document.StringValue(object["layoutMode"]),
		Gap:                    document.NumberValue(object["itemSpacing"]),
		PaddingTop:             document.NumberValue(object["paddingTop"]),
		PaddingRight:           document.NumberValue(object["paddingRight"]),
		PaddingBottom:          document.NumberValue(object["paddingBottom"]),
		PaddingLeft:            document.NumberValue(object["paddingLeft"]),
		LayoutAlign:            document.StringValue(object["layoutAlign"]),
		LayoutGrow:             document.NumberValue(object["layoutGrow"]),
		LayoutSizingHorizontal: document.StringValue(object["layoutSizingHorizontal"]),
		LayoutSizingVertical:   document.StringValue(object["layoutSizingVertical"]),
		PrimaryAxisSizingMode:  document.StringValue(object["primaryAxisSizingMode"]),
		CounterAxisSizingMode:  document.StringValue(object["counterAxisSizingMode"]),
		PrimaryAxisAlignItems:  document.StringValue(object["primaryAxisAlignItems"]),
		CounterAxisAlignItems:  document.StringValue(object["counterAxisAlignItems"]),
		Wrap:                   document.StringValue(object["layoutWrap"]),
		CounterAxisSpacing:     document.NumberValue(object["counterAxisSpacing"]),
		Positioning:            document.StringValue(object["layoutPositioning"]),
		MinWidth:               document.NumberValue(object["minWidth"]),
		MaxWidth:               document.NumberValue(object["maxWidth"]),
		MinHeight:              document.NumberValue(object["minHeight"]),
		MaxHeight:              document.NumberValue(object["maxHeight"]),
		Constraints:            constraintsFromValue(object["constraints"]),
	}
}

func constraintsFromValue(value any) *constraintsOutput {
	constraints, _ := value.(map[string]any)
	horizontal := document.StringValue(constraints["horizontal"])
	vertical := document.StringValue(constraints["vertical"])
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
		FontFamily:          document.StringValue(style["fontFamily"]),
		FontSize:            document.NumberValue(style["fontSize"]),
		FontWeight:          document.NumberValue(style["fontWeight"]),
		LineHeight:          document.NumberValue(style["lineHeightPx"]),
		LetterSpacing:       document.NumberValue(style["letterSpacing"]),
		ParagraphSpacing:    document.NumberValue(style["paragraphSpacing"]),
		TextCase:            document.StringValue(style["textCase"]),
		TextDecoration:      document.StringValue(style["textDecoration"]),
		TextAlignHorizontal: document.StringValue(style["textAlignHorizontal"]),
		TextAlignVertical:   document.StringValue(style["textAlignVertical"]),
	}
}
