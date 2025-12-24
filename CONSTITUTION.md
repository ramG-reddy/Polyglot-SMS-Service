# Polyglot Distributed SMS Service: Project Constitution

## 1. Project Goal
To refactor a monolithic notification system into a distributed, polyglot microservices architecture. The system comprises a Java-based **SMS Sender** (Gateway) and a GoLang-based **SMS Store** (Persistence), communicating via both synchronous HTTP and asynchronous Kafka events.

## 2. Architecture & Roles

### Service A: SMS Sender (The Gateway)
* **Language:** Java (Spring Boot).
* **Responsibility:**
    * Act as the public entry point for SMS requests.
    * Validate users against a **Block List** stored in Redis.
    * Interface with a (mocked) 3rd Party SMS Vendor.
    * Propagate metadata to Service B via Kafka (for logging) and/or HTTP (direct communication).

### Service B: SMS Store (The Vault)
* **Language:** GoLang (Standard Lib + Mongo Driver).
* **Responsibility:**
    * Ingest SMS records.
    * Persist data into **MongoDB** (Note: Architecture diagram mentions MySQL, but requirements strictly mandate MongoDB).
    * Serve retrieval APIs for user message history.

### Infrastructure
* **Message Broker:** Apache Kafka.
* **Cache:** Redis (for user block list).
* **Database:** MongoDB.
* **Containerization:** All services (Java SMS Sender, Go SMS Store, Kafka, Redis, MongoDB, Zookeeper) will run in Docker containers via docker-compose.

## 3. The Laws of Data Flow
The system must adhere to the following strict operational flow:
1.  **Ingestion:** Client POSTs request to Java Service.
2.  **Validation:** Java Service checks Redis. If blocked -> Stop.
3.  **Execution:** Java Service calls 3rd Party API (Mock: Randomly return SUCCESS/FAIL).
4.  **Logging (Async):** Java Service produces an event to Kafka.
5.  **Persistence:** Go Service consumes the Kafka event AND/OR accepts synchronous HTTP calls to store the record in MongoDB.
6.  **Retrieval:** Client GETs history from Go Service.

## 4. Coding Standards & Compliance
* **No Frameworks for Go HTTP:** The Go service must use the standard `net/http` library for routing, not Gin or Echo.
* **Polyglot Persistence:** Java owns Redis; Go owns MongoDB.
* **Error Handling:** Both services must handle timeouts and 3rd party failures gracefully.
* **Testing:** Core business logic must be unit tested in both languages.

## 5. Deliverables 
1. API Documentation: A simple README.md file detailing the endpoints for both 
services and instructions on how to run them locally. 
2. Demonstration: A script or documented steps demonstrating the full, end-to-end flow: 
Call the Java service, check the GoLang service's logs for the internal call, and finally 
retrieve the record using the GoLang service's history API.

## 6. Recommended Practices 
1. Code Structure: Organize code logically (handlers, services, models/structs). 
2. Error Handling: Implement robust logging and error handling, especially for inter-service communication timeouts or failures. 
3. Testing: Include basic Unit Tests for core business logic in both services. 
4. Best Practices: Follow language-specific best practices (Spring Boot conventions for Java and idiomatic Go for the Go service). 