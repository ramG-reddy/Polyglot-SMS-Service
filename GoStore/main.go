package main

import (
	"context"
	"fmt"
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
	"github.com/ramG-reddy/sms-store/services"
)

func main() {
	log.Println("Starting SMS Store Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize MongoDB connection
	if err := db.InitMongoDB(cfg.MongoURI, cfg.MongoDatabase); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	// Verify indexes (created by MongoDB initialization script)
	if err := db.ValidateIndexes(); err != nil {
		log.Printf("Warning: Index validation failed: %v", err)
		// Continue anyway - indexes should exist from MongoDB init
	}

	// Initialize services
	smsService := services.NewSMSService()

	// Start Kafka consumer
	consumer, err := kafka.StartConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, smsService)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}
	defer consumer.Stop()

	// Setup HTTP handlers
	smsHandler := handlers.NewSMSHandler(smsService)

	http.HandleFunc("/v0/user/", smsHandler.GetUserMessages)
	http.HandleFunc("/health", smsHandler.HealthCheck)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "SMS Store Service - Use /v0/user/{user_id}/messages to retrieve messages")
	})

	// Start HTTP server
	serverAddr := ":" + cfg.ServerPort
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      nil,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("HTTP server listening on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
