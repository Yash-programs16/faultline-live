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
	"net/url"
	"time"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/models"
)

// GeminiClient runs incident investigations through Google's Generative
// Language API (Gemini). It is tried before Ollama: on Render/Vercel
// deployments there is no local Ollama daemon, so Gemini is the primary
// cloud investigator. When both are unavailable the handler falls back to
// the deterministic rule-based analysis in fallback.go.
type GeminiClient struct {
	apiKey     string
	model      string
	timeout    time.Duration
	maxRetries int
	enabled    bool
	client     *http.Client
	logger     *slog.Logger
}

type geminiGenerateRequest struct {
	SystemInstruction geminiContent   `json:"system_instruction"`
	Contents          []geminiContent `json:"contents"`
	GenerationConfig  geminiGenConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature      float64 `json:"temperature"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	ResponseMimeType string  `json:"responseMimeType"`
	TopP             float64 `json:"topP,omitempty"`
}

type geminiGenerateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Status  string `json:"status,omitempty"`
	} `json:"error,omitempty"`
}

func NewGeminiClient(cfg *config.GeminiConfig, logger *slog.Logger) *GeminiClient {
	return &GeminiClient{
		apiKey:     cfg.APIKey,
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

// Configured reports whether cloud investigations can be attempted.
func (c *GeminiClient) Configured() bool {
	return c != nil && c.enabled && c.apiKey != ""
}

// ModelName returns the configured Gemini model for status displays.
func (c *GeminiClient) ModelName() string {
	if c == nil {
		return ""
	}
	return c.model
}

// CheckReady validates the API key against the Generative Language API
// without running an investigation.
func (c *GeminiClient) CheckReady(ctx context.Context) error {
	if !c.Configured() {
		return errors.New("Gemini integration is not configured (GEMINI_API_KEY is missing or GEMINI_ENABLED=false)")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s?key=%s",
		url.PathEscape(c.model), url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create Gemini readiness request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Gemini API is unreachable: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("Gemini API rejected the request (HTTP %d): check GEMINI_API_KEY and model %q", resp.StatusCode, c.model)
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Gemini model %q was not found (HTTP 404)", c.model)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Gemini API returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// Investigate sends the shared incident prompt to Gemini and parses the
// JSON investigation from the response.
func (c *GeminiClient) Investigate(ctx context.Context, req InvestigationRequest) (*models.Investigation, error) {
	if !c.Configured() {
		return nil, fmt.Errorf("gemini integration not configured")
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	prompt := BuildInvestigationPrompt(req)
	body := geminiGenerateRequest{
		SystemInstruction: geminiContent{Parts: []geminiPart{{Text: SystemPromptText()}}},
		Contents:          []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
		GenerationConfig: geminiGenConfig{
			Temperature:      0.3,
			MaxOutputTokens:  2048,
			ResponseMimeType: "application/json",
			TopP:             0.9,
		},
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		text, err := c.generate(ctx, body)
		if err == nil {
			inv, perr := ParseInvestigationPayload(text, req.Incident.ID, "gemini:"+c.model)
			if perr == nil {
				return inv, nil
			}
			lastErr = perr
		} else {
			lastErr = err
		}
		if c.logger != nil {
			c.logger.Warn("gemini request failed, retrying", "attempt", attempt+1, "error", lastErr)
		}
		if attempt < c.maxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * time.Second):
			}
		}
	}
	return nil, fmt.Errorf("gemini investigation failed after %d attempt(s): %w", c.maxRetries+1, lastErr)
}

func (c *GeminiClient) generate(ctx context.Context, body geminiGenerateRequest) (string, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(c.model), url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Header-based auth as well as the key query parameter; the API
	// accepts either and some proxies strip query strings from logs.
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, 2<<20)
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(limited)
		return "", fmt.Errorf("gemini API returned HTTP %d: %s", resp.StatusCode, truncate(string(raw), 300))
	}

	var parsed geminiGenerateResponse
	if err := json.NewDecoder(limited).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini API error: %s", parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no candidates")
	}
	text := ""
	for _, p := range parsed.Candidates[0].Content.Parts {
		text += p.Text
	}
	if text == "" {
		return "", ErrInvalidResponse
	}
	return text, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
