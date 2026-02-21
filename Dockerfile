# Build stage
FROM golang:1.24-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git 

WORKDIR /app

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Buat user non-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup -u 1000

# Working directory yang accessible oleh appuser
WORKDIR /app

# Copy binary dari builder
COPY --from=builder /app/main .

# Beri ownership ke appuser
RUN chown -R appuser:appgroup /app

# Switch ke non-root user
USER appuser

EXPOSE 3000
CMD ["./main"]