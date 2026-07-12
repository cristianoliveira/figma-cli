package imagecontext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey  string `json:"apiKey"`
	Model   string `json:"model"`
	BaseURL string `json:"baseUrl"`
}

func LoadConfig(modelOverride string) (Config, error) {
	path := os.Getenv("PI_SPECTACLES_CONFIG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, err
		}
		path = filepath.Join(home, ".pi", "agent", "pi-spectacles.json")
	}
	config := Config{}
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("read pi-spectacles config: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return Config{}, fmt.Errorf("invalid pi-spectacles config at %s: %w", path, err)
		}
	}
	if value := os.Getenv("OPENROUTER_API_KEY"); value != "" {
		config.APIKey = value
	}
	if value := os.Getenv("OPENROUTER_MEDIA_MODEL"); value != "" {
		config.Model = value
	}
	if value := os.Getenv("OPENROUTER_BASE_URL"); value != "" {
		config.BaseURL = value
	}
	if modelOverride != "" {
		config.Model = modelOverride
	}
	if config.Model == "" {
		config.Model = DefaultModel
	}
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	if config.APIKey == "" {
		return Config{}, fmt.Errorf("OpenRouter API key is required; configure %s or OPENROUTER_API_KEY", path)
	}
	return config, nil
}
