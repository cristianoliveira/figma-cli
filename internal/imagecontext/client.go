package imagecontext

import (
	"context"
	"fmt"
)

// Client adds advisory visual context without changing deterministic image evidence.
type Client interface {
	Describe(context.Context, Input) (Result, error)
}

func NewClient(provider string, config Config) (Client, error) {
	switch provider {
	case "openrouter":
		return NewOpenRouter(config.APIKey, config.Model, config.BaseURL), nil
	case "openai":
		return NewOpenAI(config.APIKey, config.Model, config.BaseURL), nil
	default:
		return nil, fmt.Errorf("unsupported visual context provider %q", provider)
	}
}
