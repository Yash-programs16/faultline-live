package faults

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"faultline-live/backend/internal/store"
)

type Scenario string

const (
	ScenarioLatency  Scenario = "latency"
	ScenarioFailure  Scenario = "failure"
)

type FaultState struct {
	Scenario   Scenario         `json:"scenario"`
	Active     bool             `json:"active"`
	Params     map[string]any   `json:"params"`
	ActivatedAt time.Time       `json:"activated_at"`
	ExpiresAt  *time.Time      `json:"expires_at,omitempty"`
	FaultID    int64           `json:"fault_id"`
}

type Config struct {
	MaxLatencyMs      int
	MaxFailureRate    float64
	MaxDurationSeconds int
	LocalOnly         bool
}

type FaultController struct {
	mu           sync.RWMutex
	state        *FaultState
	config       Config
	store        *store.Store
	rand         *rand.Rand
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

func NewFaultController(config Config, store *store.Store) *FaultController {
	return &FaultController{
		config: config,
		store:  store,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		stopCh: make(chan struct{}),
	}
}

func (fc *FaultController) Activate(scenario Scenario, params map[string]any, durationSeconds int) (*FaultState, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fc.state != nil && fc.state.Active {
		return nil, errors.New("a fault is already active")
	}

	switch scenario {
	case ScenarioLatency:
		delayMs, ok := params["delay_ms"].(float64)
		if !ok {
			return nil, errors.New("latency scenario requires delay_ms parameter")
		}
		if delayMs < 100 || delayMs > float64(fc.config.MaxLatencyMs) {
			return nil, errors.New("delay_ms must be between 100 and max configured")
		}
	case ScenarioFailure:
		failureRate, ok := params["failure_rate"].(float64)
		if !ok {
			return nil, errors.New("failure scenario requires failure_rate parameter")
		}
		if failureRate <= 0 || failureRate > fc.config.MaxFailureRate {
			return nil, errors.New("failure_rate must be between 0 and max configured")
		}
	default:
		return nil, errors.New("unknown scenario")
	}

	if durationSeconds < 10 || durationSeconds > fc.config.MaxDurationSeconds {
		return nil, errors.New("duration_seconds out of allowed range")
	}

	now := time.Now()
	expiresAt := now.Add(time.Duration(durationSeconds) * time.Second)

	var faultID int64
	if fc.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		faultID, err = fc.store.RecordFaultActivation(ctx, string(scenario), params)
		if err != nil {
			return nil, err
		}
	}

	fc.state = &FaultState{
		Scenario:    scenario,
		Active:      true,
		Params:      params,
		ActivatedAt: now,
		ExpiresAt:   &expiresAt,
		FaultID:     faultID,
	}

	fc.wg.Add(1)
	go fc.autoExpire(faultID, expiresAt)

	return fc.state, nil
}

func (fc *FaultController) autoExpire(faultID int64, expiresAt time.Time) {
	defer fc.wg.Done()

	timer := time.NewTimer(time.Until(expiresAt))
	defer timer.Stop()

	select {
	case <-timer.C:
		fc.mu.Lock()
		if fc.state != nil && fc.state.Active && fc.state.FaultID == faultID {
			fc.deactivateLocked()
		}
		fc.mu.Unlock()
	case <-fc.stopCh:
		return
	}
}

func (fc *FaultController) Deactivate() (*FaultState, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fc.state == nil || !fc.state.Active {
		return nil, errors.New("no active fault to deactivate")
	}

	return fc.deactivateLocked(), nil
}

func (fc *FaultController) deactivateLocked() *FaultState {
	deactivated := fc.state
	fc.state.Active = false
	now := time.Now()
	fc.state.ExpiresAt = &now

	if fc.store != nil && fc.state.FaultID > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = fc.store.RecordFaultDeactivation(ctx, fc.state.FaultID)
	}

	fc.state = nil
	return deactivated
}

func (fc *FaultController) GetState() *FaultState {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	if fc.state == nil {
		return nil
	}

	stateCopy := *fc.state
	paramsCopy := make(map[string]any, len(fc.state.Params))
	for k, v := range fc.state.Params {
		paramsCopy[k] = v
	}
	stateCopy.Params = paramsCopy

	if fc.state.ExpiresAt != nil {
		expiresCopy := *fc.state.ExpiresAt
		stateCopy.ExpiresAt = &expiresCopy
	}

	return &stateCopy
}

func (fc *FaultController) ApplyLatency(ctx context.Context) error {
	fc.mu.RLock()
	state := fc.state
	fc.mu.RUnlock()

	if state == nil || !state.Active || state.Scenario != ScenarioLatency {
		return nil
	}

	delayMs := int(state.Params["delay_ms"].(float64))
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Duration(delayMs) * time.Millisecond):
		return nil
	}
}

func (fc *FaultController) ShouldFail() bool {
	fc.mu.RLock()
	state := fc.state
	fc.mu.RUnlock()

	if state == nil || !state.Active || state.Scenario != ScenarioFailure {
		return false
	}

	failureRate := state.Params["failure_rate"].(float64)
	return fc.rand.Float64() < failureRate
}

func (fc *FaultController) Shutdown() {
	close(fc.stopCh)
	fc.wg.Wait()

	fc.mu.Lock()
	if fc.state != nil && fc.state.Active && fc.store != nil && fc.state.FaultID > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = fc.store.RecordFaultDeactivation(ctx, fc.state.FaultID)
	}
	fc.mu.Unlock()
}

func (fc *FaultController) ValidateRequest(body map[string]any) (Scenario, map[string]any, int, error) {
	scenarioRaw, ok := body["scenario"].(string)
	if !ok {
		return "", nil, 0, errors.New("scenario is required")
	}

	scenario := Scenario(scenarioRaw)
	if scenario != ScenarioLatency && scenario != ScenarioFailure {
		return "", nil, 0, errors.New("invalid scenario")
	}

	paramsRaw, ok := body["params"].(map[string]any)
	if !ok {
		return "", nil, 0, errors.New("params is required")
	}

	var durationSeconds int
	if durRaw, ok := paramsRaw["duration_seconds"]; ok {
		switch d := durRaw.(type) {
		case float64:
			durationSeconds = int(d)
		case int:
			durationSeconds = d
		default:
			return "", nil, 0, errors.New("duration_seconds must be a number")
		}
		delete(paramsRaw, "duration_seconds")
	} else {
		durationSeconds = 60
	}

	return scenario, paramsRaw, durationSeconds, nil
}

func (fc *FaultController) StateToJSON(state *FaultState) ([]byte, error) {
	if state == nil {
		return json.Marshal(map[string]any{"active": false})
	}
	return json.Marshal(state)
}