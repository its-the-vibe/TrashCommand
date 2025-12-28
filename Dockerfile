# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Copy source files
COPY . .

# Build the binary with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o trashcommand .

# Runtime stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /app/trashcommand /trashcommand

# Copy CA certificates for HTTPS requests to Slack API
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Run the binary
ENTRYPOINT ["/trashcommand"]
