package cmd

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"
)

func addNodeIDFlag(command *cobra.Command, usage string) {
	command.Flags().String("id", "", usage)
	command.Flags().String("node", "", "alias for --id")
}

func explicitNodeIDFlag(cmd *cobra.Command) (string, error) {
	id, _ := cmd.Flags().GetString("id")
	node, _ := cmd.Flags().GetString("node")
	if id != "" && node != "" && id != node {
		return "", fmt.Errorf("--id and --node must match when both are provided")
	}
	if id != "" {
		return id, nil
	}
	return node, nil
}

func addNodeIDsFlag(command *cobra.Command, usage string) {
	command.Flags().StringArray("id", nil, usage)
	command.Flags().StringArray("node", nil, "alias for --id")
}

func explicitNodeIDsFlag(cmd *cobra.Command) ([]string, error) {
	ids, _ := cmd.Flags().GetStringArray("id")
	nodes, _ := cmd.Flags().GetStringArray("node")
	if len(ids) > 0 && len(nodes) > 0 && !slices.Equal(ids, nodes) {
		return nil, fmt.Errorf("--id and --node must match when both are provided")
	}
	if len(ids) > 0 {
		return ids, nil
	}
	return nodes, nil
}
