package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

const version = "0.1.0"

func main() {
	// Define CLI flags
	var (
		token         = flag.String("token", "", "Figma personal access token (overrides FIGMA_ACCESS_TOKEN)")
		outputFormat  = flag.String("output", "", "Output format: json, yaml, text (default \"text\")")
		exportDir     = flag.String("export-dir", "", "Directory for exported assets (default \".\")")
		apiTimeout    = flag.Duration("api-timeout", 0, "Timeout for API requests (e.g., 30s)")
		apiMaxRetries = flag.Int("api-max-retries", -1, "Maximum retries for failed API requests")
		apiBaseURL    = flag.String("api-base-url", "", "Base URL for Figma API (default \"https://api.figma.com/v1\")")
		apiDebug      = flag.Bool("api-debug", false, "Enable debug logging for API requests")
		showVersion   = flag.Bool("version", false, "Show version information")
		showHelp      = flag.Bool("help", false, "Show help")
	)

	// Parse flags; stop at first non-flag argument (command)
	flag.Parse()

	// Load configuration from files and environment
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Merge CLI flag values into config
	cfg.Merge(&config.CLIFlags{
		Token:         *token,
		OutputFormat:  *outputFormat,
		ExportDir:     *exportDir,
		APITimeout:    *apiTimeout,
		APIMaxRetries: *apiMaxRetries,
		APIBaseURL:    *apiBaseURL,
		APIDebug:      *apiDebug,
	})

	// Initialize logger with config
	loggingCfg := cfg.LoggingConfig()
	logger, err := logging.NewLogger(loggingCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create a request ID for this invocation
	ctx := logging.NewContextWithRequestID(context.Background())

	// Handle global flags
	if *showVersion {
		fmt.Printf("figma-cli version %s\n", version)
		return
	}
	if *showHelp {
		printHelp()
		return
	}

	// Determine command from remaining arguments
	args := flag.Args()
	if len(args) == 0 {
		logger.Error(ctx, "No command provided")
		printUsage()
		os.Exit(1)
	}

	command := args[0]
	logger.Info(ctx, "Command executed", logging.String("command", command))

	switch command {
	case "version", "--version", "-v":
		fmt.Printf("figma-cli version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	case "parse", "nodes", "text", "export":
		// Command not yet implemented
		logger.Error(ctx, "Command not yet implemented", logging.String("command", command))
		fmt.Printf("Command %q is not yet implemented.\n", command)
		os.Exit(1)
	default:
		logger.Error(ctx, "Unknown command", logging.String("command", command))
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	logger.Info(ctx, "Command completed")
}

func printUsage() {
	fmt.Println("Usage: figma-cli [global-options] <command> [command-options] <figma-url>")
	fmt.Println()
	fmt.Println("Global options:")
	flag.PrintDefaults()
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
	fmt.Println("  figma-cli --token $TOKEN parse https://...")
	fmt.Println()
	fmt.Println("Use 'figma-cli help <command>' for more information about a command.")
}

func printHelp() {
	printUsage()
	fmt.Println()
	fmt.Println("For detailed documentation, see:")
	fmt.Println("  https://github.com/cristianoliveira/figma-cli")
	fmt.Println()
	fmt.Println("Configuration sources (in order of increasing precedence):")
	fmt.Println("  1. Default values")
	fmt.Println("  2. .env file in current directory")
	fmt.Println("  3. JSON config file at ~/.config/figma/config.json")
	fmt.Println("  4. Environment variables")
	fmt.Println("  5. Command-line flags")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  FIGMA_ACCESS_TOKEN  Figma personal access token")
	fmt.Println("  OUTPUT_FORMAT       Output format (json, yaml, text)")
	fmt.Println("  EXPORT_DIR          Export directory")
	fmt.Println("  API_TIMEOUT         API timeout duration (e.g., 30s)")
	fmt.Println("  API_MAX_RETRIES     Maximum retries for API requests")
	fmt.Println("  API_BASE_URL        Base URL for Figma API")
	fmt.Println("  API_DEBUG           Enable API debug logging (true/false)")
	fmt.Println("  LOG_LEVEL           Log level (debug, info, warn, error)")
	fmt.Println("  LOG_FORMAT          Log format (json, text)")
	fmt.Println("  LOG_FILE            Log file path")
	fmt.Println("  LOG_MAX_SIZE        Max log file size in MB")
	fmt.Println("  LOG_MAX_BACKUPS     Max number of rotated logs")
	fmt.Println("  LOG_MAX_AGE         Max age of logs in days")
	fmt.Println("  LOG_COMPRESS        Compress rotated logs (true/false)")
}
