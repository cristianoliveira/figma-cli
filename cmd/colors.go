package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

type colorEntry struct {
	Color string   `json:"color"`
	Count int      `json:"count"`
	Usage []string `json:"usage"`
}

var colorsCmd = &cobra.Command{
	Use:   "colors [figma-url-or-file-id]",
	Short: "Extract the color palette from a Figma node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			fmt.Fprintln(os.Stderr, "error: colors requires --id or a Figma URL with node-id")
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

		palette := collectColors(doc)
		output, err := json.MarshalIndent(palette, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func collectColors(value any) []colorEntry {
	colorMap := map[string]*colorEntry{}
	walkColors(value, colorMap)

	entries := make([]colorEntry, 0, len(colorMap))
	for _, e := range colorMap {
		entries = append(entries, *e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if entries == nil {
		entries = []colorEntry{}
	}
	return entries
}

func walkColors(value any, colorMap map[string]*colorEntry) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}

	name := figma.StringValue(object["name"])

	collectFills := func(paints any) {
		items, ok := paints.([]any)
		if !ok {
			return
		}
		for _, p := range items {
			paint, ok := p.(map[string]any)
			if !ok {
				continue
			}
			if paint["visible"] == false {
				continue
			}
			c := colorHexFromPaint(paint)
			if c == "" || c == "#000000" {
				continue
			}
			entry, exists := colorMap[c]
			if !exists {
				entry = &colorEntry{Color: c, Usage: []string{}}
				colorMap[c] = entry
			}
			entry.Count++
			if len(entry.Usage) < 3 {
				entry.Usage = append(entry.Usage, name)
			}
		}
	}

	collectFills(object["fills"])
	collectFills(object["strokes"])

	if bg, ok := object["backgroundColor"].(map[string]any); ok {
		c := colorHexFromPaint(bg)
		if c != "" && c != "#000000" {
			entry, exists := colorMap[c]
			if !exists {
				entry = &colorEntry{Color: c, Usage: []string{}}
				colorMap[c] = entry
			}
			entry.Count++
			if len(entry.Usage) < 3 {
				entry.Usage = append(entry.Usage, name+" (bg)")
			}
		}
	}

	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkColors(child, colorMap)
	}
}

func init() {
	colorsCmd.Flags().String("id", "", "node ID to extract colors from; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(colorsCmd)
}
