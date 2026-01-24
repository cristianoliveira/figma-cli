package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

func init() {
	parseCmd.RunE = runParse
}

// parseOutput represents the structured output of the parse command.
type parseOutput struct {
	FileKey   string `json:"fileKey"`
	NodeID    string `json:"nodeId"`
	FileName  string `json:"fileName"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// runParse implements the parse command.
func runParse(cmd *cobra.Command, args []string) error {
	url := args[0]
	ctx := cmd.Context()

	// Get logger (optional)
	logger, err := GetLogger(cmd)
	if err != nil {
		logger = logging.NewNopLogger()
	}

	// Parse URL
	logger.Debug(ctx, "Parsing Figma URL", logging.String("url", url))
	parsed, err := figma.ParseURL(url)
	if err != nil {
		logger.Error(ctx, "Failed to parse Figma URL", logging.Err(err))
		return fmt.Errorf("invalid Figma URL: %w", err)
	}
	logger.Debug(ctx, "Parsed URL",
		logging.String("file_key", parsed.FileKey),
		logging.String("node_id", parsed.NodeID),
		logging.String("file_name", parsed.FileName),
		logging.String("version", parsed.Version),
		logging.String("timestamp", parsed.Timestamp),
	)

	// Build output
	output := parseOutput{
		FileKey:   parsed.FileKey,
		NodeID:    parsed.NodeID,
		FileName:  parsed.FileName,
		Version:   parsed.Version,
		Timestamp: parsed.Timestamp,
	}

	// Marshal JSON
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logger.Error(ctx, "Failed to marshal JSON output", logging.Err(err))
		return fmt.Errorf("failed to generate JSON output: %w", err)
	}

	// Output
	cmd.Println(string(data))
	return nil
}
