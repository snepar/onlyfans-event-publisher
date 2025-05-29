FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with proper flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /onlyfans-event-publisher ./cmd/publisher

# Create a minimal runtime container
FROM alpine:3.18

# Add CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /onlyfans-event-publisher .

# Make the binary executable
RUN chmod +x /app/onlyfans-event-publisher

# Run the binary
ENTRYPOINT ["/app/onlyfans-event-publisher"]