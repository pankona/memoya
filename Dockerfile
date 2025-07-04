# Dockerfile for Cloud Run deployment  
FROM golang:1.23 AS builder

# Install dependencies
RUN apt-get update && apt-get install -y ca-certificates git make wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate OpenAPI code
RUN make generate

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o memoya-server ./cmd/memoya-server

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates tzdata wget && rm -rf /var/lib/apt/lists/*
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/memoya-server .

# Set timezone
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./memoya-server"]