// Package diff provides Figma document diff utilities.
package diff

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type TextNode struct {
	ID   string
	Name string
	Text string
}

type TextNodeOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
}

type ChangedTextOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

type TextOutput struct {
	Added   []TextNodeOutput    `json:"added"`
	Removed []TextNodeOutput    `json:"removed"`
	Changed []ChangedTextOutput `json:"changed"`
}

func Text(from []TextNode, to []TextNode) TextOutput {
	fromByID := map[string]TextNode{}
	toByID := map[string]TextNode{}

	for _, node := range from {
		fromByID[node.ID] = node
	}
	for _, node := range to {
		toByID[node.ID] = node
	}

	var diff TextOutput
	for id, fromNode := range fromByID {
		toNode, ok := toByID[id]
		if !ok {
			diff.Removed = append(diff.Removed, TextNodeOutput(fromNode))
			continue
		}
		if fromNode.Text != toNode.Text {
			diff.Changed = append(diff.Changed, ChangedTextOutput{ID: id, Name: toNode.Name, From: fromNode.Text, To: toNode.Text})
		}
	}
	for id, toNode := range toByID {
		if _, ok := fromByID[id]; !ok {
			diff.Added = append(diff.Added, TextNodeOutput(toNode))
		}
	}

	sort.Slice(diff.Added, func(i, j int) bool { return diff.Added[i].ID < diff.Added[j].ID })
	sort.Slice(diff.Removed, func(i, j int) bool { return diff.Removed[i].ID < diff.Removed[j].ID })
	sort.Slice(diff.Changed, func(i, j int) bool { return diff.Changed[i].ID < diff.Changed[j].ID })
	return diff
}

func BuildFileVersionURL(fileID string, nodeIDs []string, versionID string) (string, error) {
	apiURL := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileID)
	u, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("version", versionID)
	if len(nodeIDs) > 0 {
		q.Set("ids", strings.Join(nodeIDs, ","))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func ExtractTextNodes(value any) []TextNode {
	var nodes []TextNode
	walkTextNodes(value, &nodes)
	return nodes
}

func walkTextNodes(value any, nodes *[]TextNode) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}

	if object["type"] == "TEXT" {
		id, _ := object["id"].(string)
		name, _ := object["name"].(string)
		text, _ := object["characters"].(string)
		if id != "" {
			*nodes = append(*nodes, TextNode{ID: id, Name: name, Text: text})
		}
	}

	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkTextNodes(child, nodes)
	}
}

func FetchFigmaJSON(apiURL string, token string) (map[string]any, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Figma-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}
	return result, nil
}
