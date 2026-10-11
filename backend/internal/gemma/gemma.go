package gemma

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	Incident       models.Incident
	EndpointMetrics map[string]models.MetricsResponse
	ActiveFaults   []models.Fault
	ServiceHealth  []ServiceHealth
}

type ServiceHealth struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs float64 `json:"latency_ms,omitempty"`
	ErrorRate float64 `json:"error_rate,omitempty"`
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
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	return nil, fmt.Errorf("gemma investigation failed after %d retries: %w", c.maxRetries+1, lastErr)
}

func (c *GemmaClient) chat(ctx context.Context, messages []Message) (string, error) {
	reqBody := ollamaChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
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
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

func (c *GemmaClient) systemPrompt() string {
	return `You are an expert site reliability engineer analyzing production incidents. 
Your task is to analyze the provided telemetry data and evidence to determine the root cause of an incident.

You must respond with a valid JSON object matching the Investigation schema. Do not include any explanation outside the JSON.

Key principles:
1. Base conclusions ONLY on the provided evidence. Do not invent facts.
2. Distinguish between directly observed facts and hypotheses.
3. If evidence is insufficient, state so clearly and indicate what additional data would help.
4. Assign confidence scores honestly - low confidence when evidence is thin.
5. Provide actionable, safe recommendations.
6. Never suggest actions that could cause data loss or require manual database manipulation.`
}

func (c *GemmaClient) buildPrompt(req InvestigationRequest) string {
	incident := req.Incident
	
	var evidenceBuilder bytes.Buffer
	evidenceBuilder.WriteString("## Incident Details\n")
	evidenceBuilder.WriteString(fmt.Sprintf("- **ID**: %s\n", incident.ID))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Type**: %s\n", incident.Type))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Severity**: %s\n", incident.Severity))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Title**: %s\n", incident.Title))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Description**: %s\n", incident.Description))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Affected Endpoint**: %s\n", incident.AffectedEndpoint))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Started At**: %s\n", incident.StartedAt.Format(time.RFC3339)))
	evidenceBuilder.WriteString(fmt.Sprintf("- **Status**: %s\n\n", incident.Status))

	evidenceBuilder.WriteString("## Observed Metrics\n")
	for k, v := range incident.Observed {
		evidenceBuilder.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
	}
	evidenceBuilder.WriteString("\n")

	evidenceBuilder.WriteString("## Baseline Metrics (Normal Behavior)\n")
	for k, v := range incident.Baseline {
		evidenceBuilder.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
	}
	evidenceBuilder.WriteString("\n")

	evidenceBuilder.WriteString("## Thresholds\n")
	for k, v := range incident.Threshold {
		evidenceBuilder.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
	}
	evidenceBuilder.WriteString("\n")

	evidenceBuilder.WriteString("## Collected Evidence\n")
	for _, e := range incident.Evidence {
		evidenceBuilder.WriteString(fmt.Sprintf("- %s\n", e))
	}
	evidenceBuilder.WriteString("\n")

	if len(req.EndpointMetrics) > 0 {
		evidenceBuilder.WriteString("## Endpoint-Level Metrics\n")
		for endpoint, m := range req.EndpointMetrics {
			evidenceBuilder.WriteString(fmt.Sprintf("### %s\n", endpoint))
			evidenceBuilder.WriteString(fmt.Sprintf("- Request Count: %d\n", m.RequestCount))
			evidenceBuilder.WriteString(fmt.Sprintf("- Error Count: %d\n", m.ErrorCount))
			evidenceBuilder.WriteString(fmt.Sprintf("- Error Rate: %.4f\n", m.ErrorRate))
			evidenceBuilder.WriteString(fmt.Sprintf("- Latency P50: %.2fms\n", m.LatencyP50))
			evidenceBuilder.WriteString(fmt.Sprintf("- Latency P95: %.2fms\n", m.LatencyP95))
			evidenceBuilder.WriteString(fmt.Sprintf("- Latency P99: %.2fms\n", m.LatencyP99))
			evidenceBuilder.WriteString(fmt.Sprintf("- DB Latency P50: %.2fms\n", m.DBLatencyP50))
			evidenceBuilder.WriteString(fmt.Sprintf("- DB Latency P95: %.2fms\n\n", m.DBLatencyP95))
		}
	}

	if len(req.ActiveFaults) > 0 {
		evidenceBuilder.WriteString("## Active Fault Injection\n")
		for _, f := range req.ActiveFaults {
			evidenceBuilder.WriteString(fmt.Sprintf("- **%s** on %s: %v\n", f.Type, f.TargetRoute, f.Parameters))
		}
		evidenceBuilder.WriteString("\n")
	}

	if len(req.ServiceHealth) > 0 {
		evidenceBuilder.WriteString("## Service Health\n")
		for _, s := range req.ServiceHealth {
			evidenceBuilder.WriteString(fmt.Sprintf("- **%s**: %s (latency: %.0fms, error_rate: %.4f)\n", s.Name, s.Status, s.LatencyMs, s.ErrorRate))
		}
		evidenceBuilder.WriteString("\n")
	}

	evidenceBuilder.WriteString(`## Required Output Format

Return a JSON object with the following structure:
{
  "summary": "Concise 1-2 sentence summary of the incident",
  "root_cause": "Most likely root cause based on evidence",
  "hypotheses": [
    {
      "rank": 1,
      "hypothesis": "Description of hypothesis",
      "likelihood": 0.85,
      "supporting_evidence": ["Evidence 1", "Evidence 2"],
      "contradicting_evidence": ["Evidence that contradicts"]
    }
  ],
  "missing_evidence": ["What additional data would help confirm/refute hypotheses"],
  "recommended_steps": ["Step 1", "Step 2", "Step 3"],
  "recovery_recommendations": ["Safe recovery action 1", "Safe recovery action 2"],
  "confidence": 0.75,
  "confidence_explanation": "Explanation of confidence level"
}

Constraints:
- likelihood must be between 0 and 1
- confidence must be between 0 and 1
- All arrays must have at least 1 item
- If evidence is insufficient, include a hypothesis with low likelihood and note missing evidence
- Do not suggest destructive actions in recovery_recommendations`)

	return evidenceBuilder.String()
}

