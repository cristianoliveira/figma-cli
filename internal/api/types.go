package api

// FileMeta represents metadata about a Figma file.
type FileMeta struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	LastModified string `json:"lastModified"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Version      string `json:"version"`
	// Add other fields as needed
}

// Version represents a version in a file's version history.
type Version struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"createdAt"`
	Label       string `json:"label"`
	Description string `json:"description"`
	User        User   `json:"user"`
}

// Reaction represents a reaction to a comment.
type Reaction struct {
	Emoji string `json:"emoji"`
	User  User   `json:"user"`
}

// Component represents a published component.
type Component struct {
	Key             string `json:"key"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	User            User   `json:"user"`
	ContainingFrame *Node  `json:"containingFrame,omitempty"`
	// Additional fields can be added
}

// ComponentSet represents a collection of component variants.
type ComponentSet struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	User        User   `json:"user"`
}

// Team represents a Figma team.
type Team struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Project represents a Figma project.
type Project struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"createdAt"`
	ModifiedAt string `json:"modifiedAt"`
}

// Variable represents a design variable (Enterprise feature).
type Variable struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Key                  string `json:"key"`
	VariableCollectionID string `json:"variableCollectionId"`
	// Additional fields as needed
}

// DevResource represents a developer resource.
type DevResource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Additional fields as needed
}

// Webhook represents a webhook configuration.
type Webhook struct {
	ID        string `json:"id"`
	EventType string `json:"eventType"`
	URL       string `json:"url"`
	Status    string `json:"status"`
	// Additional fields as needed
}

// WebhookRequest represents a request to create a webhook.
type WebhookRequest struct {
	EventType string `json:"eventType"`
	URL       string `json:"url"`
	// Additional fields as needed
}

// WebhookUpdate represents updates to a webhook.
type WebhookUpdate struct {
	Status string `json:"status,omitempty"`
	// Additional fields as needed
}

// ReactionRequest represents a request to add a reaction.
type ReactionRequest struct {
	Emoji string `json:"emoji"`
}

// Color represents an RGBA color with values normalized to 0..1.
type Color struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
	A float64 `json:"a"`
}

// Paint represents a fill or stroke paint.
type Paint struct {
	Type    string  `json:"type"` // "SOLID", "GRADIENT_LINEAR", "IMAGE", etc.
	Color   *Color  `json:"color,omitempty"`
	Opacity float64 `json:"opacity,omitempty"`
	// TODO: add gradient, image fields
}

// LayoutConstraint represents how a node is positioned within its parent.
type LayoutConstraint struct {
	Vertical   string `json:"vertical,omitempty"`   // "TOP", "BOTTOM", "CENTER", "TOP_BOTTOM", "SCALE"
	Horizontal string `json:"horizontal,omitempty"` // "LEFT", "RIGHT", "CENTER", "LEFT_RIGHT", "SCALE"
}

// TextStyle represents text style properties.
type TextStyle struct {
	FontFamily          string  `json:"fontFamily,omitempty"`
	FontPostScriptName  string  `json:"fontPostScriptName,omitempty"`
	FontStyle           string  `json:"fontStyle,omitempty"`
	Italic              bool    `json:"italic,omitempty"`
	FontWeight          int     `json:"fontWeight,omitempty"`
	FontSize            float64 `json:"fontSize,omitempty"`
	TextCase            string  `json:"textCase,omitempty"`            // "ORIGINAL", "UPPER", "LOWER", "TITLE", "SMALL_CAPS", "SMALL_CAPS_FORCED"
	TextAlignHorizontal string  `json:"textAlignHorizontal,omitempty"` // "LEFT", "RIGHT", "CENTER", "JUSTIFIED"
	TextAlignVertical   string  `json:"textAlignVertical,omitempty"`   // "TOP", "CENTER", "BOTTOM"
	LetterSpacing       float64 `json:"letterSpacing,omitempty"`
	LineHeight          float64 `json:"lineHeight,omitempty"`
	ParagraphSpacing    float64 `json:"paragraphSpacing,omitempty"`
	ParagraphIndent     float64 `json:"paragraphIndent,omitempty"`
	TextDecoration      string  `json:"textDecoration,omitempty"` // "NONE", "STRIKETHROUGH", "UNDERLINE"
	TextAutoResize      string  `json:"textAutoResize,omitempty"` // "NONE", "WIDTH_AND_HEIGHT", "HEIGHT", "TRUNCATE"
}

// Effect represents a visual effect like shadow or blur.
type Effect struct {
	Type    string  `json:"type"` // "DROP_SHADOW", "INNER_SHADOW", "LAYER_BLUR", "BACKGROUND_BLUR"
	Color   *Color  `json:"color,omitempty"`
	Offset  *Offset `json:"offset,omitempty"`
	Radius  float64 `json:"radius,omitempty"`
	Spread  float64 `json:"spread,omitempty"`
	Visible bool    `json:"visible,omitempty"`
}

// Offset represents a 2D offset.
type Offset struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// LayoutGrid represents a layout grid (row, column, pixel).
type LayoutGrid struct {
	Type    string  `json:"type"` // "ROWS", "COLUMNS", "PIXEL"
	Color   *Color  `json:"color,omitempty"`
	Size    float64 `json:"size,omitempty"`
	Gutter  float64 `json:"gutter,omitempty"`
	Offset  float64 `json:"offset,omitempty"`
	Visible bool    `json:"visible,omitempty"`
}
