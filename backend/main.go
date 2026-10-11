package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"faultline-live/backend/internal/config"
	"faultline-live/backend/internal/database"
	"faultline-live/backend/internal/defenses"
	"faultline-live/backend/internal/faults"
	"faultline-live/backend/internal/gemma"
	"faultline-live/backend/internal/handlers"
	"faultline-live/backend/internal/incidents"
	"faultline-live/backend/internal/middleware"
	"faultline-live/backend/internal/repositories"
	"faultline-live/backend/internal/telemetry"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Apply pending SQL migrations first when AUTO_MIGRATE=true so hosted
	// deploys (Render) never boot against an empty database.
	ctx, lifecycleCancel := context.WithCancel(context.Background())
	defer lifecycleCancel()
	if cfg.AutoMigrate {
		logger.Info("running automatic database migrations")
		if err := database.RunMigrations(ctx, cfg.Database.URL); err != nil {
			logger.Error("automatic migrations failed", "error", err)
			os.Exit(1)
		}
		logger.Info("migrations up to date")
	}

	// Initialize database
	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connected", "max_conns", cfg.Database.MaxConns)

	// Initialize telemetry before repositories so DB-backed request handlers can report query latency.
	tel := telemetry.NewTelemetry(db, cfg.Telemetry.MetricsWindow, cfg.Telemetry.IncidentCooldown)
	defer tel.Stop()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db, tel)
	orderRepo := repositories.NewOrderRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize fault manager
	faultCfg := faults.FaultConfig{
		Enabled:      cfg.Faults.Enabled,
		MaxDuration:  cfg.Faults.MaxDuration,
		MaxLatencyMs: cfg.Faults.MaxLatencyMs,
		MaxErrorRate: cfg.Faults.MaxErrorRate,
	}
	faultManager := faults.NewFaultManager(db, tel, logger, faultCfg)
	faultManager.Start(ctx)

	// Initialize incident detector
	detector := incidents.NewIncidentDetector(db, tel, logger)
	detector.Start(ctx)

	// Initialize Gemma client
	gemmaClient := gemma.NewGemmaClient(&cfg.Ollama, logger)
	geminiClient := gemma.NewGeminiClient(&cfg.Gemini, logger)
	if geminiClient.Configured() {
		logger.Info("gemini investigator configured", "model", geminiClient.ModelName())
	} else {
		logger.Info("gemini investigator not configured; set GEMINI_API_KEY to enable cloud investigations")
	}

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productRepo)
	authHandler := handlers.NewAuthHandler(userRepo, &cfg.Auth)
	accountHandler := handlers.NewAccountHandler(accountRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, productRepo, accountRepo, db.Pool)
	defenseManager := defenses.NewManager()
	opsHandler := handlers.NewOpsHandlerWithDefenses(db, tel, detector, faultManager, defenseManager, gemmaClient, cfg)
	opsHandler.SetGeminiClient(geminiClient)
	defer opsHandler.Stop()

	// Setup router
	router := mux.NewRouter()

	// Global middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS(cfg.CORSOrigins))
	router.Use(middleware.Metrics(tel))

	// Register routes
	productHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(router)
	accountHandler.RegisterRoutes(router)
	orderHandler.RegisterRoutes(router)
	opsHandler.RegisterRoutes(router)

	// Apply fault injection middleware (after metrics so we can track it)
	// Defense middleware runs before faults so floods are shed with 429
	// instead of consuming backend work.
	router.Use(defenseManager.Middleware())
	router.Use(faultManager.Middleware())

	// Create server
	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: cfg.Ollama.Timeout + 15*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if cfg.AllowPublicDemo && cfg.DemoMode {
		logger.Warn("public demo mode enabled: ops/simulation endpoints are reachable without login; use only for throwaway demos")
	}
	logger.Info("cors origins", "origins", cfg.CORSOrigins)

	// Start server
	go func() {
		logger.Info("starting server", "addr", cfg.Server.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown
	lifecycleCancel()
	detector.Stop()
	faultManager.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}
