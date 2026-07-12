package imagecontext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const DefaultOpenAIModel = "gpt-5.4-mini"
const DefaultOpenAIBaseURL = "https://api.openai.com/v1"

type OpenAI struct {
	apiKey, model, baseURL string
	client                 *http.Client
}

func NewOpenAI(apiKey, model, baseURL string) *OpenAI {
	return &OpenAI{apiKey: apiKey, model: model, baseURL: strings.TrimRight(baseURL, "/"), client: http.DefaultClient}
}
func (o *OpenAI) Describe(ctx context.Context, input Input) (Result, error) {
	ref, err := imageDataURL(input.ReferencePath)
	if err != nil {
		return Result{}, fmt.Errorf("read reference for visual context: %w", err)
	}
	actual, err := imageDataURL(input.ActualPath)
	if err != nil {
		return Result{}, fmt.Errorf("read actual for visual context: %w", err)
	}
	prompt := visualContextPrompt(input.Regions)
	payload := map[string]any{"model": o.model, "input": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "input_text", "text": prompt}, map[string]any{"type": "input_image", "image_url": ref}, map[string]any{"type": "input_image", "image_url": actual}}}}}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")
	response, err := o.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("request visual context: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	var envelope struct {
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return Result{}, fmt.Errorf("decode visual context response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if envelope.Error != nil {
			return Result{}, fmt.Errorf("OpenAI: %s", envelope.Error.Message)
		}
		return Result{}, fmt.Errorf("OpenAI returned HTTP %d", response.StatusCode)
	}
	content := ""
	for _, output := range envelope.Output {
		for _, part := range output.Content {
			if part.Type == "output_text" {
				content += part.Text
			}
		}
	}
	regions, err := parseRegionContexts(content, input.Regions)
	if err != nil {
		return Result{}, err
	}
	return Result{Provider: "openai", Model: o.model, Advisory: true, Regions: regions}, nil
}
