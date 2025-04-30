package main

import (
	"context"
	"errors"
	"flag"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/api"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/config"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/repository"
	"happy-fridge-rest-api/happy-fridge-rest-api/cmd/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "configs/config.toml", "Path to configuration file")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Printf("Warning: Could not load .env file: %v", err)
	}
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	logger.SetOutput(os.Stdout)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Fatalf("Invalid log level in config '%s': %v", cfg.LogLevel, err)
	}
	logger.SetLevel(logLevel)
	logger.Infof("Configuration loaded successfully. Log level set to '%s'.", cfg.LogLevel)

	store, err := repository.NewStore(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize database store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Errorf("Error closing database connection: %v", err)
		} else {
			logger.Info("Database connection closed.")
		}
	}()
	logger.Info("Repository store initialized.")

	services := service.NewService(store, cfg.JwtSecretKey, cfg.JwtExpirationTime)
	logger.Info("Services initialized.")

	router := api.NewRouterGin(logger, services, cfg.JwtSecretKey)
	logger.Info("HTTP Gin router initialized.")

	srv := &http.Server{
		Addr:         cfg.BindAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	serverStopped := make(chan struct{})

	go func() {
		logger.Infof("Starting Gin server on %s", cfg.BindAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Could not listen on %s: %v\n", cfg.BindAddr, err)
		}
		close(serverStopped)
		logger.Info("Server goroutine stopped.")
	}()

	sig := <-shutdownChan
	logger.Infof("Received signal: %s. Shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown Failed: %v", err)
	}

	<-serverStopped
	logger.Info("Server exited properly.")
}
