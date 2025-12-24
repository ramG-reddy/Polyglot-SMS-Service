# Project Tasks - Polyglot SMS Service

## Phase 1: Project Setup & Scaffolding

### 1.1 Java SMS Sender Service Scaffolding
- [x] Generate Spring Boot 3.5.9 project from Spring Initializr
- [x] Extract to `JavaSender/` directory
- [x] Verify `pom.xml` contains all required dependencies
- [x] Verify Maven 3.9+ compatibility
- [x] Create directory structure: `src/main/java/com/sms/sender/{controller,service,model,config,kafka}`

### 1.2 Go SMS Store Service Scaffolding
- [x] Create `GoStore/` directory
- [x] Initialize Go module: `go mod init github.com/yourname/sms-store`
- [x] Create directory structure: `{handlers,models,services,kafka,db,config}`
- [x] Create basic `main.go` skeleton

### 1.3 Docker & Infrastructure Setup
- [ ] Create `docker-compose.yml` with all 5 services (Kafka, Zookeeper, Redis, MongoDB, Java, Go)
- [ ] Configure Kafka v7.6.0
- [ ] Configure Zookeeper v7.6.0
- [ ] Configure Redis v7.2-alpine
- [ ] Configure MongoDB v6.0 with non-root user/password

---

## Phase 2: Java SMS Sender Service Implementation

### 2.1 Configuration
- [ ] Create `src/main/resources/application.yml`
- [ ] Configure Spring Boot properties (server port 8080)
- [ ] Configure Kafka bootstrap servers (kafka:9092)
- [ ] Configure Redis host/port (redis:6379)
- [ ] Set application name `sms-sender`

### 2.2 Models & Data Structures
- [ ] Create `SmsRequest` model (phoneNumber, message)
- [ ] Create `SmsResponse` model (status, message, timestamp)
- [ ] Create `KafkaEvent` model for Kafka messages

### 2.3 Redis Integration
- [ ] Create `RedisConfig` bean configuration
- [ ] Create `BlockListService` for checking Redis `blocked_users` key
- [ ] Implement startup logic to populate Redis with dummy blocked users
- [ ] Add logging for block list checks

### 2.4 Kafka Integration
- [ ] Create `KafkaProducerConfig` bean
- [ ] Create `SmsKafkaProducer` service class
- [ ] Configure Kafka topic: `sms.events`
- [ ] Implement message serialization to JSON

### 2.5 REST API & Business Logic
- [ ] Create `SmsController` for `POST /v0/sms/send` endpoint
- [ ] Create `SmsService` with core business logic:
  - Check Redis block list
  - Mock 3rd party API call (simulate latency & random 200/500 status)
  - Produce Kafka event
- [ ] Implement error handling and logging
- [ ] Add input validation for phoneNumber and message

### 2.6 Testing
- [ ] Create unit tests for `SmsService`
- [ ] Create unit tests for `BlockListService`
- [ ] Create unit tests for `SmsKafkaProducer`
- [ ] Create integration tests for REST endpoint

---

## Phase 3: Go SMS Store Service Implementation

