# Build stage
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 is needed for static linking, ensuring it runs on Alpine
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux go build -o etcdfinder main.go

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/etcdfinder .

# Copy config file
# Assuming the config is located at internal/config/config.yaml
COPY --from=builder /app/internal/config/config.yaml .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./etcdfinder"]
