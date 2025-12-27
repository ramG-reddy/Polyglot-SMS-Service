# Polyglot SMS Service

A distributed microservices architecture demonstrating polyglot persistence and asynchronous communication between Java and Go services.

## 🏗️ Architecture

```
┌─────────────┐    HTTP    ┌──────────────────┐    Kafka     ┌─────────────────┐
│   Client    │──────────► │  Java SMS Sender │─────────────►│  Go SMS Store   │
└─────────────┘            │  (Spring Boot)   │              │  (Standard Lib) │
                           └──────────────────┘              └─────────────────┘
                                    │ │                              │
                                    │ │ Redis                        │ MongoDB
                                    │ ▼                              ▼
                                    │ ┌──────────────┐    ┌─────────────────────┐
                                    │ │  Block List  │    │  SMS Records Store  │
                                    │ └──────────────┘    └─────────────────────┘
                                    │
                                    │ HTTP (Mock)
                                    ▼
                            ┌──────────────────┐
                            │ 3rd Party Vendor │
                            │    SMS API       │
                            └──────────────────┘
```

### Services

- **Java SMS Sender** (Port 8080): Gateway service that validates users against a Redis block list, mocks vendor API calls, and publishes events to Kafka
- **Go SMS Store** (Port 8090): Persistence service that consumes Kafka events and stores SMS records in MongoDB

### Infrastructure

- **Kafka + Zookeeper**: Asynchronous message broker for event streaming
- **Redis**: Cache for user block list (O(1) lookups)
- **MongoDB**: Document database for SMS record persistence

---

## 🚀 Quick Start

### Prerequisites

- Docker Desktop (Windows/Mac/Linux)
- 8GB RAM available for containers
- ~5GB disk space

### Run the System

```powershell
# Start all services
docker compose up -d

# Check service health (wait ~30-60 seconds)
docker compose ps

# View logs
docker compose logs -f
```

### Send an SMS

```powershell
curl -X POST http://localhost:8080/v0/sms/send `
  -H "Content-Type: application/json" `
  -d '{\"phoneNumber\": \"+1234567890\", \"message\": \"Hello World!\"}'
```

### Retrieve Messages

```powershell
curl http://localhost:8090/v0/user/+1234567890/messages
```

### Stop the System

```powershell
docker compose down
```

---

## 📚 Documentation

### Core Documentation

| File | Purpose |
|------|---------|
| **[CONSTITUTION.md](CONSTITUTION.md)** | Project principles, architecture decisions, and compliance rules. Read this first to understand the *why* behind design choices. |
| **[SPECS.md](SPECS.md)** | Technical specifications, technology stack versions, performance targets, and implementation requirements. |
| **[CONTRACTS.md](CONTRACTS.md)** | Complete API contracts, request/response schemas, validation rules, and example scenarios for all endpoints. |

### Configuration & Schema

| File | Purpose |
|------|---------|
| **[ENVIRONMENT.md](ENVIRONMENT.md)** | Comprehensive list of all environment variables, their defaults, descriptions, and security notes. |
| **[KAFKA_SCHEMA.md](KAFKA_SCHEMA.md)** | Kafka topic details, message schemas, serialization formats, and monitoring commands. |

### Testing & Development

| File | Purpose |
|------|---------|
| **[TEST_SCRIPTS.md](TEST_SCRIPTS.md)** | Complete test suite with 13 test categories, PowerShell scripts, and validation procedures. |
| **[TASKS.md](TASKS.md)** | Project task breakdown across 7 phases, tracking implementation progress (77/80 tasks completed). |

---

## 🔧 API Reference

### Java SMS Sender Service

**Send SMS**
```http
POST http://localhost:8080/v0/sms/send
Content-Type: application/json

{
  "phoneNumber": "+1234567890",
  "message": "Your message here"
}
```

**Response Statuses:**
- `SUCCESS`: SMS sent successfully
- `FAILED`: Vendor API error (random 30% failure)
- `BLOCKED`: Phone number in block list

**Health Check**
```http
GET http://localhost:8080/actuator/health
```

### Go SMS Store Service

**Get User Messages**
```http
GET http://localhost:8090/v0/user/{user_id}/messages
```

Returns array of SMS records sorted by timestamp (most recent first).

**Health Check**
```http
GET http://localhost:8090/health
```

---

## 🧪 Testing

### Quick Test Script

Run the included test script for rapid validation:

