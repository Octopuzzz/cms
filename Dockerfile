# Build Stage
FROM golang:1.23-alpine AS builder

# Set the working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev libc-dev tzdata

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-w -s -X main.Version=1.0.0" -o /app/bin/cms-backend ./cmd/server

# Run Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache tzdata ca-certificates libc6-compat

# Copy the built binary from the builder stage
COPY --from=builder /app/bin/cms-backend /app/cms-backend

# Set the entrypoint
ENTRYPOINT ["/app/cms-backend"]

# Expose port
EXPOSE 8080
