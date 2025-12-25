#!/bin/bash
# MongoDB initialization script
# This script creates a non-root user, collection, and indexes for the sms_store database
# Runs automatically on first container startup with environment variables
# Security: All inputs are properly quoted and validated

set -e

# Validate required environment variables
if [ -z "${MONGO_INITDB_DATABASE}" ] || [ -z "${MONGO_APP_USER}" ] || [ -z "${MONGO_APP_PASSWORD}" ]; then
  echo "ERROR: Required environment variables are not set"
  echo "Required: MONGO_INITDB_DATABASE, MONGO_APP_USER, MONGO_APP_PASSWORD"
  exit 1
fi

echo "========================================="
echo "Starting MongoDB initialization..."
echo "Database: ${MONGO_INITDB_DATABASE}"
echo "App User: ${MONGO_APP_USER}"
echo "========================================="

# Wait for MongoDB to be ready (with timeout)
MAX_RETRIES=5
RETRY_COUNT=0
until mongosh --quiet --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
  RETRY_COUNT=$((RETRY_COUNT+1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "ERROR: MongoDB did not become ready in time"
    exit 1
  fi
  echo "Waiting for MongoDB to be ready... (${RETRY_COUNT}/${MAX_RETRIES})"
  sleep 2
done

echo "MongoDB is ready. Creating user, collection, and indexes..."

# Execute MongoDB commands using mongosh with proper error handling
mongosh --quiet <<EOF
// Switch to the database
use ${MONGO_INITDB_DATABASE}

// Create application user with read/write permissions
try {
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
  print('✓ User created successfully')
} catch(e) {
  if (e.code === 51003) {
    print('⚠ User already exists, skipping...')
  } else {
    print('✗ Error creating user: ' + e.message)
    throw e
  }
}

// Create sms_records collection
try {
  db.createCollection('sms_records')
  print('✓ Collection created successfully')
} catch(e) {
  if (e.code === 48) {
    print('⚠ Collection already exists, skipping...')
  } else {
    print('✗ Error creating collection: ' + e.message)
    throw e
  }
}

// Create indexes with custom names (matching Go application expectations)
try {
  // Single field index on user_id
  db.sms_records.createIndex(
    { user_id: 1 },
    { name: 'idx_user_id' }
  )
  print('✓ Index idx_user_id created')
  
  // Single field index on created_at (descending for recent queries)
  db.sms_records.createIndex(
    { created_at: -1 },
    { name: 'idx_created_at' }
  )
  print('✓ Index idx_created_at created')
  
  // Compound index for efficient user queries sorted by time
  db.sms_records.createIndex(
    { user_id: 1, created_at: -1 },
    { name: 'idx_user_id_created_at' }
  )
  print('✓ Index idx_user_id_created_at created')
} catch(e) {
  if (e.code === 85 || e.code === 86) {
    print('⚠ Some indexes already exist, skipping...')
  } else {
    print('✗ Error creating indexes: ' + e.message)
    throw e
  }
}

// Verify setup
var indexCount = db.sms_records.getIndexes().length
print('✓ Total indexes: ' + indexCount)

EOF

EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
  echo "========================================="
  echo "MongoDB initialization completed successfully!"
  echo "✓ User: ${MONGO_APP_USER} (readWrite role)"
  echo "✓ Collection: sms_records"
  echo "✓ Indexes: idx_user_id, idx_created_at, idx_user_id_created_at"
  echo "========================================="
else
  echo "========================================="
  echo "ERROR: MongoDB initialization failed!"
  echo "========================================="
  exit $EXIT_CODE
fi

print('MongoDB initialization completed successfully')
EOF

echo "========================================="
echo "MongoDB initialization completed!"
echo "User '${MONGO_APP_USER}' created with readWrite permissions"
echo "Collection 'sms_records' created with indexes"
echo "Indexes: user_id, created_at, user_id+created_at"
echo "========================================="
