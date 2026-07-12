package extract

// StyleBinding identifies a Figma style and includes metadata when available.
type StyleBinding struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// VariableBinding identifies a Figma variable and its collection when available.
type VariableBinding struct {
	ID             string `json:"id"`
	Name           string `json:"name,omitempty"`
	Type           string `json:"type,omitempty"`
	CollectionID   string `json:"collectionId,omitempty"`
	CollectionName string `json:"collectionName,omitempty"`
}

type relativeBoundsOutput struct {
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Width      float64 `json:"width,omitempty"`
	Height     float64 `json:"height,omitempty"`
	RelativeTo string  `json:"relativeTo"`
}

// InspectOutput is a curated single-node summary, used by `figma inspect`.
type InspectOutput struct {
	ID                  string                       `json:"id"`
	Name                string                       `json:"name"`
	Type                string                       `json:"type"`
	Text                string                       `json:"text,omitempty"`
	ComponentID         string                       `json:"componentId,omitempty"`
	ComponentSetID      string                       `json:"componentSetId,omitempty"`
	VariantProperties   map[string]any               `json:"variantProperties,omitempty"`
	ComponentProperties map[string]any               `json:"componentProperties,omitempty"`
	PropertyDefinitions map[string]any               `json:"propertyDefinitions,omitempty"`
	Fills               []string                     `json:"fills,omitempty"`
	Strokes             []string                     `json:"strokes,omitempty"`
	Paints              *paintsOutput                `json:"paints,omitempty"`
	StrokeWeight        float64                      `json:"strokeWeight,omitempty"`
	StrokeAlign         string                       `json:"strokeAlign,omitempty"`
	Opacity             *float64                     `json:"opacity,omitempty"`
	CornerRadius        *float64                     `json:"cornerRadius,omitempty"`
	Bounds              boundsOutput                 `json:"bounds"`
	RelativeBounds      *relativeBoundsOutput        `json:"relativeBounds,omitempty"`
	SpacingFromPrevious *LayoutSpacing               `json:"spacingFromPrevious,omitempty"`
	Layout              layoutOutput                 `json:"layout,omitempty"`
	Typography          typographyOutput             `json:"typography,omitempty"`
	StyleOverrideIDs    []int                        `json:"styleOverrideIds,omitempty"`
	StyleOverrides      map[string]typographyOutput  `json:"styleOverrides,omitempty"`
	Effects             []effectOutput               `json:"effects,omitempty"`
	BackgroundColor     string                       `json:"backgroundColor,omitempty"`
	StyleBindings       map[string]string            `json:"styleBindings,omitempty"`
	ResolvedStyles      map[string]StyleBinding      `json:"resolvedStyles,omitempty"`
	VariableBindings    map[string][]string          `json:"variableBindings,omitempty"`
	ResolvedVariables   map[string][]VariableBinding `json:"resolvedVariables,omitempty"`
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

// InspectTree returns full implementation specs in Figma tree order.
func InspectTree(value any) []InspectOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	outputs := []InspectOutput{NodeToInspectOutput(object)}
	children, _ := object["children"].([]any)
	var previousSibling map[string]any
	for _, child := range children {
		childObject, ok := child.(map[string]any)
		if !ok {
			continue
		}
		childOutputs := InspectTree(childObject)
		if len(childOutputs) == 0 {
			continue
		}
		childOutputs[0].SpacingFromPrevious = measureInspectSiblingSpacing(object, previousSibling, childObject)
		outputs = append(outputs, childOutputs...)
		if measurableInspectSibling(childObject) {
			previousSibling = childObject
		}
	}
	return outputs
}

func measureInspectSiblingSpacing(parent, previous, current map[string]any) *LayoutSpacing {
	spacing := measureSiblingSpacing(parent, previous, current)
	if spacing == nil {
		return nil
	}
	spacing.ParentID = StringValue(parent["id"])
	return spacing
}

func measurableInspectSibling(object map[string]any) bool {
	if object == nil || object["visible"] == false {
		return false
	}
	if StringValue(object["layoutPositioning"]) == layoutPositioningAbsolute {
		return false
	}
	_, ok := layoutBoundsFor(object)
	return ok
}

