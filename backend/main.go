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

	// Initialize database
	ctx := context.Background()
	db, err := database.New(ctx, &cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connected", "max_conns", cfg.Database.MaxConns)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	productRepo := repositories.NewProductRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	accountRepo := repositories.NewAccountRepository(db)

	// Initialize telemetry
	tel := telemetry.NewTelemetry(db, cfg.Telemetry.MetricsWindow, cfg.Telemetry.IncidentCooldown)

	// Initialize fault manager
	faultCfg := faults.FaultConfig{
		Enabled:       cfg.Faults.Enabled,
		MaxDuration:   cfg.Faults.MaxDuration,
		MaxLatencyMs:  cfg.Faults.MaxLatencyMs,
		MaxErrorRate:  cfg.Faults.MaxErrorRate,
	}
	faultManager := faults.NewFaultManager(db, tel, logger, faultCfg)

	// Initialize incident detector
	detector := incidents.NewIncidentDetector(db, tel, logger)
	detector.Start(ctx)

	// Initialize Gemma client
	gemmaClient := gemma.NewGemmaClient(&cfg.Ollama, logger)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productRepo)
	authHandler := handlers.NewAuthHandler(userRepo, &cfg.Auth)
	accountHandler := handlers.NewAccountHandler(accountRepo)
	orderHandler := handlers.NewOrderHandler(orderRepo, productRepo, accountRepo, db.Pool)
	opsHandler := handlers.NewOpsHandler(db, tel, detector, faultManager, gemmaClient, cfg)

	// Setup router
	router := mux.NewRouter()

	// Global middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.CORS([]string{"http://localhost:5173", "http://127.0.0.1:5173"}))
	router.Use(middleware.Metrics(tel))

	// Register routes
	productHandler.RegisterRoutes(router)
	authHandler.RegisterRoutes(router)
	accountHandler.RegisterRoutes(router)
	orderHandler.RegisterRoutes(router)
	opsHandler.RegisterRoutes(router)

	// Apply fault injection middleware (after metrics so we can track it)
	router.Use(faultManager.Middleware())

	// Create server
	server := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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