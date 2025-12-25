# Technical Specifications & Environment Constraints

## 1. Global Environment (Docker)
The entire system must run in a dockerized environment. Use `docker-compose.yml`.

**IMPORTANT:** No local installation of JDK, Go, Maven, or any development tools is required. All development, building, and execution happens inside Docker containers.

### Infrastructure Versions
* **Kafka:** `confluentinc/cp-kafka:7.6.0` (Compatible with standard Kafka clients).
* **Zookeeper:** `confluentinc/cp-zookeeper:7.6.0` (Required for this version of Kafka).
* **Redis:** `redis:7.2-alpine` (Lightweight, stable).
* **MongoDB:** `mongo:6.0` (Standard LTS).
    * *Constraint:* Must be initialized with a non-root user/password.

---

## 2. Service 1: SMS Sender (Java)

### Core Stack
* **Language:** Java 21 (LTS).
* **Framework:** Spring Boot 3.5.9.
* **Build Tool:** Maven 3.9+.
* **Configuration Format:** YAML (`application.yml`)
* **Execution:** Will run in a Docker container.

### Configuration File (application.yml)
Java service configuration should be managed via `src/main/resources/application.yml`. Key configurations include:
* **Server Port:** `server.port: 8080`
* **Kafka Bootstrap Servers:** `spring.kafka.bootstrap-servers: kafka:9092`
* **Redis Connection:** `spring.data.redis.host: redis`, `spring.data.redis.port: 6379`
* **Application Name:** `spring.application.name: sms-sender`

### Dependency Requirements (pom.xml)

