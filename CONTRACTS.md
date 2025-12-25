# API Contracts & Data Structures

This document defines all communication contracts, request/response structures, and data models used across the Polyglot SMS Service.

---

## Table of Contents
1. [Java SMS Sender Service](#java-sms-sender-service)
2. [Go SMS Store Service](#go-sms-store-service)
3. [Kafka Events](#kafka-events)
4. [MongoDB Schema](#mongodb-schema)
5. [Redis Data Structures](#redis-data-structures)

---

## Java SMS Sender Service

### REST API Endpoint

#### Send SMS
**Endpoint:** `POST /v0/sms/send`

**Request Body:**
```json
{
  "phoneNumber": "string",
  "message": "string"
}
```

**Request Schema (SmsRequest):**
| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| phoneNumber | string | Yes | Regex: `^\\+?[1-9]\\d{9,14}$` | Phone number in international format (10-15 digits) |
| message | string | Yes | Length: 1-160 chars | SMS message content |

**Response Body:**
```json
{
  "status": "string",
  "message": "string",
  "timestamp": "2025-12-25T10:30:00",
  "phoneNumber": "string"
}
```

**Response Schema (SmsResponse):**
| Field | Type | Description |
|-------|------|-------------|
| status | string | Operation status: `SUCCESS`, `FAILED`, or `BLOCKED` |
| message | string | Descriptive message about the operation result |
| timestamp | LocalDateTime (ISO-8601) | When the operation was processed |
| phoneNumber | string | The phone number that was processed |

**Status Codes:**
- `200 OK` - SMS sent successfully
- `400 Bad Request` - Invalid input (validation errors)
- `403 Forbidden` - Phone number is blocked
- `500 Internal Server Error` - Vendor failure or system error

**Example Success Response:**
```json
{
  "status": "SUCCESS",
  "message": "SMS sent successfully",
  "timestamp": "2025-12-25T10:30:00",
  "phoneNumber": "+1234567890"
}
```

**Example Blocked Response:**
```json
{
  "status": "BLOCKED",
  "message": "Phone number is in the block list",
  "timestamp": "2025-12-25T10:30:00",
  "phoneNumber": "+1234567890"
}
```

**Example Failed Response:**
```json
{
  "status": "FAILED",
  "message": "Vendor API returned error",
  "timestamp": "2025-12-25T10:30:00",
  "phoneNumber": "+1234567890"
}
```

---

## Go SMS Store Service

### REST API Endpoint

#### Get User Messages
**Endpoint:** `GET /v0/user/{user_id}/messages`

**Path Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| user_id | string | User identifier (phoneNumber) |

**Response Body:**
```json
[
  {
    "id": "string",
    "user_id": "string",
    "phone_number": "string",
    "message": "string",
    "status": "string",
    "created_at": "2025-12-25T10:30:00Z"
  }
]
```

**Response Schema (Array of SMSRecord):**
| Field | Type | Description |
|-------|------|-------------|
| id | string | MongoDB document ID |
| user_id | string | User identifier (same as phoneNumber) |
| phone_number | string | Phone number that received the SMS |
| message | string | SMS message content |
| status | string | SMS status: `SUCCESS` or `FAILED` |
| created_at | time.Time (RFC3339) | When the record was created |

**Status Codes:**
- `200 OK` - Messages retrieved successfully (may be empty array)
- `400 Bad Request` - Invalid user_id format
- `500 Internal Server Error` - Database error

**Example Response:**
```json
[
  {
    "id": "674c5f8a1234567890abcdef",
    "user_id": "+1234567890",
    "phone_number": "+1234567890",
    "message": "Hello, this is a test message",
    "status": "SUCCESS",
    "created_at": "2025-12-25T10:30:00Z"
  },
  {
    "id": "674c5f8b1234567890abcdeg",
    "user_id": "+1234567890",
    "phone_number": "+1234567890",
    "message": "Another test message",
    "status": "SUCCESS",
    "created_at": "2025-12-25T10:31:00Z"
  }
]
```

---

## Kafka Events

### Topic: `sms.events`

**Purpose:** Asynchronous communication from Java SMS Sender to Go SMS Store for logging and persistence.

**Event Schema (KafkaEvent):**
```json
{
  "eventId": "string (UUID)",
  "userId": "string",
  "phoneNumber": "string",
  "message": "string",
  "status": "string",
  "createdAt": "2025-12-25T10:30:00"
}
```

**Field Descriptions:**
| Field | Type | Description |
|-------|------|-------------|
| eventId | string (UUID) | Unique identifier for the event |
| userId | string | User identifier (same as phoneNumber) |
| phoneNumber | string | Phone number that received the SMS |
| message | string | SMS message content |
| status | string | SMS operation status: `SUCCESS` or `FAILED` |
| createdAt | LocalDateTime (ISO-8601) | When the event was created |

**Example Kafka Event:**
```json
{
  "eventId": "550e8400-e29b-41d4-a716-446655440000",
  "userId": "+1234567890",
  "phoneNumber": "+1234567890",
  "message": "Hello, this is a test message",
  "status": "SUCCESS",
  "createdAt": "2025-12-25T10:30:00"
}
```

**Producer:** Java SMS Sender Service  
**Consumer:** Go SMS Store Service (Consumer Group: `sms-store-consumer-group`)

**Serialization:**
- **Producer (Java):** JSON (Spring Kafka JsonSerializer)
- **Consumer (Go):** JSON (segmentio/kafka-go or confluent-kafka-go)

**Delivery Guarantees:**
- Producer: `acks=all`, `retries=3`
- Consumer: Auto-commit enabled (configurable)

---

## MongoDB Schema

### Database: `sms_store`
### Collection: `sms_records`

**Document Structure:**
```json
{
  "_id": ObjectId("674c5f8a1234567890abcdef"),
  "user_id": "+1234567890",
  "phone_number": "+1234567890",
  "message": "Hello, this is a test message",
  "status": "SUCCESS",
  "created_at": ISODate("2025-12-25T10:30:00.000Z")
}
```

**Schema Definition:**
| Field | BSON Type | Index | Description |
|-------|-----------|-------|-------------|
| _id | ObjectId | Primary Key | MongoDB document ID |
| user_id | string | Yes (Single) | User identifier (phoneNumber) |
| phone_number | string | No | Phone number (redundant with user_id) |
| message | string | No | SMS message content |
| status | string | No | `SUCCESS` or `FAILED` |
| created_at | Date | Yes (Descending) | Record creation timestamp |

**Indexes:**
1. **Single Index:** `{ user_id: 1 }` - For efficient user lookup
2. **Single Index:** `{ created_at: -1 }` - For time-based queries
3. **Compound Index:** `{ user_id: 1, created_at: -1 }` - For paginated user history

**Access:**
- **Username:** `smsapp` (or as configured in `MONGO_APP_USER`)
- **Password:** `smsapp123` (or as configured in `MONGO_APP_PASSWORD`)
- **Permissions:** readWrite on `sms_store` database

---

## Redis Data Structures

### Key: `blocked_users`

**Type:** SET

**Purpose:** Maintains a list of blocked phone numbers that should not receive SMS messages.

**Operations:**
- **Check Membership:** `SISMEMBER blocked_users "+1234567890"`
- **Add to Blocklist:** `SADD blocked_users "+1234567890"`
- **Remove from Blocklist:** `SREM blocked_users "+1234567890"`
- **Get All Blocked:** `SMEMBERS blocked_users`

**Value Format:** Phone numbers in international format (e.g., `+1234567890`)

**Example Members:**
```
"+1234567890"
"+9876543210"
"+1122334455"
```

**Initialization:**
Populated on Java service startup with dummy blocked users for testing:
- `+1111111111`
- `+2222222222`
- `+3333333333`

**Access:**
- **Host:** redis
- **Port:** 6379
- **No authentication required** (default Redis configuration)

---

## Data Flow Summary

```
1. Client sends POST /v0/sms/send with SmsRequest
   ↓
2. Java service checks Redis blocked_users SET
   ↓
3. If blocked: Return SmsResponse with status=BLOCKED
   If not blocked: Continue
   ↓
4. Java service mocks 3rd party vendor call (random SUCCESS/FAIL)
   ↓
5. Java service publishes KafkaEvent to sms.events topic
   ↓
6. Java service returns SmsResponse to client
   ↓
7. Go service consumes KafkaEvent from Kafka
   ↓
8. Go service persists SMSRecord to MongoDB sms_records collection
   ↓
9. Client can retrieve history via GET /v0/user/{user_id}/messages
   ↓
10. Go service returns array of SMSRecord from MongoDB
```

---

## Notes

### Timestamp Formats
- **Java (LocalDateTime):** ISO-8601 format without timezone: `2025-12-25T10:30:00`
- **Go (time.Time):** RFC3339 format with timezone: `2025-12-25T10:30:00Z`
- **MongoDB (ISODate):** BSON Date type, stored as UTC

### Status Values
- **SUCCESS** - SMS sent successfully
- **FAILED** - SMS failed to send (vendor error)
- **BLOCKED** - Phone number in block list (only in SmsResponse)

### Phone Number Format
- Recommended: International format with `+` prefix (e.g., `+1234567890`)
- Validation: 10-15 digits, may optionally start with `+`
- Pattern: `^\\+?[1-9]\\d{9,14}$`

### Error Handling
- All services use structured error responses
- Proper HTTP status codes
- Descriptive error messages
- Logging at appropriate levels (INFO, WARN, ERROR)
