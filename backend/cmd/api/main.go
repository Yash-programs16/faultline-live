package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"faultline-live/backend/internal/api"
	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/faults"
	"faultline-live/backend/internal/store"
	"faultline-live/backend/internal/telemetry"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	st, err := store.New(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer st.Close()

	if err := st.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	collector := telemetry.NewCollector(
		cfg.Telemetry.WindowSize,
		cfg.Telemetry.BufferSize,
		st,
		cfg.Telemetry.PersistEnabled,
	)
	collector.StartPersister(ctx)
	defer collector.Stop()

	faultCtrl := faults.NewFaultController(
		faults.Config{
			MaxLatencyMs:       cfg.Faults.MaxLatencyMs,
			MaxFailureRate:     cfg.Faults.MaxFailureRate,
			MaxDurationSeconds: cfg.Faults.MaxDurationSeconds,
			LocalOnly:          cfg.Faults.LocalOnly,
		},
		st,
	)
	defer faultCtrl.Shutdown()

	handler := api.NewHandler(st, collector, faultCtrl, cfg)

	router := mux.NewRouter()
	router.Use(api.TelemetryMiddleware(collector))
	handler.RegisterRoutes(router)

	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}