# Multi-stage build for minimal image
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the CLI tool
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w -buildid=' \
    -o /jitter-token ./cmd/jitter-token

# Final stage - minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /jitter-token ./

# Create non-root user
RUN addgroup -g 1001 -S jitter && \
    adduser -S jitter -u 1001

USER jitter

ENTRYPOINT ["./jitter-token"]
CMD ["--count", "3"]
