# Build stage
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/server/main.go

# Final stage
FROM alpine:latest

# Install necessary runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Create necessary directories
WORKDIR /app
RUN mkdir -p /app/uploads

# Copy built binary and config
COPY --from=builder /app/app /app/
COPY --from=builder /app/config/config.yaml /app/config/
COPY --from=builder /app/docs /app/docs

# Set proper permissions
RUN chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8080

# Set environment variables
ENV CONFIG_PATH="/app/config/config.yaml"

# Command to run the application
CMD ["/app/app"]