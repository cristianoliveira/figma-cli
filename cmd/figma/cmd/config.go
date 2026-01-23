package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configListCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long: `Manage configuration settings for the Figma CLI.

Configuration can be set via command line, environment variables, or config file.
Settings are loaded with the following precedence (highest to lowest):
1. CLI flags
2. Environment variables
3. JSON config file (~/.config/figma/config.json)
4. Default values

Use 'figma config list' to view all current settings.
Use 'figma config get <key>' to view a specific setting.
Use 'figma config set <key> <value>' to change a setting.

Examples:
  figma config get token
  figma config set output_format json
  figma config set api.timeout 60s
  figma config set export_dir ~/figma-exports`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := GetConfig(cmd)
		if err != nil {
			return err
		}

		// Output based on global output format
		outputFormat := cfg.OutputFormat
		if outputFormat == "" {
			outputFormat = "text"
		}

		switch outputFormat {
		case "json":
			return printConfigJSON(cmd, cfg)
		case "yaml":
			return printConfigYAML(cmd, cfg)
		default:
			return printConfigText(cmd, cfg)
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := GetConfig(cmd)
		if err != nil {
			return err
		}

		key := args[0]
		value, err := getConfigValue(cfg, key)
		if err != nil {
			return err
		}

		outputFormat := cfg.OutputFormat
		if outputFormat == "" {
			outputFormat = "text"
		}

		switch outputFormat {
		case "json":
			return printValueJSON(cmd, key, value)
		case "yaml":
			return printValueYAML(cmd, key, value)
		default:
			return printValueText(cmd, key, value)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value and save it to the config file.

The key can be a top-level field (token, output_format, export_dir) or a nested field
using dot notation (api.timeout, api.max_retries, logging.level).

Examples:
  figma config set token pat_xxx
  figma config set output_format json
  figma config set api.timeout 60s
  figma config set export_dir ~/figma-exports`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := GetConfig(cmd)
		if err != nil {
			return err
		}

		key := args[0]
		valueStr := args[1]

		// Parse value based on key type
		if err := setConfigValue(cfg, key, valueStr); err != nil {
			return err
		}

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		// Save to config file
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cmd.Printf("Updated %s = %s\n", key, valueStr)
		configPath, _ := config.ConfigPath()
		cmd.Println("Configuration saved to", configPath)

		// If setting token, also store in auth manager for secure storage
		if key == "token" || key == "token_type" {
			// TODO: integrate with auth manager
			cmd.PrintErrln("Note: For secure token storage, use 'figma auth login'.")
		}

		return nil
	},
}

// Helper functions (to be implemented)

func getConfigValue(cfg *config.Config, key string) (interface{}, error) {
	val, err := getFieldByKey(cfg, key)
	if err != nil {
		return nil, err
	}
	return val.Interface(), nil
}

func setConfigValue(cfg *config.Config, key, value string) error {
	field, err := getFieldByKey(cfg, key)
	if err != nil {
		return err
	}
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", key)
	}
	// Convert value based on field type
	val, err := parseValue(value, field.Type())
	if err != nil {
		return err
	}
	field.Set(val)
	return nil
}

func getFieldByKey(cfg *config.Config, key string) (reflect.Value, error) {
	parts := strings.Split(key, ".")
	current := reflect.ValueOf(cfg).Elem()
	for i, part := range parts {
		// Find field by JSON tag or field name
		field, err := findFieldByTag(current, part)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid key %s: %v", key, err)
		}
		if i == len(parts)-1 {
			return field, nil
		}
		if field.Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("key %s: %s is not a struct", key, part)
		}
		current = field
	}
	return reflect.Value{}, fmt.Errorf("unexpected error")
}

func findFieldByTag(v reflect.Value, key string) (reflect.Value, error) {
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("not a struct")
	}
	t := v.Type()
	// First try exact field name match (CamelCase)
	field := v.FieldByName(snakeToCamel(key))
	if field.IsValid() {
		return field, nil
	}
	// Try match JSON tag
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		tag := fieldType.Tag.Get("json")
		if tag == "" {
			continue
		}
		// Remove omitempty suffix
		if idx := strings.Index(tag, ","); idx != -1 {
			tag = tag[:idx]
		}
		if tag == key {
			return v.Field(i), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("field not found for key %s", key)
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func parseValue(s string, t reflect.Type) (reflect.Value, error) {
	// Special case for time.Duration
	if t == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(s)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid duration: %v", err)
		}
		return reflect.ValueOf(d), nil
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(s), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid integer: %v", err)
		}
		// Convert to exact type
		return reflect.ValueOf(i).Convert(t), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid unsigned integer: %v", err)
		}
		return reflect.ValueOf(u).Convert(t), nil
	case reflect.Bool:
		b := s == "true" || s == "1" || s == "yes"
		return reflect.ValueOf(b), nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("invalid float: %v", err)
		}
		return reflect.ValueOf(f).Convert(t), nil
	default:
		return reflect.Value{}, fmt.Errorf("unsupported type %v", t)
	}
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func printConfigText(cmd *cobra.Command, cfg *config.Config) error {
	keys := []string{
		"token",
		"token_type",
		"oauth_client_id",
		"oauth_scopes",
		"output_format",
		"export_dir",
		"api.timeout",
		"api.max_retries",
		"api.base_url",
		"api.debug",
		"api.tier",
		"api.seat_type",
		"logging.level",
		"logging.format",
		"logging.file",
		"logging.max_size",
		"logging.max_backups",
		"logging.max_age",
		"logging.compress",
	}

	for _, key := range keys {
		val, err := getConfigValue(cfg, key)
		if err != nil {
			// Skip invalid keys
			continue
		}
		// Mask token values
		if strings.Contains(key, "token") && val != nil {
			if s, ok := val.(string); ok {
				val = maskToken(s)
			}
		}
		cmd.Printf("%-30s = %v\n", key, val)
	}
	return nil
}

func printConfigJSON(cmd *cobra.Command, cfg *config.Config) error {
	// Convert config to map for masking
	var m map[string]interface{}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	// Mask token fields
	if token, ok := m["token"].(string); ok && token != "" {
		m["token"] = maskToken(token)
	}
	data, err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(data))
	return nil
}

func printConfigYAML(cmd *cobra.Command, cfg *config.Config) error {
	// Convert config to map for masking
	var m map[string]interface{}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	// Mask token fields
	if token, ok := m["token"].(string); ok && token != "" {
		m["token"] = maskToken(token)
	}
	data, err = yaml.Marshal(m)
	if err != nil {
		return err
	}
	cmd.Println(string(data))
	return nil
}

func printValueText(cmd *cobra.Command, key string, value interface{}) error {
	cmd.Printf("%s = %v\n", key, value)
	return nil
}

func printValueJSON(cmd *cobra.Command, key string, value interface{}) error {
	// Mask token values
	if strings.Contains(key, "token") {
		if s, ok := value.(string); ok {
			value = maskToken(s)
		}
	}
	m := map[string]interface{}{key: value}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	cmd.Println(string(data))
	return nil
}

func printValueYAML(cmd *cobra.Command, key string, value interface{}) error {
	// Mask token values
	if strings.Contains(key, "token") {
		if s, ok := value.(string); ok {
			value = maskToken(s)
		}
	}
	m := map[string]interface{}{key: value}
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	cmd.Println(string(data))
	return nil
}
