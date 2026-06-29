# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/fintrack-api ./cmd/api

# Production stage
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates for HTTPS calls and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/fintrack-api .

# Copy migrations directory (needed for auto-migrate)
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

ENTRYPOINT ["./fintrack-api"]