**Dependencies to Add from [Spring Initializer](https://start.spring.io/):**
1.  **Spring Web** - For creating REST endpoints (`spring-boot-starter-web`)
2.  **Spring for Apache Kafka** - For producing events (`spring-kafka`)
3.  **Spring Data Redis (Access+Driver)** - For Block List checks (`spring-boot-starter-data-redis`)
4.  **Spring Reactive Web** - For WebClient HTTP client (`spring-boot-starter-webflux`)
5.  **Lombok** - For boilerplate reduction
6.  **Validation** - For request payload validation (`spring-boot-starter-validation`)

**Automatically Included (No Need to Add):**
- `spring-boot-starter-test` - Included in every Spring Boot project by default
- `jackson-databind` - Included with Spring Web for JSON serialization

**Compatibility Notes:**
- Spring Boot 3.5.9 automatically manages compatible versions for all dependencies
- Spring Kafka v3.1.x+ is Java 21 compatible
- Redis uses Lettuce client v6.3.x+ (Java 21 compatible)
- All listed libraries are Java 21 (LTS) compatible

### API Contract (Java)
* **Endpoint:** `POST /v0/sms/send`
* **Payload:**
    ```json
    {
      "phoneNumber": "string",
      "message": "string"
    }
    ```
* **Logic:**
    * Check Redis key `blocked_users`.
    * Mock external call (simulate latency and random 200/500 status).
    * Produce message to Kafka topic `sms.events`.

---

## 3. Service 2: SMS Store (GoLang)

### Core Stack
* **Language:** Go 1.25.
* **Server:** Standard Library `net/http` (Strict Requirement: No external web frameworks like Gin/Echo).
* **Execution:** Will run in a Docker container.

### Module Requirements (go.mod)
* **Database:** `go.mongodb.org/mongo-driver` (v1.17.0+). *Go 1.25 Compatible: v1.17.x+ fully supports Go 1.25*.
* **Kafka Consumer:** `github.com/segmentio/kafka-go` (v0.4.47+) OR `github.com/confluentinc/confluent-kafka-go` (v2.6.0+).
    * *Note:* `segmentio/kafka-go` v0.4.47+ is Go 1.25 compatible and easier to setup without CGO dependencies in Docker.
    * *Note:* `confluent-kafka-go` v2.6.0+ supports Go 1.25 but requires librdkafka C library.

**Go 1.25 Compatibility Notes:**
- All standard library packages (`net/http`, `time`, `context`, etc.) are fully compatible with Go 1.25
- MongoDB driver v1.17.x+ is tested and compatible with Go 1.25
- Kafka client libraries are compatible with Go 1.25
- Recommended to use Go 1.25 for latest performance improvements and security patches

### Data Structures
* **Struct:** `SMSRecord`
    ```go
    type SMSRecord struct {
        ID          string    `bson:"_id,omitempty" json:"id"`
        UserID      string    `bson:"user_id" json:"user_id"`
        PhoneNumber string    `bson:"phone_number" json:"phone_number"`
        Message     string    `bson:"message" json:"message"`
        Status      string    `bson:"status" json:"status"` // SUCCESS/FAIL
        CreatedAt   time.Time `bson:"created_at" json:"created_at"`
    }
    ```
    *Requirement:* Must follow idiomatic Go struct tags.

### API Contract (Go)
* **Endpoint:** `GET /v0/user/{user_id}/messages` (where `user_id` is the phoneNumber).
* **Response:** JSON array of `SMSRecord`.

---

## 4. Development Workflow Instructions
1.  **Setup:** Run `docker-compose up -d` to start Redis, Mongo, and Kafka.
2.  **Java Init:** Populate Redis with dummy blocked users on startup.
3.  **Go Init:** Ensure MongoDB indexes are created for `user_id` (phoneNumber).
4.  **Inter-Service Logic:**
    * The Java service must define a `KafkaProducer` bean.
    * The Go service must run a background goroutine acting as the `KafkaConsumer`.

---

## 5. Implementation Clarifications

### 5.1 Kafka Message Format
* **Serialization**: JSON format for cross-language compatibility
* **Compression**: None (for simplicity with polyglot consumers)
* **Key**: null (round-robin distribution)
* **Value Schema**:
  ```json
  {
    "eventId": "UUID string",
    "userId": "E.164 phone number",
    "phoneNumber": "E.164 phone number",
    "message": "SMS content (1-160 chars)",
    "status": "SUCCESS|FAILED|BLOCKED",
    "createdAt": "ISO-8601 datetime (yyyy-MM-dd'T'HH:mm:ss)"
  }
  ```

### 5.2 Phone Number Format
* **Standard**: E.164 format (e.g., `+1234567890`)
* **Validation Regex**: `^\\+[1-9]\\d{1,14}$`
* **Length**: 1-15 digits after the '+' sign
* **Examples**: `+1234567890` (US), `+441234567890` (UK)

### 5.3 Message Length Constraints
* **Minimum**: 1 character
* **Maximum**: 160 characters (standard SMS length)
* **Validation**: Applied at Java service layer via `@Size(min=1, max=160)`

### 5.4 Status Values
* **SUCCESS**: SMS successfully sent by vendor API
* **FAILED**: Vendor API returned error
* **BLOCKED**: User found in Redis block list (SMS rejected)

### 5.5 Error Codes
* **400 Bad Request**: Invalid input (validation failure)
* **415 Unsupported Media Type**: Non-JSON content type
* **500 Internal Server Error**: Service/database/Kafka failure

### 5.6 Redis Block List
* **Key**: `blocked_users`
* **Data Structure**: Redis Set (SMEMBERS, SADD, SISMEMBER)
* **Initial Values**: `+1111111111`, `+2222222222`, `+3333333333`
* **Check**: O(1) membership test via `SISMEMBER`

### 5.7 MongoDB Schema
* **Database**: `sms_store`
* **Collection**: `sms_records`
* **Indexes**:
  - `idx_user_id`: Single field index on `user_id`
  - `idx_created_at`: Single field index on `created_at` (descending)
  - `idx_user_id_created_at`: Compound index on `(user_id, created_at DESC)`

### 5.8 Mock Vendor API Behavior
* **Latency**: Random delay between 100-500ms (configurable)
* **Failure Rate**: 30% chance of failure (configurable via `APP_MOCK_VENDOR_FAILURE_RATE`)
* **Success Response**: Simulated 200 OK
* **Failure Response**: Simulated 500 Internal Server Error

### 5.9 Kafka Consumer Behavior
* **Start Offset**: `FirstOffset` (reads from beginning for new consumer groups)
* **Commit Strategy**: Manual commit after successful MongoDB persistence
* **Error Handling**: Skip malformed messages; retry database errors without committing
* **Consumer Group**: `sms-store-consumer-group`

### 5.10 Docker Health Checks
* **Java Service**: Polls `/actuator/health` endpoint
* **Go Service**: Polls `/health` endpoint
* **Kafka**: Runs `kafka-broker-api-versions` command
* **MongoDB**: Runs `mongosh --eval "db.adminCommand('ping')"`
* **Redis**: Runs `redis-cli ping`
* **Intervals**: 30s for application services, 10s for infrastructure

---

## 6. Performance Targets

### 6.1 Latency
* **Java API Response**: < 500ms (95th percentile)
* **Go API Response**: < 100ms (95th percentile)
* **Kafka Produce**: < 50ms (95th percentile)
* **Kafka Consume + Persist**: < 200ms (95th percentile)

### 6.2 Throughput
* **Java Service**: 100+ requests/second
* **Go Service**: 500+ requests/second (read-only)
* **Kafka**: 1000+ messages/second

### 6.3 Resource Limits (Docker)
* **Java Service**: 1 CPU, 1GB RAM
* **Go Service**: 0.5 CPU, 256MB RAM
* **Kafka**: 2 CPU, 2GB RAM
* **MongoDB**: 1 CPU, 512MB RAM
* **Redis**: 0.25 CPU, 256MB RAM

---

## 7. Testing Requirements

### 7.1 Unit Testing
* **Java**: JUnit 5 + Mockito for service layer tests
* **Go**: Go testing package for handler and service tests
* **Coverage Target**: 70%+ for core business logic

### 7.2 Integration Testing
* **End-to-End Flow**: Send SMS → Verify Kafka → Verify MongoDB → Retrieve
* **Error Scenarios**: Test timeouts, failures, invalid inputs
* **Load Testing**: 50+ concurrent requests

### 7.3 Test Data
* **Valid Phone**: `+1234567890`
* **Blocked Phones**: `+1111111111`, `+2222222222`, `+3333333333`
* **Invalid Phones**: `invalid`, `123`, `+0000000000`

---

## 8. Deployment Configuration

### 8.1 Docker Compose Services
1. **zookeeper**: Kafka coordination (port 2181)
2. **kafka**: Message broker (ports 9092, 29092)
3. **redis**: Block list cache (port 6379)
4. **mongodb**: Persistence layer (port 27017)
5. **sms-sender**: Java service (port 8080)
6. **sms-store**: Go service (port 8090)

### 8.2 Service Dependencies
```
sms-sender depends_on: [kafka, redis]
sms-store depends_on: [kafka, mongodb]
kafka depends_on: [zookeeper]
```

### 8.3 Networks
* **polyglot-network**: Bridge network (subnet 172.25.0.0/16)

### 8.4 Volumes
* **redis-data**: Persistent Redis storage
* **mongodb-data**: Persistent MongoDB data
* **mongodb-config**: Persistent MongoDB configuration

---

**Last Updated**: December 26, 2025  
**Specification Version**: 1.1