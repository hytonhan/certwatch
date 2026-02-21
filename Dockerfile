# ---------- Builder Stage ----------
    FROM golang:1.26.0-alpine AS builder

    WORKDIR /app
    
    # Install build dependencies
    RUN apk add --no-cache build-base
    
    # Copy go mod files first (better layer caching)
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy source
    COPY . .
    
    # Build static binary
    RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
        go build -ldflags="-s -w" -o certwatch ./cmd/server
    
    # ---------- Runtime Stage ----------
    FROM alpine:3.23
    
    # Create non-root user
    RUN addgroup -S appgroup && adduser -S appuser -G appgroup
    
    WORKDIR /app
    
    # Copy binary from builder
    COPY --from=builder /app/certwatch .
    
    # Create directory for SQLite database
    RUN mkdir -p /data && chown -R appuser:appgroup /data
    
    USER appuser
    
    EXPOSE 8080
    
    ENTRYPOINT ["./certwatch"]