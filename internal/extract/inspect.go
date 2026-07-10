package extract

// InspectOutput is a curated single-node summary, used by `figma inspect`.
type InspectOutput struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Type             string              `json:"type"`
	Text             string              `json:"text,omitempty"`
	ComponentID      string              `json:"componentId,omitempty"`
	ComponentSetID   string              `json:"componentSetId,omitempty"`
	Fills            []string            `json:"fills,omitempty"`
	Strokes          []string            `json:"strokes,omitempty"`
	StrokeWeight     float64             `json:"strokeWeight,omitempty"`
	StrokeAlign      string              `json:"strokeAlign,omitempty"`
	Opacity          *float64            `json:"opacity,omitempty"`
	CornerRadius     *float64            `json:"cornerRadius,omitempty"`
	Bounds           boundsOutput        `json:"bounds"`
	Layout           layoutOutput        `json:"layout,omitempty"`
	Typography       typographyOutput    `json:"typography,omitempty"`
	Effects          []effectOutput      `json:"effects,omitempty"`
	BackgroundColor  string              `json:"backgroundColor,omitempty"`
	StyleBindings    map[string]string   `json:"styleBindings,omitempty"`
	VariableBindings map[string][]string `json:"variableBindings,omitempty"`
}

// FindNodeByID searches a document tree for the node with targetID.
func FindNodeByID(value any, targetID string) map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	id, _ := object["id"].(string)
	if id == targetID {
		return object
	}
	children, ok := object["children"].([]any)
	if !ok {
		return nil
	}
	for _, child := range children {
		if found := FindNodeByID(child, targetID); found != nil {
			return found
		}
	}
	return nil
}

// NodeToInspectOutput builds an InspectOutput from a raw node map.
func NodeToInspectOutput(object map[string]any) InspectOutput {
	bg, _ := object["backgroundColor"].(map[string]any)
	bgColor := ""
	if bg != nil {
		bgColor = colorHexFromPaint(bg)
	}
	return InspectOutput{
		ID:               StringValue(object["id"]),
		Name:             StringValue(object["name"]),
		Type:             StringValue(object["type"]),
		Text:             StringValue(object["characters"]),
		ComponentID:      StringValue(object["componentId"]),
		ComponentSetID:   StringValue(object["componentSetId"]),
		Fills:            colorsFromPaints(object["fills"]),
		Strokes:          colorsFromPaints(object["strokes"]),
		StrokeWeight:     numberValue(object["strokeWeight"]),
		StrokeAlign:      StringValue(object["strokeAlign"]),
		Opacity:          optionalNumber(object["opacity"]),
		CornerRadius:     optionalNumber(object["cornerRadius"]),
		Bounds:           boundsFromValue(object["absoluteBoundingBox"]),
		Layout:           layoutFromObject(object),
		Typography:       typographyFromValue(object["style"]),
		Effects:          effectsFromValue(object["effects"]),
		BackgroundColor:  bgColor,
		StyleBindings:    styleBindingsFromValue(object["styles"]),
		VariableBindings: variableBindingsFromValue(object["boundVariables"]),
	}
}

func styleBindingsFromValue(value any) map[string]string {
	bindings := make(map[string]string)
	styles, _ := value.(map[string]any)
	for property, rawID := range styles {
		if id, ok := rawID.(string); ok && id != "" {
			bindings[property] = id
		}
	}
	return bindings
}

func variableBindingsFromValue(value any) map[string][]string {
	bindings := make(map[string][]string)
	variables, _ := value.(map[string]any)
	for property, rawBinding := range variables {
		ids := variableAliasIDs(rawBinding)
		if len(ids) > 0 {
			bindings[property] = ids
		}
	}
	return bindings
}

func variableAliasIDs(value any) []string {
	switch binding := value.(type) {
	case map[string]any:
		id, idOK := binding["id"].(string)
		typeName, typeOK := binding["type"].(string)
		if idOK && id != "" && typeOK && typeName == "VARIABLE_ALIAS" {
			return []string{id}
		}
	case []any:
		ids := make([]string, 0, len(binding))
		for _, item := range binding {
			ids = append(ids, variableAliasIDs(item)...)
		}
		return ids
	}
	return nil
}
