package telemetry

import (
	"context"
	"sync"
	"time"

	"faultline-live/backend/internal/store"
)

type Observation struct {
	Timestamp    time.Time
	RouteTemplate string
	Service      string
	StatusCode   int
	DurationMs   float64
	Success      bool
}

type AggregatedWindow struct {
	WindowStart      time.Time
	Service          string
	RequestCount     int64
	ErrorCount       int64
	ErrorRate        float64
	MeanLatencyMs    float64
	P95LatencyMs     float64
	RequestsPerSecond float64
	Durations        []float64
}

type Collector struct {
	mu           sync.RWMutex
	currentWindow *AggregatedWindow
	windowSize    time.Duration
	bufferSize   int
	buffer       []AggregatedWindow
	store        *store.Store
	persistEnabled bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

func NewCollector(windowSize time.Duration, bufferSize int, store *store.Store, persistEnabled bool) *Collector {
	c := &Collector{
		windowSize:     windowSize,
		bufferSize:     bufferSize,
		store:          store,
		persistEnabled: persistEnabled,
		buffer:         make([]AggregatedWindow, 0, bufferSize),
		stopCh:         make(chan struct{}),
	}
	c.startWindow()
	return c
}

func (c *Collector) startWindow() {
	now := time.Now().Truncate(c.windowSize)
	c.currentWindow = &AggregatedWindow{
		WindowStart: now,
		Durations:   make([]float64, 0, 1000),
	}
}

func (c *Collector) Record(obs Observation) {
	c.mu.Lock()
	defer c.mu.Unlock()

	obsTime := obs.Timestamp.Truncate(c.windowSize)
	if !obsTime.Equal(c.currentWindow.WindowStart) {
		c.finalizeCurrentWindow()
		c.startWindow()
	}

	c.currentWindow.RequestCount++
	if !obs.Success {
		c.currentWindow.ErrorCount++
	}
	c.currentWindow.Durations = append(c.currentWindow.Durations, obs.DurationMs)
}

func (c *Collector) finalizeCurrentWindow() {
	if c.currentWindow.RequestCount == 0 {
		return
	}

	durations := c.currentWindow.Durations
	if len(durations) > 0 {
		var sum float64
		for _, d := range durations {
			sum += d
		}
		c.currentWindow.MeanLatencyMs = sum / float64(len(durations))

		sorted := make([]float64, len(durations))
		copy(sorted, durations)
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i] > sorted[j] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		p95Index := int(float64(len(sorted)-1) * 0.95)
		if p95Index < 0 {
			p95Index = 0
		}
		c.currentWindow.P95LatencyMs = sorted[p95Index]
	}

	windowDuration := time.Since(c.currentWindow.WindowStart).Seconds()
	if windowDuration < 1 {
		windowDuration = 1
	}
	c.currentWindow.RequestsPerSecond = float64(c.currentWindow.RequestCount) / windowDuration
	c.currentWindow.ErrorRate = float64(c.currentWindow.ErrorCount) / float64(c.currentWindow.RequestCount)

	completed := AggregatedWindow{
		WindowStart:        c.currentWindow.WindowStart,
		Service:            c.currentWindow.Service,
		RequestCount:       c.currentWindow.RequestCount,
		ErrorCount:         c.currentWindow.ErrorCount,
		ErrorRate:          c.currentWindow.ErrorRate,
		MeanLatencyMs:      c.currentWindow.MeanLatencyMs,
		P95LatencyMs:       c.currentWindow.P95LatencyMs,
		RequestsPerSecond:  c.currentWindow.RequestsPerSecond,
	}

	c.buffer = append(c.buffer, completed)
	if len(c.buffer) > c.bufferSize {
		c.buffer = c.buffer[len(c.buffer)-c.bufferSize:]
	}

	if c.persistEnabled && c.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func(w AggregatedWindow) {
			defer cancel()
			_ = c.store.PersistTelemetryWindow(ctx, store.TelemetryWindow{
				WindowStart:       w.WindowStart,
				Service:           w.Service,
				RequestCount:      w.RequestCount,
				ErrorCount:        w.ErrorCount,
				ErrorRate:         w.ErrorRate,
				MeanLatencyMs:     w.MeanLatencyMs,
				P95LatencyMs:      w.P95LatencyMs,
				RequestsPerSecond: w.RequestsPerSecond,
			})
		}(completed)
	}
}

func (c *Collector) GetRecentWindows(service string, limit int) []AggregatedWindow {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []AggregatedWindow
	for i := len(c.buffer) - 1; i >= 0 && len(result) < limit; i-- {
		if c.buffer[i].Service == service {
			result = append(result, c.buffer[i])
		}
	}
	return result
}

func (c *Collector) GetCurrentFault() *store.FaultEvent {
	return nil
}

func (c *Collector) StartPersister(ctx context.Context) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.windowSize)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				c.mu.Lock()
				c.finalizeCurrentWindow()
				c.mu.Unlock()
				return
			case <-ticker.C:
				c.mu.Lock()
				c.finalizeCurrentWindow()
				c.mu.Unlock()
			}
		}
	}()
}

func (c *Collector) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

func NormalizeRoute(path string) string {
	if path == "/api/health" {
		return "/api/health"
	}
	if path == "/api/products" {
		return "/api/products"
	}
	if len(path) >= 14 && path[:14] == "/api/products/" {
		return "/api/products/{id}"
	}
	if path == "/api/cart/validate" {
		return "/api/cart/validate"
	}
	if path == "/api/checkout" {
		return "/api/checkout"
	}
	if path == "/api/metrics/live" {
		return "/api/metrics/live"
	}
	if path == "/api/incidents" {
		return "/api/incidents"
	}
	if path == "/api/demo/fault" {
		return "/api/demo/fault"
	}
	return "/api/other"
}

func ServiceFromRoute(route string) string {
	switch route {
	case "/api/checkout":
		return "checkout"
	case "/api/products", "/api/products/{id}":
		return "catalog"
	case "/api/cart/validate":
		return "cart"
	case "/api/health", "/api/metrics/live", "/api/incidents", "/api/demo/fault":
		return "api"
	default:
		return "api"
	}
}