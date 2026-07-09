package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

type inspectOutput struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	Text            string           `json:"text,omitempty"`
	ComponentID     string           `json:"componentId,omitempty"`
	ComponentSetID  string           `json:"componentSetId,omitempty"`
	Fills           []string         `json:"fills,omitempty"`
	Strokes         []string         `json:"strokes,omitempty"`
	StrokeWeight    float64          `json:"strokeWeight,omitempty"`
	StrokeAlign     string           `json:"strokeAlign,omitempty"`
	Opacity         *float64         `json:"opacity,omitempty"`
	CornerRadius    *float64         `json:"cornerRadius,omitempty"`
	Bounds          boundsOutput     `json:"bounds"`
	Layout          layoutOutput     `json:"layout,omitempty"`
	Typography      typographyOutput `json:"typography,omitempty"`
	Effects         []effectOutput   `json:"effects,omitempty"`
	BackgroundColor string           `json:"backgroundColor,omitempty"`
}

var inspectCmd = &cobra.Command{
	Use:   "inspect [figma-url-or-file-id]",
	Short: "Show a curated summary of a specific Figma node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		if nodeID == "" {
			fmt.Fprintln(os.Stderr, "error: --id is required")
			os.Exit(1)
		}

		input, err := figma.ParseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			fmt.Fprintln(os.Stderr, "error: could not resolve node ID")
			os.Exit(1)
		}

		token, err := env.GetFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		apiURL, err := figma.BuildFileURL(input.FileID, nodeIDs, "", "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building URL: %v\n", err)
			os.Exit(1)
		}

		client := figma.NewClient(token)
		var resp api.GetFileResponse
		if err := client.Fetch(apiURL, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma node: %v\n", err)
			os.Exit(1)
		}
		doc, err := figma.UnmarshalDocument(resp.Document)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing document: %v\n", err)
			os.Exit(1)
		}

		found := findNodeByID(doc, nodeIDs[0])
		if found == nil {
			fmt.Fprintf(os.Stderr, "error: node %s not found\n", nodeIDs[0])
			os.Exit(1)
		}

		node := nodeToInspectOutput(found)
		output, err := json.MarshalIndent(node, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func findNodeByID(value any, targetID string) map[string]any {
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
		if found := findNodeByID(child, targetID); found != nil {
			return found
		}
	}
	return nil
}

func nodeToInspectOutput(object map[string]any) inspectOutput {
	bg, _ := object["backgroundColor"].(map[string]any)
	bgColor := ""
	if bg != nil {
		bgColor = colorHexFromPaint(bg)
	}
	return inspectOutput{
		ID:              figma.StringValue(object["id"]),
		Name:            figma.StringValue(object["name"]),
		Type:            figma.StringValue(object["type"]),
		Text:            figma.StringValue(object["characters"]),
		ComponentID:     figma.StringValue(object["componentId"]),
		ComponentSetID:  figma.StringValue(object["componentSetId"]),
		Fills:           colorsFromPaints(object["fills"]),
		Strokes:         colorsFromPaints(object["strokes"]),
		StrokeWeight:    numberValue(object["strokeWeight"]),
		StrokeAlign:     figma.StringValue(object["strokeAlign"]),
		Opacity:         optionalNumber(object["opacity"]),
		CornerRadius:    optionalNumber(object["cornerRadius"]),
		Bounds:          boundsFromValue(object["absoluteBoundingBox"]),
		Layout:          layoutFromObject(object),
		Typography:      typographyFromValue(object["style"]),
		Effects:         effectsFromValue(object["effects"]),
		BackgroundColor: bgColor,
	}
}

func init() {
	inspectCmd.Flags().String("id", "", "node ID to inspect; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(inspectCmd)
}
