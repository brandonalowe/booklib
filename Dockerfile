# Build stage
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite-libs sqlite

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/server .

# Copy scripts directory (contains backup.sh)
COPY --from=builder /app/scripts ./scripts

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./server"]