### 3.1 Configuration
- [ ] Create config package with environment variable loading
- [ ] Configure MongoDB connection string (mongodb://user:password@mongodb:27017)
- [ ] Configure Kafka bootstrap servers
- [ ] Configure server port (e.g., 8090)

### 3.2 Models & Data Structures
- [ ] Create `SMSRecord` struct with BSON/JSON tags
- [ ] Create `ListMessagesResponse` model

### 3.3 MongoDB Integration
- [ ] Create `db/mongo.go` with MongoDB connection logic
- [ ] Create database initialization on startup
- [ ] Create indexes on `user_id` (phoneNumber) field
- [ ] Implement connection pooling

### 3.4 Kafka Consumer
- [ ] Create `kafka/consumer.go` background goroutine
- [ ] Configure Kafka consumer group for `sms.events` topic
- [ ] Implement message deserialization from JSON
- [ ] Implement database persistence logic
- [ ] Add error handling and retry logic

### 3.5 REST API
- [ ] Create `handlers/sms_handlers.go` for `GET /v0/user/{user_id}/messages` endpoint
- [ ] Implement message retrieval from MongoDB
- [ ] Return JSON array of `SMSRecord`
- [ ] Add input validation and error handling

### 3.6 Business Services
- [ ] Create `services/sms_service.go` for database operations
- [ ] Implement `GetMessagesByUserID()` function
- [ ] Implement `SaveMessage()` function
- [ ] Add logging throughout

### 3.7 Testing
- [ ] Create unit tests for handlers
- [ ] Create unit tests for services
- [ ] Create mock MongoDB client for testing
- [ ] Create mock Kafka consumer for testing

---

## Phase 4: Dockerization

### 4.1 Java Service Dockerfile
- [ ] Create `JavaSender/Dockerfile` with multi-stage build
- [ ] Base image: `maven:3.9-eclipse-temurin-21` for build
- [ ] Runtime image: `eclipse-temurin:21-jre-alpine`
- [ ] Copy built JAR and run Spring Boot app
- [ ] Expose port 8080
- [ ] Set health check

### 4.2 Go Service Dockerfile
- [ ] Create `GoStore/Dockerfile` with multi-stage build
- [ ] Base image: `golang:1.25-alpine` for build
- [ ] Runtime image: `alpine:latest`
- [ ] Copy compiled binary and run
- [ ] Expose port 8090
- [ ] Set health check

### 4.3 Docker Compose Configuration
- [ ] Add health checks for all services
- [ ] Set proper environment variables for Java service
- [ ] Set proper environment variables for Go service
- [ ] Define service dependencies (depends_on)
- [ ] Configure volumes for persistent data (MongoDB, Redis)
- [ ] Configure networks for inter-service communication

---

## Phase 5: Integration & Testing

### 5.1 End-to-End Testing
- [ ] Test Java service `/v0/sms/send` endpoint
- [ ] Verify Kafka event is produced
- [ ] Verify Go service consumes event
- [ ] Verify data is persisted in MongoDB
- [ ] Test Go service `/v0/user/{user_id}/messages` retrieval
- [ ] Test block list functionality

### 5.2 Error Scenarios
- [ ] Test timeout handling between services
- [ ] Test Redis connection failure
- [ ] Test Kafka connection failure
- [ ] Test MongoDB connection failure
- [ ] Test invalid input handling

### 5.3 Load & Performance
- [ ] Test with multiple concurrent requests
- [ ] Verify no data loss during high load
- [ ] Check latency metrics

---

## Phase 6: Documentation & Deliverables

### 6.1 README.md
- [ ] Create comprehensive README with:
  - Project overview
  - Architecture diagram (ASCII or link)
  - Prerequisites (Docker & Docker Compose)
  - Setup instructions
  - Running the services (`docker-compose up`)
  - API documentation for both services
  - Example curl requests
  - Troubleshooting section

### 6.2 Demonstration Script
- [ ] Create shell script or documented steps for end-to-end flow:
  1. Start Docker Compose
  2. Wait for services to be ready
  3. Send SMS via Java service
  4. Check logs for Kafka event
  5. Retrieve message from Go service
  6. Verify data in MongoDB

### 6.3 Code Documentation
- [ ] Add javadoc comments to Java code
- [ ] Add godoc comments to Go code
- [ ] Document all public functions/methods
- [ ] Add inline comments for complex logic

### 6.4 Project Documentation
- [ ] Update CONSTITUTION.md with any implementation changes
- [ ] Update SPECS.md with any clarifications
- [ ] Document environment variables used
- [ ] Document Kafka topics and schema

---

## Phase 7: Final Validation

### 7.1 Code Quality
- [ ] Run Java linter/formatter
- [ ] Run Go formatter (gofmt)
- [ ] Review error handling in all services
- [ ] Check for hardcoded values (should use config)

### 7.2 Containerization Quality
- [ ] Verify Docker images build without errors
- [ ] Check Docker image sizes are reasonable
- [ ] Verify no secrets in Docker images
- [ ] Test Docker Compose up/down cycles

### 7.3 Final Testing
- [ ] Complete end-to-end workflow test
- [ ] Verify all endpoints are accessible
- [ ] Verify all services communicate correctly
- [ ] Verify data persistence across container restarts

---

## Summary

**Total Tasks:** ~80
**Completed:** 0
**In Progress:** 0
**Remaining:** 80

### Priority Order:
1. Phase 1: Scaffolding (foundation)
2. Phase 2: Java implementation
3. Phase 3: Go implementation
4. Phase 4: Dockerization
5. Phase 5: Integration testing
6. Phase 6: Documentation
7. Phase 7: Final validation
