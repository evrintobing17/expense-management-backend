FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build API and migrate binaries
RUN go build -o /usr/local/bin/api ./cmd/api
RUN go build -o /usr/local/bin/migration ./cmd/migration

# Final lightweight image
FROM alpine:latest

# Install CA certificates and optionally dumb-init for better signal handling
RUN apk add --no-cache ca-certificates dumb-init

# Copy binaries from builder
COPY --from=builder /usr/local/bin/api /usr/local/bin/
COPY --from=builder /usr/local/bin/migration /usr/local/bin/

# Create a non-root user to run the app
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Switch to non-root user
USER appuser

EXPOSE 8080

# Use dumb-init for proper signal handling
ENTRYPOINT ["dumb-init", "--"]

# Run migrations then start API
CMD ["sh", "-c", "migrate && api"]