#!/bin/bash
# MongoDB initialization script
# This script creates a non-root user for the sms_store database
# Runs automatically on first container startup with environment variables

set -e

echo "========================================="
echo "Starting MongoDB initialization..."
echo "Database: ${MONGO_INITDB_DATABASE}"
echo "App User: ${MONGO_APP_USER}"
echo "========================================="

# Wait for MongoDB to be ready
until mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
  echo "Waiting for MongoDB to be ready..."
  sleep 2
done

echo "MongoDB is ready. Creating user and collections..."

# Execute MongoDB commands using mongosh with environment variables
mongosh <<EOF
// Switch to the sms_store database
use ${MONGO_INITDB_DATABASE}

// Create application user with read/write permissions
db.createUser({
  user: '${MONGO_APP_USER}',
  pwd: '${MONGO_APP_PASSWORD}',
  roles: [
    {
      role: 'readWrite',
      db: '${MONGO_INITDB_DATABASE}'
    }
  ]
})

// Create sms_records collection
db.createCollection('sms_records')

// Create index on user_id (phoneNumber) for efficient queries
db.sms_records.createIndex({ user_id: 1 })

// Create index on created_at for time-based queries
db.sms_records.createIndex({ created_at: -1 })

// Create compound index for better query performance
db.sms_records.createIndex({ user_id: 1, created_at: -1 })

print('MongoDB initialization completed successfully')
EOF

echo "========================================="
echo "MongoDB initialization completed!"
echo "User '${MONGO_APP_USER}' created with readWrite permissions"
echo "Collection 'sms_records' created with indexes"
echo "Indexes: user_id, created_at, user_id+created_at"
echo "========================================="
