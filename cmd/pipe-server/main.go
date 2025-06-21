package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"pipe-protocol/pkg/config"
	"pipe-protocol/pkg/server"
	"pipe-protocol/pkg/storage"
	"syscall"
	"time"
)

func main() {
	var configFile = flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	fileStorage := storage.NewFileStorage(cfg.Storage.Path)
	if err := fileStorage.Initialize(); err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create handler
	handler := server.NewPipeHandler(cfg.Server.ServerName, fileStorage)

	// Create server
	srv := server.NewServer(
		cfg.Server.Host,
		cfg.Server.Port,
		handler,
		time.Duration(cfg.Server.Timeout)*time.Second,
		cfg.Server.MaxConnections,
	)

	// Handle shutdown gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		srv.Stop()
		os.Exit(0)
	}()

	// Start server
	log.Printf("Starting PIPE server with config from %s", *configFile)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
