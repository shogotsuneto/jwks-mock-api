package main

import (
	"flag"

	"github.com/shogotsuneto/jwks-mock-api/internal/server"
	"github.com/shogotsuneto/jwks-mock-api/pkg/config"
	"github.com/shogotsuneto/jwks-mock-api/pkg/logger"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with configured level
	logger.Init(cfg.LogLevel)
	logger.Debugf("Logger initialized with level: %s", cfg.LogLevel)

	// Create and start server
	srv, err := server.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
