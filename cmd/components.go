package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	figmadiff "github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/spf13/cobra"
)

type componentOutput struct {
	ID                string           `json:"id"`
	Name              string           `json:"name"`
	Type              string           `json:"type"`
	Text              string           `json:"text,omitempty"`
	ComponentID       string           `json:"componentId,omitempty"`
	ComponentSetID    string           `json:"componentSetId,omitempty"`
	VariantProperties map[string]any   `json:"variantProperties,omitempty"`
	Fills             []string         `json:"fills,omitempty"`
	Strokes           []string         `json:"strokes,omitempty"`
	Paints            paintsOutput     `json:"paints,omitempty"`
	StrokeWeight      float64          `json:"strokeWeight,omitempty"`
	StrokeAlign       string           `json:"strokeAlign,omitempty"`
	StrokeDashes      []float64        `json:"strokeDashes,omitempty"`
	Effects           []effectOutput   `json:"effects,omitempty"`
	Opacity           *float64         `json:"opacity,omitempty"`
	Bounds            boundsOutput     `json:"bounds,omitempty"`
	CornerRadius      *float64         `json:"cornerRadius,omitempty"`
	Layout            layoutOutput     `json:"layout,omitempty"`
	Typography        typographyOutput `json:"typography,omitempty"`
}

type paintsOutput struct {
	Fills   []paintOutput `json:"fills,omitempty"`
	Strokes []paintOutput `json:"strokes,omitempty"`
}

type paintOutput struct {
	Type    string  `json:"type,omitempty"`
	Color   string  `json:"color,omitempty"`
	Opacity float64 `json:"opacity,omitempty"`
	Visible bool    `json:"visible"`
}

