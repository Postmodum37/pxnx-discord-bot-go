# Build stage
FROM golang:1.25-alpine AS builder

# Install git (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pxnx-discord-bot .

# Runtime stage
FROM alpine:latest

# Install system dependencies including Python for yt-dlp service
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    python3 \
    py3-pip \
    ffmpeg \
    && python3 -m ensurepip \
    && pip3 install --no-cache --upgrade pip setuptools

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/pxnx-discord-bot .

# Copy yt-dlp service files
COPY --from=builder /app/services/ ./services/

# Install Python dependencies for yt-dlp service
RUN pip3 install --no-cache -r services/ytdlp/requirements.txt

# Create cache directories
RUN mkdir -p /tmp/ytdlp-cache/logs

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app && \
    chown -R appuser:appgroup /tmp/ytdlp-cache

# Switch to non-root user
USER appuser

# Health check for both bot and yt-dlp service
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD pgrep pxnx-discord-bot || exit 1

# Expose ports for yt-dlp service
EXPOSE 8080

# Command to run the application
CMD ["./pxnx-discord-bot"]