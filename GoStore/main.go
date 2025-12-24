package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ramG-reddy/sms-store/config"
	"github.com/ramG-reddy/sms-store/db"
	"github.com/ramG-reddy/sms-store/handlers"
	"github.com/ramG-reddy/sms-store/kafka"
)

func main() {
	log.Println("Starting SMS Store Service...")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded: Server Port=%s, MongoDB URI=%s", cfg.ServerPort, cfg.MongoURI)

	// Initialize MongoDB connection
	mongoClient, err := db.InitMongoDB(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}()
	log.Println("MongoDB connected successfully")

	// Create indexes
	if err := db.CreateIndexes(mongoClient, cfg.MongoDatabase); err != nil {
		log.Fatalf("Failed to create MongoDB indexes: %v", err)
	}
	log.Println("MongoDB indexes created successfully")

	// Start Kafka consumer in background goroutine
	go kafka.StartConsumer(cfg, mongoClient)
	log.Println("Kafka consumer started in background")

	// Setup HTTP routes
	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux, mongoClient, cfg)
	log.Println("HTTP routes registered")

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("SMS Store Service listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("SMS Store Service stopped")
}
