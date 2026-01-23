package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/logging"
)

const version = "0.1.0"

func main() {
	// Initialize logger
	logger := logging.Default()
	defer logger.Sync()

	// Create a request ID for this invocation
	ctx := logging.NewContextWithRequestID(context.Background())

	if len(os.Args) < 2 {
		logger.Error(ctx, "No command provided")
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	logger.Info(ctx, "Command executed", logging.String("command", command))

	switch command {
	case "version", "--version", "-v":
		fmt.Printf("figma-cli version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		logger.Error(ctx, "Unknown command", logging.String("command", command))
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	logger.Info(ctx, "Command completed")
}

func printUsage() {
	fmt.Println("Usage: figma-cli <command> [options] <figma-url>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  parse      Parse a Figma URL and extract metadata")
	fmt.Println("  nodes      Fetch node hierarchy for a design")
	fmt.Println("  text       Extract text layers from a design")
	fmt.Println("  export     Export assets from a frame")
	fmt.Println("  version    Show version information")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  figma-cli parse \"https://www.figma.com/design/...\"")
	fmt.Println("  figma-cli nodes --hierarchy \"https://www.figma.com/file/...\"")
	fmt.Println()
	fmt.Println("Use 'figma-cli help <command>' for more information about a command.")
}

func printHelp() {
	printUsage()
	fmt.Println()
	fmt.Println("For detailed documentation, see:")
	fmt.Println("  https://github.com/cristianoliveira/figma-cli")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Set FIGMA_ACCESS_TOKEN environment variable with your Figma personal access token.")
}
