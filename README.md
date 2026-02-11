# CertWatch

A minimal, security-focused Certificate Inventory & Expiry Monitoring Service written in Go.

CertWatch provides a small HTTP API for registering and tracking X.509 certificate metadata. It is intentionally designed to demonstrate secure coding practices, defensive architecture, and audit-ready logging in a Golang + SQLite environment.

This project is not intended to store private keys or act as a Certificate Authority. It focuses strictly on certificate metadata lifecycle management.

## Goals

This project emphasizes:
- Secure HTTP server configuration
- Strict input validation
- Protection against SQL injection
- Protection against log injection
- Minimal error surface exposure
- Structured, audit-friendly logging
- Context-based timeouts for resource safety
- Database-level integrity enforcement
- Clean layered architecture

The goal is to demonstrate security-aware backend engineering practices suitable for certificate lifecycle management systems.

## Architecture

The application follows a layered design:
```
Client
  ↓
HTTP Layer (handlers, middleware)
  ↓
Service Layer (business logic)
  ↓
Repository Layer (parameterized SQL only)
  ↓
SQLite Database
```

### Layer Responsibilities


#### HTTP Layer

- Enforces content-type
- Limits request body size
- Rejects unknown JSON fields
- Injects request ID
- Maps internal errors to sanitized responses

#### Service Layer

- Business validation
- Expiry logic
- Domain rules enforcement

#### Repository Layer

- Uses database/sql
- Parameterized queries only
- No dynamic SQL concatenation
- Context timeouts enforced

### Database Layer

- CHECK constraints
- UNIQUE constraints
- Indexed expiry date
- Foreign key enforcement enabled

#### Features

- Register certificate metadata
- List certificates
- Retrieve certificate by ID
- Delete certificate
- Background expiry monitoring
- Structured JSON logging
- Audit event classification

## API Overview


### POST /certificates

Registers certificate metadata.

Example request:
```json
{
  "common_name": "example.com",
  "serial_number": "123456789",
  "issuer": "Example CA",
  "not_before": "2025-01-01T00:00:00Z",
  "not_after": "2026-01-01T00:00:00Z",
  "fingerprint_sha256": "64_CHAR_HEX_STRING"
}
```

Validation rules:

- All fields required
- RFC3339 timestamps
- not_after must be later than not_before
- Fingerprint must be 64-character SHA-256 hex
- Unknown JSON fields rejected

### GET /certificates

Returns all registered certificates.

### GET /certificates/{id}

Returns a single certificate record.

### DELETE /certificates/{id}

Removes a certificate entry.

## Database Schema

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE certificates (
    id TEXT PRIMARY KEY,
    common_name TEXT NOT NULL CHECK(length(common_name) <= 255),
    serial_number TEXT NOT NULL CHECK(length(serial_number) <= 128),
    issuer TEXT NOT NULL CHECK(length(issuer) <= 255),
    not_before DATETIME NOT NULL,
    not_after DATETIME NOT NULL,
    fingerprint_sha256 TEXT NOT NULL UNIQUE CHECK(length(fingerprint_sha256) = 64),
    created_at DATETIME NOT NULL
);

CREATE INDEX idx_cert_not_after ON certificates(not_after);
```

### Security Properties

- Database enforces field length constraints
- Fingerprint uniqueness prevents duplicates
- Index supports efficient expiry checks
- Constraints act as defense-in-depth against validation bugs

## Security Design Considerations

### SQL Injection Prevention

- All queries use parameterized statements
- No string concatenation for SQL
- Repository layer encapsulates DB access

### Log Injection Prevention

- Structured logging via log/slog
- No raw request bodies logged
- User-controlled values validated before logging

### Error Handling

- Internal errors are wrapped and logged
- Clients receive generic error codes
- No stack traces or SQL errors exposed

Example client error response:
```json
{
  "error": "invalid_request"
}
```

### Resource Protection

- HTTP timeouts configured
- Request body size limited
- Context timeouts for DB operations
- Controlled background goroutine lifecycle

## Logging & Audit

All logs are structured JSON and include:

- request_id
- event_type
- remote_ip
- HTTP method
- path
- status_code
- latency
- certificate_id (when applicable)

Example audit log:
```json
{
  "level": "INFO",
  "event": "certificate_created",
  "certificate_id": "uuid",
  "request_id": "abc123",
  "timestamp": "..."
}
```

Audit logging enables traceability of certificate lifecycle events.

## Expiry Monitoring

A background worker runs periodically to:

- Query certificates expiring within 30 days
- Emit structured audit events

This simulates proactive certificate lifecycle management monitoring.

## Running the Application

### Requirements

- Go 1.22+
- SQLite

### Setup
```bash
git clone <repo>
cd certwatch
go mod tidy
go run ./cmd/server
```

The server listens on :8080 by default.

## Secure HTTP Configuration

- The server enforces:
- ReadTimeout
- ReadHeaderTimeout
- WriteTimeout
- IdleTimeout
- MaxHeaderBytes
- Content-Type validation
- Body size limits

These controls reduce exposure to slowloris-style and resource exhaustion attacks.