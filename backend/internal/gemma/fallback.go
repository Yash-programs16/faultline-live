package gemma

import (
	"fmt"
	"time"

	"faultline-live/backend/internal/models"
)

// FallbackModelPrefix marks investigations produced without Ollama.
// The frontend uses this prefix to render an "offline analysis" badge
// instead of an error.
const FallbackModel = "fallback-rules-v1 (offline)"

// BuildFallbackInvestigation returns a deterministic, rule-based analysis
// focused on exactly what the user asked for: WHAT failed and WHEN.
// It never calls Ollama, so the incident page can always render useful
// hardcoded data instead of "Investigation failed with Ollama".
func BuildFallbackInvestigation(inc models.Incident, activeFaults []models.Fault, ollamaErr string) *models.Investigation {
	when := inc.StartedAt.Format(time.RFC3339)
	what, where := fallbackWhatWhere(inc)

	summary := fmt.Sprintf("%s detected on %s at %s (severity: %s).", what, where, when, inc.Severity)
	rootCause := fallbackRootCause(inc, activeFaults)

	hypotheses := []models.Hypothesis{
		{
			Rank:       1,
			Hypothesis: rootCause,
			Likelihood: 0.75,
			SupportingEvidence: append(
				[]string{fmt.Sprintf("Incident fired at %s with type %q on %q.", when, inc.Type, inc.AffectedEndpoint)},
				inc.Evidence...,
			),
			ContradictingEvidence: []string{},
		},
		{
			Rank:                  2,
			Hypothesis:            "Elevated client load or retry storm is amplifying the underlying fault.",
			Likelihood:            0.35,
			SupportingEvidence:    []string{fmt.Sprintf("Telemetry window around %s should be compared against baseline.", when)},
			ContradictingEvidence: []string{"No per-client breakdown is collected in this demo build."},
		},
	}

	if len(activeFaults) > 0 {
		f := activeFaults[0]
		hypotheses[0].SupportingEvidence = append(hypotheses[0].SupportingEvidence,
			fmt.Sprintf("Active injected fault %q (%s on %s) overlaps this incident window.", f.ID, f.Type, f.TargetRoute))
	}

	missing := []string{
		"Per-client/IP request breakdown to confirm a single abusive source.",
		"Distributed traces or backend logs around " + when + ".",
	}
	steps := []string{
		fmt.Sprintf("Confirm what failed: inspect %s telemetry at %s (observed vs baseline on the incident page).", where, when),
		"Check Simulations panel for an active fault overlapping this time window and stop it if present.",
		"For HTTP-flood / traffic_spike incidents, enable the rate-limiter defense below, then re-check error rate over two detector cycles.",
	}
	recovery := fallbackRecovery(inc)

	explain := "Offline rule-based analysis from incident thresholds and evidence."
	if ollamaErr != "" {
		explain = "Ollama unavailable (" + ollamaErr + "). " + explain
	}

	return &models.Investigation{
		ID:                      models.NewInvestigationID(),
		IncidentID:              inc.ID,
		Summary:                 summary,
		RootCause:               rootCause,
		Hypotheses:              hypotheses,
		MissingEvidence:         missing,
		RecommendedSteps:        steps,
		RecoveryRecommendations: recovery,
		Confidence:              0.6,
		ConfidenceExplanation:   explain + " Confidence is capped because an LLM review was not possible.",
		Model:                   FallbackModel,
		CreatedAt:               time.Now(),
	}
}

func fallbackWhatWhere(inc models.Incident) (what, where string) {
	where = inc.AffectedEndpoint
	if where == "" {
		where = "unknown endpoint"
	}
	switch inc.Type {
	case "high_error_rate":
		what = "Elevated HTTP 5xx error rate"
	case "high_latency":
		what = "Elevated API latency (P95)"
	case "high_db_latency":
		what = "Elevated database latency (P95)"
	case "traffic_spike":
		what = "Unusual HTTP request volume (possible flood)"
	default:
		what = "Anomaly (" + inc.Type + ")"
		if inc.Title != "" {
			what = inc.Title
		}
	}
	return what, where
}

func fallbackRootCause(inc models.Incident, activeFaults []models.Fault) string {
	for _, f := range activeFaults {
		if f.TargetRoute != "" && (inc.AffectedEndpoint == "GET "+f.TargetRoute ||
			inc.AffectedEndpoint == f.TargetRoute ||
			inc.AffectedEndpoint == "global" || inc.AffectedEndpoint == "database") {
			switch f.Type {
			case "error":
				return fmt.Sprintf("Injected HTTP-error fault %q on %s is the most likely cause of the %s failure starting %s.",
					f.ID, f.TargetRoute, inc.Type, inc.StartedAt.Format(time.RFC3339))
			case "latency":
				return fmt.Sprintf("Injected latency fault %q on %s is the most likely cause of the %s failure starting %s.",
					f.ID, f.TargetRoute, inc.Type, inc.StartedAt.Format(time.RFC3339))
			case "db_failure":
				return fmt.Sprintf("Simulated database-failure fault %q on %s is the most likely cause of the %s failure starting %s.",
					f.ID, f.TargetRoute, inc.Type, inc.StartedAt.Format(time.RFC3339))
			}
		}
	}
	switch inc.Type {
	case "traffic_spike":
		return fmt.Sprintf("Request flood on %s starting %s exceeded the demo volume threshold; a rate limiter would shed the abusive load.",
			inc.AffectedEndpoint, inc.StartedAt.Format(time.RFC3339))
	case "high_error_rate":
		return fmt.Sprintf("Backend error budget on %s was exhausted starting %s; check active faults, then apply rate limiting to protect the endpoint while recovering.",
			inc.AffectedEndpoint, inc.StartedAt.Format(time.RFC3339))
	case "high_latency":
		return fmt.Sprintf("Downstream slowness on %s starting %s pushed P95 past the threshold; remove the latency source, then rate-limit to drain the queue.",
			inc.AffectedEndpoint, inc.StartedAt.Format(time.RFC3339))
	case "high_db_latency":
		return fmt.Sprintf("Database path slowness starting %s pushed DB P95 past the threshold; stop any db_failure simulation and throttle traffic to let the DB recover.",
			inc.StartedAt.Format(time.RFC3339))
	default:
		return fmt.Sprintf("Threshold breach on %s at %s; evidence points to the observed-vs-baseline gap shown on this page.",
			inc.AffectedEndpoint, inc.StartedAt.Format(time.RFC3339))
	}
}

func fallbackRecovery(inc models.Incident) []string {
	base := []string{
		"Stop or expire the overlapping injected fault (Simulations panel or Remediate action).",
	}
	switch inc.Type {
	case "traffic_spike":
		return []string{
			"Enable the rate-limiter defense on " + inc.AffectedEndpoint + " (e.g. 10 req/s, burst 20) to shed flood traffic.",
			"Stop the traffic/flood generator, then verify request rate falls below threshold for two detector cycles.",
		}
	case "high_error_rate":
		return append(base,
			"Enable the rate limiter on "+inc.AffectedEndpoint+" to cap abusive clients while the backend recovers.",
			"Verify error rate < 50% of threshold across two detector checks before resolving.",
		)
	case "high_latency", "high_db_latency":
		return append(base,
			"Temporarily enable rate limiting to drain queued requests and protect the slow dependency.",
			"Verify P95 < 80% of threshold across two detector checks before resolving.",
		)
	default:
		return append(base, "Re-check telemetry for two detector cycles to confirm recovery.")
	}
}
