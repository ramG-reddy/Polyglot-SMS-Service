package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Collection name in MongoDB
	SMSRecordsCollection = "sms_records"
)

var (
	// Client is the MongoDB client instance
	Client *mongo.Client
	// Database is the SMS Store database
	Database *mongo.Database
)

// InitMongoDB establishes connection to MongoDB with retry logic
func InitMongoDB(uri, dbName string) error {
	log.Println("Initializing MongoDB connection...")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri).
		SetMaxPoolSize(50).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Second).
		SetServerSelectionTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	Client = client
	Database = client.Database(dbName)

	log.Printf("Successfully connected to MongoDB database: %s", dbName)
	return nil
}

// ValidateIndexes verifies that indexes exist on the sms_records collection
// Indexes are created by MongoDB initialization script on first startup
func ValidateIndexes() error {
	log.Println("Verifying MongoDB indexes...")

	collection := Database.Collection(SMSRecordsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// List existing indexes to verify setup
	indexView := collection.Indexes()
	cursor, err := indexView.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list indexes: %w", err)
	}

	var existingIndexes []bson.M
	if err = cursor.All(ctx, &existingIndexes); err != nil {
		return fmt.Errorf("failed to decode indexes: %w", err)
	}

	// Verify expected indexes exist
	expectedIndexes := map[string]bool{
		"_id_":                     false,
		"idx_user_id":              false,
		"idx_created_at":           false,
		"idx_user_id_created_at":   false,
	}

	for _, idx := range existingIndexes {
		indexName := idx["name"].(string)
		if _, expected := expectedIndexes[indexName]; expected {
			expectedIndexes[indexName] = true
			log.Printf("✓ Index verified: %s", indexName)
		}
	}

	// Check if any expected indexes are missing
	missingIndexes := []string{}
	for indexName, found := range expectedIndexes {
		if !found && indexName != "_id_" {
			missingIndexes = append(missingIndexes, indexName)
		}
	}

	if len(missingIndexes) > 0 {
		log.Printf("WARNING: Missing indexes: %v", missingIndexes)
		log.Printf("Indexes should be created by MongoDB initialization script")
		// Don't fail - service can still work, just slower
	} else {
		log.Printf("✓ All indexes verified successfully (%d total)", len(existingIndexes))
	}

	return nil
}

// GetCollection returns the sms_records collection
func GetCollection() *mongo.Collection {
	return Database.Collection(SMSRecordsCollection)
}

// Close closes the MongoDB connection gracefully
func Close() error {
	if Client == nil {
		return nil
	}

	log.Println("Closing MongoDB connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("MongoDB connection closed successfully")
	return nil
}

// HealthCheck verifies MongoDB connection is alive
func HealthCheck() error {
	if Client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := Client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	return nil
}
