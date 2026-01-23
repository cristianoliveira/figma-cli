package cmd

// Command registration pattern:
// Subcommands are added to rootCmd in init() functions of separate files.
// Each subcommand file should have its own init() that calls rootCmd.AddCommand().
// This keeps the root command definition clean and allows subcommands to be
// organized by functionality.

import (
	"context"
	"errors"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "figma",
	Short: "Figma CLI tool for exploring and interacting with Figma designs",
	Long: `Figma CLI enables developers and AI agents to fetch, analyze, and understand
Figma design data directly from the command line.

It parses Figma URLs, retrieves design metadata, and provides structured JSON output
for easy integration with automation workflows.

The tool is built with a focus on being LLM-friendly, URL-based, and context-aware,
making it ideal for AI-assisted design review, code generation, and asset export workflows.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration from files and environment
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Merge CLI flag values into config
		flags := getCLIFlags(cmd)
		cfg.Merge(flags)

		// Apply additional adjustments from flags (e.g., debug)
		adjustConfig(cfg, cmd)

		// Initialize logger with config
		loggingCfg := cfg.LoggingConfig()
		logger, err := logging.NewLogger(loggingCfg)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		// Create a request ID for this invocation and build context
		ctx := cmd.Context()
		ctx = logging.NewContextWithRequestID(ctx)
		ctx = context.WithValue(ctx, configKey{}, cfg)
		ctx = context.WithValue(ctx, loggerKey{}, logger)
		cmd.SetContext(ctx)

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Sync logger if present
		if logger, ok := cmd.Context().Value(loggerKey{}).(logging.Logger); ok {
			logger.Sync()
		}
	},
}

// configKey, loggerKey, requestIDKey are private types for context keys.
type configKey struct{}
type loggerKey struct{}
type requestIDKey struct{}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes flags.
func init() {
	// Global flags
	rootCmd.PersistentFlags().String("token", "", "Figma personal access token (overrides FIGMA_ACCESS_TOKEN)")
	rootCmd.PersistentFlags().String("output", "", "Output format: json, yaml, text (default \"text\")")
	rootCmd.PersistentFlags().String("export-dir", "", "Directory for exported assets (default \".\")")
	rootCmd.PersistentFlags().Duration("api-timeout", 0, "Timeout for API requests (e.g., 30s)")
	rootCmd.PersistentFlags().Int("api-max-retries", -1, "Maximum retries for failed API requests")
	rootCmd.PersistentFlags().String("api-base-url", "", "Base URL for Figma API (default \"https://api.figma.com/v1\")")
	rootCmd.PersistentFlags().Bool("api-debug", false, "Enable debug logging for API requests")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logging (alias for --api-debug)")

	// Mark certain flags as deprecated or hidden if needed
	// rootCmd.PersistentFlags().MarkHidden("api-debug")
}

// adjustConfig applies CLI flag adjustments that aren't covered by CLIFlags.
func adjustConfig(cfg *config.Config, cmd *cobra.Command) {
	// Set logging level to debug if --debug flag is set
	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		cfg.Logging.Level = "debug"
		cfg.API.Debug = true
	}
}

// getCLIFlags extracts CLI flag values from the command and returns a config.CLIFlags.
func getCLIFlags(cmd *cobra.Command) *config.CLIFlags {
	token, _ := cmd.Flags().GetString("token")
	output, _ := cmd.Flags().GetString("output")
	exportDir, _ := cmd.Flags().GetString("export-dir")
	apiTimeout, _ := cmd.Flags().GetDuration("api-timeout")
	apiMaxRetries, _ := cmd.Flags().GetInt("api-max-retries")
	apiBaseURL, _ := cmd.Flags().GetString("api-base-url")
	apiDebug, _ := cmd.Flags().GetBool("api-debug")
	debug, _ := cmd.Flags().GetBool("debug")

	// If --debug is set, enable API debug as well
	if debug && !apiDebug {
		apiDebug = true
	}

	return &config.CLIFlags{
		Token:         token,
		OutputFormat:  output,
		ExportDir:     exportDir,
		APITimeout:    apiTimeout,
		APIMaxRetries: apiMaxRetries,
		APIBaseURL:    apiBaseURL,
		APIDebug:      apiDebug,
	}
}

// GetConfig retrieves the config from the command context.
func GetConfig(cmd *cobra.Command) (*config.Config, error) {
	cfg, ok := cmd.Context().Value(configKey{}).(*config.Config)
	if !ok {
		return nil, errors.New("config not found in context")
	}
	return cfg, nil
}

// GetLogger retrieves the logger from the command context.
func GetLogger(cmd *cobra.Command) (logging.Logger, error) {
	logger, ok := cmd.Context().Value(loggerKey{}).(logging.Logger)
	if !ok {
		return nil, errors.New("logger not found in context")
	}
	return logger, nil
}

// GetRequestContext retrieves the request context with request ID.
func GetRequestContext(cmd *cobra.Command) (context.Context, error) {
	return cmd.Context(), nil
}
