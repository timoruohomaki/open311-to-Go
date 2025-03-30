package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/timoruohomaki/open311-to-Go/config"
	"github.com/timoruohomaki/open311-to-Go/internal/api"
	"github.com/timoruohomaki/open311-to-Go/internal/repository"
	"github.com/timoruohomaki/open311-to-Go/pkg/logger"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "./config/config.json", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)

	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize loggers
	log, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	defer log.Close()

	// Initialize second logger for access log in Apache format
	apachelog, err := logger.New(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize Apache logger: %v\n", err)
		os.Exit(1)
	}

	defer apachelog.Close()

	// Initialize MongoDB connection
	db, err := repository.NewMongoDBConnection(cfg.MongoDB)

	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
		os.Exit(1)
	}

	log.Infof("Connected MongoDB database %s", cfg.MongoDB.Database)

	defer db.Disconnect()

	// Initialize API
	api := api.New(cfg, log, db)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      api.Handler(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited properly")
}