type boundsOutput struct {
	X      float64 `json:"x,omitempty"`
	Y      float64 `json:"y,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

type layoutOutput struct {
	Mode                   string  `json:"mode,omitempty"`
	Gap                    float64 `json:"gap,omitempty"`
	PaddingTop             float64 `json:"paddingTop,omitempty"`
	PaddingRight           float64 `json:"paddingRight,omitempty"`
	PaddingBottom          float64 `json:"paddingBottom,omitempty"`
	PaddingLeft            float64 `json:"paddingLeft,omitempty"`
	LayoutAlign            string  `json:"layoutAlign,omitempty"`
	LayoutGrow             float64 `json:"layoutGrow,omitempty"`
	LayoutSizingHorizontal string  `json:"layoutSizingHorizontal,omitempty"`
	LayoutSizingVertical   string  `json:"layoutSizingVertical,omitempty"`
	PrimaryAxisSizingMode  string  `json:"primaryAxisSizingMode,omitempty"`
	CounterAxisSizingMode  string  `json:"counterAxisSizingMode,omitempty"`
	PrimaryAxisAlignItems  string  `json:"primaryAxisAlignItems,omitempty"`
	CounterAxisAlignItems  string  `json:"counterAxisAlignItems,omitempty"`
}

type typographyOutput struct {
	FontFamily          string  `json:"fontFamily,omitempty"`
	FontSize            float64 `json:"fontSize,omitempty"`
	FontWeight          float64 `json:"fontWeight,omitempty"`
	LineHeight          float64 `json:"lineHeight,omitempty"`
	LetterSpacing       float64 `json:"letterSpacing,omitempty"`
	ParagraphSpacing    float64 `json:"paragraphSpacing,omitempty"`
	TextCase            string  `json:"textCase,omitempty"`
	TextDecoration      string  `json:"textDecoration,omitempty"`
	TextAlignHorizontal string  `json:"textAlignHorizontal,omitempty"`
	TextAlignVertical   string  `json:"textAlignVertical,omitempty"`
}

type effectOutput struct {
	Type    string  `json:"type,omitempty"`
	Color   string  `json:"color,omitempty"`
	Radius  float64 `json:"radius,omitempty"`
	Visible bool    `json:"visible"`
}

var componentsCmd = &cobra.Command{
	Use:   "components [figma-url-or-file-id]",
	Short: "List descendant nodes within a Figma element as JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := parseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		nodeIDs := resolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			fmt.Fprintln(os.Stderr, "error: components requires --id or a Figma URL with node-id")
			os.Exit(1)
		}
		token, err := getFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		apiURL, err := buildFindAPIURL(input.fileID, nodeIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building components URL: %v\n", err)
			os.Exit(1)
		}
		figmaJSON, err := figmadiff.FetchFigmaJSON(apiURL, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma node: %v\n", err)
			os.Exit(1)
		}
		output, err := json.MarshalIndent(extractComponentsFromFigmaJSON(figmaJSON), "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func extractComponentsFromFigmaJSON(figmaJSON map[string]any) []componentOutput {
	if document, ok := figmaJSON["document"]; ok {
		return extractComponents(document)
	}
	var components []componentOutput
	nodes, ok := figmaJSON["nodes"].(map[string]any)
	if !ok {
		return components
	}
	for _, node := range nodes {
		nodeObject, ok := node.(map[string]any)
		if !ok {
			continue
		}
		document, ok := nodeObject["document"]
		if ok {
			components = append(components, extractComponents(document)...)
		}
	}
	return components
}

func extractComponents(value any) []componentOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	component := componentOutput{
		ID:                layerStringValue(object["id"]),
		Name:              layerStringValue(object["name"]),
		Type:              layerStringValue(object["type"]),
		Text:              layerStringValue(object["characters"]),
		ComponentID:       layerStringValue(object["componentId"]),
		ComponentSetID:    layerStringValue(object["componentSetId"]),
		VariantProperties: mapValue(object["variantProperties"]),
		Fills:             colorsFromPaints(object["fills"]),
		Strokes:           colorsFromPaints(object["strokes"]),
		Paints:            paintsFromObject(object),
		StrokeWeight:      numberValue(object["strokeWeight"]),
		StrokeAlign:       layerStringValue(object["strokeAlign"]),
		StrokeDashes:      numberSlice(object["strokeDashes"]),
		Effects:           effectsFromValue(object["effects"]),
		Opacity:           optionalNumber(object["opacity"]),
		Bounds:            boundsFromValue(object["absoluteBoundingBox"]),
		CornerRadius:      optionalNumber(object["cornerRadius"]),
		Layout:            layoutFromObject(object),
		Typography:        typographyFromValue(object["style"]),
	}
	components := []componentOutput{component}
	children, ok := object["children"].([]any)
	if !ok {
		return components
	}
	for _, child := range children {
		components = append(components, extractComponents(child)...)
	}
	return components
}

func colorsFromPaints(value any) []string {
	paints, ok := value.([]any)
	if !ok {
		return nil
	}
	colors := make([]string, 0, len(paints))
	for _, paint := range paints {
		paintObject, ok := paint.(map[string]any)
		if !ok || paintObject["visible"] == false {
			continue
		}
		color := colorHexFromPaint(paintObject)
		if color != "" {
			colors = append(colors, color)
		}
	}
	return colors
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
			Type:    layerStringValue(paintObject["type"]),
			Color:   colorHexFromPaint(paintObject),
			Opacity: numberValue(paintObject["opacity"]),
			Visible: paintObject["visible"] != false,
		})
	}
	return outputs
}

func colorHexFromPaint(paint map[string]any) string {
	color, ok := paint["color"].(map[string]any)
	if !ok {
		return ""
	}
	return fmt.Sprintf("#%02X%02X%02X", colorChannel(color["r"]), colorChannel(color["g"]), colorChannel(color["b"]))
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
		outputs = append(outputs, effectOutput{
			Type:    layerStringValue(effectObject["type"]),
			Color:   colorHexFromPaint(effectObject),
			Radius:  numberValue(effectObject["radius"]),
			Visible: effectObject["visible"] != false,
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
		Mode:                   layerStringValue(object["layoutMode"]),
		Gap:                    numberValue(object["itemSpacing"]),
		PaddingTop:             numberValue(object["paddingTop"]),
		PaddingRight:           numberValue(object["paddingRight"]),
		PaddingBottom:          numberValue(object["paddingBottom"]),
		PaddingLeft:            numberValue(object["paddingLeft"]),
		LayoutAlign:            layerStringValue(object["layoutAlign"]),
		LayoutGrow:             numberValue(object["layoutGrow"]),
		LayoutSizingHorizontal: layerStringValue(object["layoutSizingHorizontal"]),
		LayoutSizingVertical:   layerStringValue(object["layoutSizingVertical"]),
		PrimaryAxisSizingMode:  layerStringValue(object["primaryAxisSizingMode"]),
		CounterAxisSizingMode:  layerStringValue(object["counterAxisSizingMode"]),
		PrimaryAxisAlignItems:  layerStringValue(object["primaryAxisAlignItems"]),
		CounterAxisAlignItems:  layerStringValue(object["counterAxisAlignItems"]),
	}
}

func typographyFromValue(value any) typographyOutput {
	style, ok := value.(map[string]any)
	if !ok {
		return typographyOutput{}
	}
	return typographyOutput{
		FontFamily:          layerStringValue(style["fontFamily"]),
		FontSize:            numberValue(style["fontSize"]),
		FontWeight:          numberValue(style["fontWeight"]),
		LineHeight:          numberValue(style["lineHeightPx"]),
		LetterSpacing:       numberValue(style["letterSpacing"]),
		ParagraphSpacing:    numberValue(style["paragraphSpacing"]),
		TextCase:            layerStringValue(style["textCase"]),
		TextDecoration:      layerStringValue(style["textDecoration"]),
		TextAlignHorizontal: layerStringValue(style["textAlignHorizontal"]),
		TextAlignVertical:   layerStringValue(style["textAlignVertical"]),
	}
}

func optionalNumber(value any) *float64 {
	number, ok := value.(float64)
	if !ok {
		return nil
	}
	return &number
}

func mapValue(value any) map[string]any {
	object, _ := value.(map[string]any)
	return object
}

func numberSlice(value any) []float64 {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	numbers := make([]float64, 0, len(items))
	for _, item := range items {
		numbers = append(numbers, numberValue(item))
	}
	return numbers
}

func numberValue(value any) float64 {
	number, ok := value.(float64)
	if !ok {
		return 0
	}
	return number
}

func colorChannel(value any) int {
	return int(math.Round(numberValue(value) * 255))
}

func init() {
	componentsCmd.Flags().String("id", "", "node ID to inspect; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(componentsCmd)
}
