// Package extract turns parsed Figma document trees into the structured output
// models emitted by the CLI commands. It holds no cobra or IO concerns — only
// pure transforms from map[string]any document nodes into typed output.
package extract

// Shared output shapes embedded by the command-specific output structs.

const (
	layoutModeHorizontal      = "HORIZONTAL"
	layoutModeVertical        = "VERTICAL"
	layoutPositioningAbsolute = "ABSOLUTE"
)

type gradientStopOutput struct {
	Position float64 `json:"position"`
	Color    string  `json:"color"`
	Opacity  float64 `json:"opacity"`
}

type paintOutput struct {
	Type          string               `json:"type,omitempty"`
	Color         string               `json:"color,omitempty"`
	Opacity       float64              `json:"opacity,omitempty"`
	Visible       bool                 `json:"visible"`
	ImageRef      string               `json:"imageRef,omitempty"`
	ScaleMode     string               `json:"scaleMode,omitempty"`
	GradientStops []gradientStopOutput `json:"gradientStops,omitempty"`
}

type paintsOutput struct {
	Fills   []paintOutput `json:"fills,omitempty"`
	Strokes []paintOutput `json:"strokes,omitempty"`
}

type effectOutput struct {
	Type      string  `json:"type,omitempty"`
	Color     string  `json:"color,omitempty"`
	Radius    float64 `json:"radius,omitempty"`
	Spread    float64 `json:"spread,omitempty"`
	OffsetX   float64 `json:"offsetX,omitempty"`
	OffsetY   float64 `json:"offsetY,omitempty"`
	BlendMode string  `json:"blendMode,omitempty"`
	Visible   bool    `json:"visible"`
}

type boundsOutput struct {
	X      float64 `json:"x,omitempty"`
	Y      float64 `json:"y,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

type constraintsOutput struct {
	Horizontal string `json:"horizontal,omitempty"`
	Vertical   string `json:"vertical,omitempty"`
}

type layoutOutput struct {
	Mode                   string             `json:"mode,omitempty"`
	Gap                    float64            `json:"gap,omitempty"`
	PaddingTop             float64            `json:"paddingTop,omitempty"`
	PaddingRight           float64            `json:"paddingRight,omitempty"`
	PaddingBottom          float64            `json:"paddingBottom,omitempty"`
	PaddingLeft            float64            `json:"paddingLeft,omitempty"`
	LayoutAlign            string             `json:"layoutAlign,omitempty"`
	LayoutGrow             float64            `json:"layoutGrow,omitempty"`
	LayoutSizingHorizontal string             `json:"layoutSizingHorizontal,omitempty"`
	LayoutSizingVertical   string             `json:"layoutSizingVertical,omitempty"`
	PrimaryAxisSizingMode  string             `json:"primaryAxisSizingMode,omitempty"`
	CounterAxisSizingMode  string             `json:"counterAxisSizingMode,omitempty"`
	PrimaryAxisAlignItems  string             `json:"primaryAxisAlignItems,omitempty"`
	CounterAxisAlignItems  string             `json:"counterAxisAlignItems,omitempty"`
	Wrap                   string             `json:"wrap,omitempty"`
	CounterAxisSpacing     float64            `json:"counterAxisSpacing,omitempty"`
	Positioning            string             `json:"positioning,omitempty"`
	MinWidth               float64            `json:"minWidth,omitempty"`
	MaxWidth               float64            `json:"maxWidth,omitempty"`
	MinHeight              float64            `json:"minHeight,omitempty"`
	MaxHeight              float64            `json:"maxHeight,omitempty"`
	Constraints            *constraintsOutput `json:"constraints,omitempty"`
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
