# Technical Specifications & Environment Constraints

## 1. Global Environment (Docker)
The entire system must run in a dockerized environment. Use `docker-compose.yml`.

### Infrastructure Versions
* **Kafka:** `confluentinc/cp-kafka:7.6.0` (Compatible with standard Kafka clients).
* **Zookeeper:** `confluentinc/cp-zookeeper:7.6.0` (Required for this version of Kafka).
* **Redis:** `redis:7.2-alpine` (Lightweight, stable).
* **MongoDB:** `mongo:6.0` (Standard LTS).
    * *Constraint:* Must be initialized with a non-root user/password.

---

## 2. Service 1: SMS Sender (Java)

### Core Stack
* **Language:** Java 17 (LTS).
* **Framework:** Spring Boot 3.2.3.
* **Build Tool:** Maven 3.9+.
* **Execution:** Will run in a Docker container.

### Dependency Requirements (pom.xml)
1.  **Web:** `spring-boot-starter-web` (For creating REST endpoints).
2.  **Kafka:** `spring-kafka` (For producing events).
3.  **Redis:** `spring-boot-starter-data-redis` (For Block List checks).
4.  **HTTP Client:** `spring-boot-starter-webflux` (For `WebClient`) OR standard `RestTemplate`. *Preference: WebClient for better non-blocking support, though RestTemplate is acceptable per requirements*.
5.  **Utilities:** `lombok` (For boilerplate reduction), `jackson-databind` (JSON serialization).
6.  **Testing:** `spring-boot-starter-test`.

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
* **Language:** Go 1.22 (Latest Stable).
* **Server:** Standard Library `net/http` (Strict Requirement: No external web frameworks like Gin/Echo).
* **Execution:** Will run in a Docker container.

### Module Requirements (go.mod)
* **Database:** `go.mongodb.org/mongo-driver` (v1.13.0+).
* **Kafka Consumer:** `github.com/segmentio/kafka-go` (v0.4.47+) OR `github.com/confluentinc/confluent-kafka-go`.
    * *Note:* `segmentio` is often easier to setup without CGO dependencies in Docker.

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