package figma

import "github.com/cristianoliveira/figma-cli/internal/figma/api"

// PublishedComponent is the component metadata needed by component parity checks.
type PublishedComponent struct {
	Name   string
	NodeID string
}

// FetchPublishedComponents returns published components from a Figma main file.
func FetchPublishedComponents(client *Client, fileID string) ([]PublishedComponent, error) {
	var response api.GetFileComponentsResponse
	if err := client.Fetch(BuildFileComponentsURL(fileID), &response); err != nil {
		return nil, err
	}
	components := make([]PublishedComponent, 0, len(response.Meta.Components))
	for _, component := range response.Meta.Components {
		components = append(components, PublishedComponent{Name: component.Name, NodeID: component.NodeId})
	}
	return components, nil
}