func (c *GemmaClient) parseInvestigation(response, incidentID string) (*models.Investigation, error) {
	// Try to extract JSON from response
	start := -1
	for i, c := range response {
		if c == '{' {
			start = i
			break
		}
	}
	if start == -1 {
		return nil, ErrInvalidResponse
	}

	// Find matching closing brace
	depth := 0
	end := -1
	for i := start; i < len(response); i++ {
		if response[i] == '{' {
			depth++
		} else if response[i] == '}' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}
	if end == -1 {
		return nil, ErrInvalidResponse
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		Summary                 string   `json:"summary"`
		RootCause               string   `json:"root_cause"`
		Hypotheses              []struct {
			Rank                  int      `json:"rank"`
			Hypothesis            string   `json:"hypothesis"`
			Likelihood            float64  `json:"likelihood"`
			SupportingEvidence    []string `json:"supporting_evidence"`
			ContradictingEvidence []string `json:"contradicting_evidence"`
		} `json:"hypotheses"`
		MissingEvidence         []string `json:"missing_evidence"`
		RecommendedSteps        []string `json:"recommended_steps"`
		RecoveryRecommendations []string `json:"recovery_recommendations"`
		Confidence              float64  `json:"confidence"`
		ConfidenceExplanation   string   `json:"confidence_explanation"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("parse investigation JSON: %w", err)
	}

	// Validate required fields
	if parsed.Summary == "" || parsed.RootCause == "" || len(parsed.Hypotheses) == 0 {
		return nil, ErrInvalidResponse
	}

	// Clamp confidence
	if parsed.Confidence < 0 {
		parsed.Confidence = 0
	}
	if parsed.Confidence > 1 {
		parsed.Confidence = 1
	}

	hypotheses := make([]models.Hypothesis, len(parsed.Hypotheses))
	for i, h := range parsed.Hypotheses {
		hypotheses[i] = models.Hypothesis{
			Rank:                  h.Rank,
			Hypothesis:            h.Hypothesis,
			Likelihood:            clamp(h.Likelihood, 0, 1),
			SupportingEvidence:    h.SupportingEvidence,
			ContradictingEvidence: h.ContradictingEvidence,
		}
	}

	return &models.Investigation{
		ID:                       models.NewInvestigationID(),
		IncidentID:               incidentID,
		Summary:                  parsed.Summary,
		RootCause:                parsed.RootCause,
		Hypotheses:               hypotheses,
		MissingEvidence:          parsed.MissingEvidence,
		RecommendedSteps:         parsed.RecommendedSteps,
		RecoveryRecommendations:  parsed.RecoveryRecommendations,
		Confidence:               parsed.Confidence,
		ConfidenceExplanation:    parsed.ConfidenceExplanation,
		Model:                    c.model,
		CreatedAt:                time.Now(),
	}, nil
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