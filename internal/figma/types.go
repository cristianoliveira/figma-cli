package figma

// TextNode represents a TEXT node in a Figma document.
type TextNode struct {
	ID   string
	Name string
	Text string
}

// TextNodeOutput is a JSON-serializable text node.
type TextNodeOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
}

// ChangedTextOutput represents a changed text node in a diff.
type ChangedTextOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

// TextOutput is the result of a text diff between two versions.
type TextOutput struct {
	Added   []TextNodeOutput    `json:"added"`
	Removed []TextNodeOutput    `json:"removed"`
	Changed []ChangedTextOutput `json:"changed"`
}

// LayerTextOutput represents a layer with its text content.
type LayerTextOutput struct {
	ID    string           `json:"id"`
	Name  string           `json:"name"`
	Type  string           `json:"type"`
	Texts []TextNodeOutput `json:"texts"`
}
