FROM golang:1.19-alpine

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies with retries
RUN go mod download || go mod download || go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o api ./cmd/api

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./api"]