// InspectTreeRelativeToScope returns implementation specs with bounds relative to the scoped root.
func InspectTreeRelativeToScope(value any, scopeID string) []InspectOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	scopeBounds := boundsFromValue(object["absoluteBoundingBox"])
	outputs := InspectTree(value)
	for index := range outputs {
		outputs[index].RelativeBounds = relativeBounds(outputs[index].Bounds, scopeBounds, scopeID)
	}
	return outputs
}

func relativeBounds(bounds boundsOutput, scope boundsOutput, scopeID string) *relativeBoundsOutput {
	return &relativeBoundsOutput{
		X:          bounds.X - scope.X,
		Y:          bounds.Y - scope.Y,
		Width:      bounds.Width,
		Height:     bounds.Height,
		RelativeTo: scopeID,
	}
}

// NodeToInspectOutput builds an InspectOutput from a raw node map.
func NodeToInspectOutput(object map[string]any) InspectOutput {
	bg, _ := object["backgroundColor"].(map[string]any)
	bgColor := ""
	if bg != nil {
		bgColor = colorHexFromPaint(bg)
	}
	return InspectOutput{
		ID:                  StringValue(object["id"]),
		Name:                StringValue(object["name"]),
		Type:                StringValue(object["type"]),
		Text:                StringValue(object["characters"]),
		ComponentID:         StringValue(object["componentId"]),
		ComponentSetID:      StringValue(object["componentSetId"]),
		VariantProperties:   mapValue(object["variantProperties"]),
		ComponentProperties: mapValue(object["componentProperties"]),
		PropertyDefinitions: mapValue(object["componentPropertyDefinitions"]),
		Fills:               colorsFromPaints(object["fills"]),
		Strokes:             colorsFromPaints(object["strokes"]),
		Paints:              optionalPaintsFromObject(object),
		StrokeWeight:        numberValue(object["strokeWeight"]),
		StrokeAlign:         StringValue(object["strokeAlign"]),
		Opacity:             optionalNumber(object["opacity"]),
		CornerRadius:        optionalNumber(object["cornerRadius"]),
		Bounds:              boundsFromValue(object["absoluteBoundingBox"]),
		Layout:              layoutFromObject(object),
		Typography:          typographyFromValue(object["style"]),
		StyleOverrideIDs:    styleOverrideIDs(object["characterStyleOverrides"]),
		StyleOverrides:      textStyleOverrides(object["styleOverrideTable"]),
		Effects:             effectsFromValue(object["effects"]),
		BackgroundColor:     bgColor,
		StyleBindings:       styleBindingsFromValue(object["styles"]),
		VariableBindings:    variableBindingsFromValue(object["boundVariables"]),
	}
}

// ResolveInspectStyleBindings enriches raw style IDs without removing them.
func ResolveInspectStyleBindings(output *InspectOutput, styles map[string]map[string]any) {
	if len(output.StyleBindings) == 0 {
		return
	}
	output.ResolvedStyles = make(map[string]StyleBinding, len(output.StyleBindings))
	for property, id := range output.StyleBindings {
		metadata := styles[id]
		output.ResolvedStyles[property] = StyleBinding{
			ID:   id,
			Name: StringValue(metadata["name"]),
			Type: StringValue(metadata["styleType"]),
		}
	}
}

// ResolveInspectVariableBindings enriches raw variable IDs without removing them.
func ResolveInspectVariableBindings(output *InspectOutput, meta map[string]any) {
	if len(output.VariableBindings) == 0 {
		return
	}
	variables, _ := meta["variables"].(map[string]any)
	collections, _ := meta["variableCollections"].(map[string]any)
	output.ResolvedVariables = make(map[string][]VariableBinding, len(output.VariableBindings))
	for property, ids := range output.VariableBindings {
		resolved := make([]VariableBinding, 0, len(ids))
		for _, id := range ids {
			variable, _ := variables[id].(map[string]any)
			collectionID := StringValue(variable["variableCollectionId"])
			collection, _ := collections[collectionID].(map[string]any)
			resolved = append(resolved, VariableBinding{
				ID:             id,
				Name:           StringValue(variable["name"]),
				Type:           StringValue(variable["resolvedType"]),
				CollectionID:   collectionID,
				CollectionName: StringValue(collection["name"]),
			})
		}
		output.ResolvedVariables[property] = resolved
	}
}

func optionalPaintsFromObject(object map[string]any) *paintsOutput {
	paints := paintsFromObject(object)
	if len(paints.Fills) == 0 && len(paints.Strokes) == 0 {
		return nil
	}
	return &paints
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
