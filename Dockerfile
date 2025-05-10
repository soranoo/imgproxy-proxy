# Build stage
FROM golang:1.24-alpine AS build

WORKDIR /app

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o imgproxy-proxy -ldflags="-s -w" ./cmd/server

# Runtime stage
FROM alpine:3.21

# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from build stage
COPY --from=build /app/imgproxy-proxy .

# Set non-root user for security
RUN adduser -D -H -h /app appuser
USER appuser

# Environment variables (these can be overridden at runtime)
ENV DOCKER_ENV=true \
    IMGPROXY_BASE_URL=http://imgproxy:8080 \
    IMGPROXY_ENCODE=true \
    METRICS_ENABLED=true \
    METRICS_ENDPOINT=/metrics \
    METRICS_NAMESPACE=imgproxy_proxy \
    LOG_LEVEL=1

# Expose the default port
EXPOSE 8080

# Command to run
ENTRYPOINT ["/app/imgproxy-proxy"]
