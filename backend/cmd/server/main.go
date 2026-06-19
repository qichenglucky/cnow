package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cnow/backend/internal/app"
	"cnow/backend/internal/config"
	"cnow/backend/internal/db"
	"cnow/backend/internal/pkg/audit"
	"cnow/backend/internal/pkg/idempotency"
	httphandler "cnow/backend/internal/platform/http"
	"cnow/backend/internal/repo"
	"cnow/backend/internal/workflow"
	"cnow/backend/internal/workflow/temporal"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	var log *zap.Logger
	var err error
	if cfg.Log.Format == "console" {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer log.Sync()

	log.Info("starting cnow server",
		zap.String("env", cfg.Env),
		zap.String("addr", cfg.HTTPAddr),
	)

	ctx := context.Background()

	// Database
	dbCfg := db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		MaxConns: cfg.DB.MaxConns,
	}
	pool, err := db.NewPool(ctx, dbCfg, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, dbCfg.ConnString(), "", log); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	// Ensure idempotency table exists
	if err := idempotency.EnsureTable(ctx, pool); err != nil {
		log.Fatal("failed to ensure idempotency table", zap.Error(err))
	}

	// Audit
	auditWriter := audit.NewWriter(pool, log, 4096)
	defer auditWriter.Close()

	// Repos
	serviceRepo := repo.NewServiceRepository(pool)
	releaseRepo := repo.NewReleaseRepository(pool)
	envRepo := repo.NewEnvironmentRepository(pool)
	eventRepo := repo.NewEventRepository(pool)

	// Repos (additional)
	approvalRepo := repo.NewApprovalRepository(pool)
	pipelineRepo := repo.NewPipelineRepository(pool)
	buildRepo := repo.NewBuildRepository(pool)

	// App layers
	serviceApp := app.NewServiceApp(serviceRepo, auditWriter, log)
	releaseApp := app.NewReleaseApp(releaseRepo, eventRepo, serviceRepo, envRepo, auditWriter, log)
	envApp := app.NewEnvironmentApp(envRepo, serviceRepo, log)
	approvalApp := app.NewApprovalApp(approvalRepo, releaseRepo, eventRepo, log)
	pipelineApp := app.NewPipelineApp(pipelineRepo, buildRepo, log)

	// Temporal workflow engine
	var workflowEngine workflow.Engine
	temporalCfg := temporal.LoadTemporalConfig()
	temporalClient, err := temporal.NewTemporalClient(temporalCfg, log)
	if err != nil {
		log.Warn("failed to connect to Temporal, running without workflow engine",
			zap.Error(err),
			zap.String("host", temporalCfg.HostPort),
		)
		workflowEngine = nil
	} else {
		activities := temporal.NewActivities(log)
		temporalWorker := temporal.NewWorker(temporalClient, temporalCfg.TaskQueue, activities)
		if err := temporalWorker.Start(); err != nil {
			log.Warn("failed to start Temporal worker", zap.Error(err))
		} else {
			log.Info("Temporal worker started",
				zap.String("task_queue", temporalCfg.TaskQueue),
				zap.String("namespace", temporalCfg.Namespace),
			)
		}
		defer temporalWorker.Stop()
		adapter := temporal.New(temporalClient, temporalCfg.TaskQueue, log)
		workflowEngine = adapter
		log.Info("Temporal workflow engine connected")
	}

	// HTTP server
	handler := httphandler.NewServer(pool, serviceApp, releaseApp, envApp, approvalApp, pipelineApp, workflowEngine, auditWriter, log)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("http server listening", zap.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("http server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Info("shutting down", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", zap.Error(err))
	}

	if temporalClient != nil {
		temporalClient.Close()
	}

	log.Info("server stopped")
}
