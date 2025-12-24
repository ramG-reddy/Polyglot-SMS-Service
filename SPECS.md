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

**Dependencies to Add from Spring Initializr:**
1.  **Spring Web** - For creating REST endpoints (`spring-boot-starter-web`)
2.  **Spring for Apache Kafka** - For producing events (`spring-kafka`)
3.  **Spring Data Redis (Access+Driver)** - For Block List checks (`spring-boot-starter-data-redis`)
4.  **Spring Reactive Web** - For WebClient HTTP client (`spring-boot-starter-webflux`)
5.  **Lombok** - For boilerplate reduction

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