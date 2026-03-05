# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a \
    -ldflags "-w -s -X main.Version=1.0.0" \
    -o /app/cms-backend ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/cms-backend .
COPY .env.example .env

EXPOSE 8080

ENTRYPOINT ["./cms-backend"]
