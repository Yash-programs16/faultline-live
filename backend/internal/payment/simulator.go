package payment

import (
	"context"
	"errors"
	"time"

	"faultline-live/backend/internal/faults"
	"faultline-live/backend/internal/telemetry"
)

type PaymentResult struct {
	Success         bool
	ErrorMessage    string
	ProcessingTimeMs int64
}

type Simulator struct {
	faultCtrl *faults.FaultController
	telemetry *telemetry.Collector
}

func NewSimulator(faultCtrl *faults.FaultController, telemetry *telemetry.Collector) *Simulator {
	return &Simulator{
		faultCtrl: faultCtrl,
		telemetry: telemetry,
	}
}

func (s *Simulator) ProcessPayment(ctx context.Context) PaymentResult {
	start := time.Now()
	
	if s.faultCtrl != nil {
		if err := s.faultCtrl.ApplyLatency(ctx); err != nil {
			durationMs := time.Since(start).Milliseconds()
			s.recordPaymentObservation(ctx, durationMs, false)
			if err == context.DeadlineExceeded || err == context.Canceled {
				return PaymentResult{
					Success:         false,
					ErrorMessage:    "payment processing timeout",
					ProcessingTimeMs: durationMs,
				}
			}
			return PaymentResult{
				Success:         false,
				ErrorMessage:    err.Error(),
				ProcessingTimeMs: durationMs,
			}
		}

		if s.faultCtrl.ShouldFail() {
			durationMs := time.Since(start).Milliseconds()
			s.recordPaymentObservation(ctx, durationMs, false)
			return PaymentResult{
				Success:         false,
				ErrorMessage:    "payment processing failed (simulated)",
				ProcessingTimeMs: durationMs,
			}
		}
	}

	durationMs := time.Since(start).Milliseconds()
	s.recordPaymentObservation(ctx, durationMs, true)
	
	return PaymentResult{
		Success:          true,
		ProcessingTimeMs: durationMs,
	}
}

func (s *Simulator) recordPaymentObservation(ctx context.Context, durationMs int64, success bool) {
	if s.telemetry == nil {
		return
	}
	
	obs := telemetry.Observation{
		Timestamp:     time.Now(),
		RouteTemplate: "/api/payment",
		Service:       "payment",
		StatusCode:    200,
		DurationMs:    float64(durationMs),
		Success:       success,
	}
	if !success {
		obs.StatusCode = 502
	}
	s.telemetry.Record(obs)
}

var ErrPaymentTimeout = errors.New("payment processing timeout")
var ErrPaymentFailed = errors.New("payment processing failed (simulated)")