# Dockerfile
FROM golang:1.19-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test-ordent .

# Create smaller final image
FROM alpine:latest  

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/test-ordent .
COPY --from=builder /app/config.yaml .

# Create uploads directory
RUN mkdir -p ./uploads

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./test-ordent"]