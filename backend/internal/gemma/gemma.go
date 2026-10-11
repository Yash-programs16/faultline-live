package gemma

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/models"
)

var (
	ErrOllamaUnavailable = errors.New("ollama service unavailable")
	ErrInvalidResponse   = errors.New("invalid response from Gemma")
	ErrTimeout           = errors.New("request timeout")
)

type GemmaClient struct {
	baseURL    string
	model      string
	timeout    time.Duration
	maxRetries int
	enabled    bool
	client     *http.Client
	logger     *slog.Logger
}

type ollamaChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Format   string    `json:"format"`
	Options  Options   `json:"options"`
}

type ollamaChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Options struct {
	Temperature float64 `json:"temperature"`
	NumPredict  int     `json:"num_predict"`
	TopP        float64 `json:"top_p"`
}

type InvestigationRequest struct {
	Incident        models.Incident
	EndpointMetrics map[string]models.MetricsResponse
	ActiveFaults    []models.Fault
	ServiceHealth   []ServiceHealth
}

type ServiceHealth struct {
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	ErrorRate float64 `json:"error_rate,omitempty"`
}

// CheckReady verifies both the Ollama HTTP endpoint and the configured model.
// A configured URL alone is not evidence that Gemma can run an investigation.
func (c *GemmaClient) CheckReady(ctx context.Context) error {
	if !c.enabled {
		return errors.New("Gemma integration is disabled")
	}

	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("create Ollama readiness request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama is unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&payload); err != nil {
		return fmt.Errorf("decode Ollama model list: %w", err)
	}
	for _, model := range payload.Models {
		if model.Name == c.model {
			return nil
		}
	}
	return fmt.Errorf("configured model %q is not installed in Ollama", c.model)
}

func NewGemmaClient(cfg *config.OllamaConfig, logger *slog.Logger) *GemmaClient {
	return &GemmaClient{
		baseURL:    cfg.BaseURL,
		model:      cfg.Model,
		timeout:    cfg.Timeout,
		maxRetries: cfg.MaxRetries,
		enabled:    cfg.Enabled,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

func (c *GemmaClient) Investigate(ctx context.Context, req InvestigationRequest) (*models.Investigation, error) {
	if !c.enabled {
		return nil, fmt.Errorf("gemma integration disabled")
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	prompt := c.buildPrompt(req)

	messages := []Message{
		{Role: "system", Content: c.systemPrompt()},
		{Role: "user", Content: prompt},
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		result, err := c.chat(ctx, messages)
		if err == nil {
			return c.parseInvestigation(result, req.Incident.ID)
		}
		lastErr = err
		c.logger.Warn("gemma request failed, retrying", "attempt", attempt+1, "error", err)
		if attempt < c.maxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * time.Second):
			}
		}
	}

	return nil, fmt.Errorf("gemma investigation failed after %d retries: %w", c.maxRetries+1, lastErr)
}

func (c *GemmaClient) chat(ctx context.Context, messages []Message) (string, error) {
	reqBody := ollamaChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
		Format:   "json",
		Options: Options{
			Temperature: 0.3,
			NumPredict:  2048,
			TopP:        0.9,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

func (c *GemmaClient) systemPrompt() string {
	return SystemPromptText()
}

func (c *GemmaClient) buildPrompt(req InvestigationRequest) string {
	return BuildInvestigationPrompt(req)
}

func (c *GemmaClient) parseInvestigation(response, incidentID string) (*models.Investigation, error) {
	return ParseInvestigationPayload(response, incidentID, c.model)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
