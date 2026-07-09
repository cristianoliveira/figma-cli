package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fixNulls(&root)
	out, err := yaml.Marshal(&root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(out)
}

func fixNulls(node *yaml.Node) {
	switch node.Kind {
	case yaml.MappingNode:
		// Find properties with type:'null' and convert to type:string + nullable:true
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			val := node.Content[i+1]
			if key.Value == "type" && val.Value == "null" {
				val.Value = "string"
				val.Tag = "!!str"
				// Add nullable: true after this pair
				nullKey := &yaml.Node{Kind: yaml.ScalarNode, Value: "nullable", Tag: "!!str"}
				nullVal := &yaml.Node{Kind: yaml.ScalarNode, Value: "true", Tag: "!!bool"}
				// Insert nullable pair right after type pair
				node.Content = append(node.Content[:i+2], append([]*yaml.Node{nullKey, nullVal}, node.Content[i+2:]...)...)
				i += 2 // skip the newly inserted pair
				continue
			}
			fixNulls(val)
		}
	case yaml.SequenceNode:
		// Remove sequence items that are just {type:'null'}
		var filtered []*yaml.Node
		for _, child := range node.Content {
			if child.Kind == yaml.MappingNode && isNullTypeMap(child) {
				continue
			}
			fixNulls(child)
			filtered = append(filtered, child)
		}
		node.Content = filtered
	default:
		for _, c := range node.Content {
			fixNulls(c)
		}
	}
}

func isNullTypeMap(node *yaml.Node) bool {
	if node.Kind != yaml.MappingNode || len(node.Content) != 2 {
		return false
	}
	return node.Content[0].Value == "type" && node.Content[1].Value == "null"
}
