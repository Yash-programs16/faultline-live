package gemma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"faultline-live/backend/internal/models"
)

// SystemPromptText is the shared system instruction for every LLM provider
// (Ollama/Gemma and Gemini). A single source of truth keeps investigations
// consistent no matter which backend answers.
func SystemPromptText() string {
	return `You are an expert site reliability engineer analyzing production incidents. 
Your task is to analyze the provided telemetry data and evidence to determine the root cause of an incident.

You must respond with a valid JSON object matching the Investigation schema. Do not include any explanation outside the JSON.
IMPORTANT: Respond in English only. All summary, root_cause, hypotheses, evidence, steps, and explanations must be in English, regardless of any other language in the input.

Key principles:
1. Base conclusions ONLY on the provided evidence. Do not invent facts.
2. Distinguish between directly observed facts and hypotheses.
3. If evidence is insufficient, state so clearly and indicate what additional data would help.
4. Assign confidence scores honestly - low confidence when evidence is thin.
5. Provide actionable, safe recommendations.
6. Never suggest actions that could cause data loss or require manual database manipulation.`
}

// BuildInvestigationPrompt renders the incident evidence into the user prompt
// sent to any LLM provider.
func BuildInvestigationPrompt(req InvestigationRequest) string {
	incident := req.Incident

	var evidenceBuilder bytes.Buffer
	evidenceBuilder.WriteString("Known topology: one Go API → PostgreSQL; Go API → LLM for investigations. No distributed traces or log excerpts are supplied; acknowledge these limits. Severity assessment and affected components must be included in your summary.\n")
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

// ParseInvestigationPayload validates provider output against the
// Investigation schema. It is shared by every LLM provider so malformed
// model output is rejected identically everywhere.
func ParseInvestigationPayload(response, incidentID, modelName string) (*models.Investigation, error) {
	// A JSON decoder handles braces inside quoted strings correctly.
	start := strings.Index(response, "{")
	if start < 0 {
		return nil, ErrInvalidResponse
	}
	var raw json.RawMessage
	if json.NewDecoder(strings.NewReader(response[start:])).Decode(&raw) != nil {
		return nil, ErrInvalidResponse
	}
	jsonStr := string(raw)

	var parsed struct {
		Summary    string `json:"summary"`
		RootCause  string `json:"root_cause"`
		Hypotheses []struct {
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

	if parsed.Confidence < 0 || parsed.Confidence > 1 || parsed.ConfidenceExplanation == "" || len(parsed.RecommendedSteps) == 0 || len(parsed.RecoveryRecommendations) == 0 {
		return nil, ErrInvalidResponse
	}
	if parsed.MissingEvidence == nil {
		parsed.MissingEvidence = []string{}
	}

	hypotheses := make([]models.Hypothesis, len(parsed.Hypotheses))
	for i, h := range parsed.Hypotheses {
		if h.Rank < 1 || h.Hypothesis == "" || h.Likelihood < 0 || h.Likelihood > 1 || len(h.SupportingEvidence) == 0 {
			return nil, ErrInvalidResponse
		}
		if h.ContradictingEvidence == nil {
			h.ContradictingEvidence = []string{}
		}
		hypotheses[i] = models.Hypothesis{
			Rank:                  h.Rank,
			Hypothesis:            h.Hypothesis,
			Likelihood:            clamp(h.Likelihood, 0, 1),
			SupportingEvidence:    h.SupportingEvidence,
			ContradictingEvidence: h.ContradictingEvidence,
		}
	}

	return &models.Investigation{
		ID:                      models.NewInvestigationID(),
		IncidentID:              incidentID,
		Summary:                 parsed.Summary,
		RootCause:               parsed.RootCause,
		Hypotheses:              hypotheses,
		MissingEvidence:         parsed.MissingEvidence,
		RecommendedSteps:        parsed.RecommendedSteps,
		RecoveryRecommendations: parsed.RecoveryRecommendations,
		Confidence:              parsed.Confidence,
		ConfidenceExplanation:   parsed.ConfidenceExplanation,
		Model:                   modelName,
		CreatedAt:               time.Now(),
	}, nil
}
