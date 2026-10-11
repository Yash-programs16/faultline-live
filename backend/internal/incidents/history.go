package incidents

import (
	"context"
	"encoding/json"
	"faultline-live/backend/internal/models"
)

// Reload real incident records, investigations, and remediation evidence from
// existing tables. No storefront data or schema is changed.
func (d *IncidentDetector) loadHistory(ctx context.Context) {
	if d.db == nil {
		return
	}
	rows, err := d.db.Pool.Query(ctx, `SELECT id,type,severity,title,description,affected_endpoint,observed,baseline,threshold,evidence,status,started_at,resolved_at,created_at FROM incidents ORDER BY started_at DESC LIMIT 500`)
	if err != nil {
		d.logger.Warn("incident history unavailable")
		return
	}
	for rows.Next() {
		var inc models.Incident
		var observed, baseline, threshold, evidence []byte
		if rows.Scan(&inc.ID, &inc.Type, &inc.Severity, &inc.Title, &inc.Description, &inc.AffectedEndpoint, &observed, &baseline, &threshold, &evidence, &inc.Status, &inc.StartedAt, &inc.ResolvedAt, &inc.CreatedAt) != nil {
			continue
		}
		if json.Unmarshal(observed, &inc.Observed) != nil || json.Unmarshal(baseline, &inc.Baseline) != nil || json.Unmarshal(threshold, &inc.Threshold) != nil || json.Unmarshal(evidence, &inc.Evidence) != nil {
			continue
		}
		inc.LastObservedAt = inc.StartedAt
		inc.Actions = []models.Remediation{}
		d.mu.Lock()
		d.incidents[inc.ID] = &inc
		d.mu.Unlock()
	}
	if rows.Err() != nil {
		d.logger.Warn("incident history read incomplete")
	}
	rows.Close()
	findings, err := d.db.Pool.Query(ctx, `SELECT id,incident_id,summary,root_cause,hypotheses,missing_evidence,recommended_steps,recovery_recommendations,confidence,confidence_explanation,model,created_at FROM investigations ORDER BY created_at ASC`)
	if err == nil {
		for findings.Next() {
			var i models.Investigation
			var hypotheses []byte
			if findings.Scan(&i.ID, &i.IncidentID, &i.Summary, &i.RootCause, &hypotheses, &i.MissingEvidence, &i.RecommendedSteps, &i.RecoveryRecommendations, &i.Confidence, &i.ConfidenceExplanation, &i.Model, &i.CreatedAt) == nil && json.Unmarshal(hypotheses, &i.Hypotheses) == nil {
				d.SetInvestigation(i.IncidentID, &i)
			}
		}
		findings.Close()
	}
	audits, err := d.db.Pool.Query(ctx, `SELECT metadata FROM audit_logs WHERE action LIKE 'remediation.%' AND action <> 'remediation.requested' ORDER BY created_at ASC`)
	if err == nil {
		for audits.Next() {
			var data []byte
			var a models.Remediation
			if audits.Scan(&data) == nil && json.Unmarshal(data, &a) == nil {
				d.AddAction(a.IncidentID, a)
			}
		}
		audits.Close()
	}
}