```powershell
# Send test messages
curl -X POST http://localhost:8080/v0/sms/send -H "Content-Type: application/json" -d '{\"phoneNumber\": \"+1234567890\", \"message\": \"Test message\"}'

# Wait for Kafka processing
Start-Sleep -Seconds 3

# Retrieve messages
curl http://localhost:8090/v0/user/+1234567890/messages

# Check MongoDB
docker exec -it polyglot-mongodb mongosh -u smsapp -p smsapp123 --authenticationDatabase sms_store --eval "db.sms_records.countDocuments()"
```

For comprehensive testing, see **[test.md](test.md)** with 13 detailed test scenarios.

---

## 🛠️ Development

### Project Structure

```
Polyglot/
├── JavaSender/          # Spring Boot SMS sender service
│   ├── src/main/java/com/sms/sender/
│   │   ├── controller/  # REST controllers
│   │   ├── service/     # Business logic
│   │   ├── kafka/       # Kafka producer
│   │   ├── model/       # Data models
│   │   └── config/      # Configuration beans
│   ├── Dockerfile       # Multi-stage build
│   └── pom.xml          # Maven dependencies
│
├── GoStore/             # Go SMS store service
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business services
│   ├── kafka/           # Kafka consumer
│   ├── models/          # Data models
│   ├── db/              # MongoDB client
│   ├── config/          # Configuration
│   ├── Dockerfile       # Multi-stage build
│   ├── go.mod           # Go dependencies
│   └── main.go          # Entry point
│
├── mongo-init/          # MongoDB initialization script
├── docker compose.yml   # Service orchestration
└── *.md                 # Documentation files
```

### Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Java Service | Spring Boot | 3.5.9 |
| Java Runtime | Eclipse Temurin | 21 |
| Go Service | Go | 1.25 |
| Message Broker | Apache Kafka | 7.6.0 |
| Cache | Redis | 7.2 |
| Database | MongoDB | 6.0 |

---

## 🔍 Monitoring

### Service Logs

```powershell
# All services
docker compose logs -f

# Specific service
docker compose logs -f sms-sender
docker compose logs -f sms-store
docker compose logs -f kafka
```

### Check Kafka Topics

```powershell
docker exec -it polyglot-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### Check Redis Block List

```powershell
docker exec -it polyglot-redis redis-cli SMEMBERS blocked_users
```

### Query MongoDB

```powershell
docker exec -it polyglot-mongodb mongosh -u smsapp -p smsapp123 --authenticationDatabase sms_store
```

---

## 🎯 Key Features

✅ **Polyglot Architecture**: Java (JVM) + Go (Native) services working together  
✅ **Asynchronous Communication**: Event-driven architecture via Kafka  
✅ **Polyglot Persistence**: Redis for cache + MongoDB for document storage  
✅ **Docker Native**: Full containerization with health checks  
✅ **Production Patterns**: Circuit breaker, retry logic, idempotency  
✅ **Comprehensive Testing**: 13 test categories with automation scripts  
✅ **Complete Documentation**: Architecture, APIs, configuration, and testing  

---

## 🐛 Troubleshooting

### Services Won't Start

```powershell
# Check service status
docker compose ps

# View error logs
docker compose logs <service-name>

# Restart specific service
docker compose restart <service-name>
```

### Port Conflicts

Check if ports are already in use:
```powershell
netstat -ano | findstr "8080 8090 9092 27017 6379"
```

### Clean Restart

```powershell
# Remove all containers and volumes
docker compose down -v

# Start fresh
docker compose up -d
```

### Common Issues

| Issue | Solution |
|-------|----------|
| "Connection refused" | Wait for health checks to pass (~30s) |
| "Kafka consumer lag" | Check `docker compose logs kafka` |
| "MongoDB auth failed" | Verify credentials in `.env` or docker compose.yml |
| "Redis keys not found" | Check Java service initialized block list |

For detailed troubleshooting, see **[test.md](test.md)** Section 17.

---

## 🔒 Security Notes

⚠️ **WARNING**: This is a development/demonstration setup.

For production deployments:
- Change all default passwords (Redis, MongoDB)
- Enable TLS/SSL for all inter-service communication
- Use secrets management (Docker Secrets, Vault)
- Implement authentication for REST APIs
- Enable Kafka SASL authentication
- Set up proper network segmentation

See **[ENVIRONMENT.md](ENVIRONMENT.md)** for security best practices.

---

## 📝 License

This is a demonstration project for educational purposes.

---

## 👥 Contributing

1. Follow the architecture principles in **[CONSTITUTION.md](CONSTITUTION.md)**
2. Adhere to API contracts in **[CONTRACTS.md](CONTRACTS.md)**
3. Update documentation when making changes
4. Run full test suite from **[TEST_SCRIPTS.md](TEST_SCRIPTS.md)**
5. Ensure all health checks pass

**Last Updated**: December 27, 2025