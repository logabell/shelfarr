# ============================================================
# Shelfarr Production Dockerfile
# Multi-stage build with Go backend + React frontend
# Bundled with Calibre ebook-convert and FFmpeg
# ============================================================

# ------------------------------------------------------------
# Stage 1: Build Go Backend
# ------------------------------------------------------------
FROM golang:1.22-alpine AS backend-builder

WORKDIR /build

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

# Copy go mod files first for caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ .

# Build static binary with CGO for SQLite
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -ldflags '-linkmode external -extldflags "-static" -s -w' \
    -o shelfarr \
    ./cmd/shelfarr

# ------------------------------------------------------------
# Stage 2: Build React Frontend
# ------------------------------------------------------------
FROM node:20-alpine AS frontend-builder

WORKDIR /build

# Copy package files for caching
COPY frontend/package*.json ./
RUN npm ci --silent

# Copy source and build
COPY frontend/ .
RUN npm run build

# ------------------------------------------------------------
# Stage 3: Final Production Image
# ------------------------------------------------------------
FROM alpine:3.19

LABEL maintainer="Shelfarr" \
      description="Self-hosted book management system" \
      version="1.0.0"

# Install runtime dependencies
# - ca-certificates: HTTPS support
# - ffmpeg: Audio processing for audiobooks (MP3 to M4B)
# - calibre: Ebook format conversion
# - tzdata: Timezone support
# - wget: Health checks
RUN apk add --no-cache \
    ca-certificates \
    ffmpeg \
    calibre \
    tzdata \
    wget \
    && rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1000 -S shelfarr && \
    adduser -u 1000 -S shelfarr -G shelfarr

# Create required directories with proper permissions
RUN mkdir -p /config /books /audiobooks /downloads /app/public \
    && chown -R shelfarr:shelfarr /config /books /audiobooks /downloads /app

# Copy binary from backend builder
COPY --from=backend-builder --chown=shelfarr:shelfarr /build/shelfarr /usr/local/bin/shelfarr

# Copy frontend assets from frontend builder
COPY --from=frontend-builder --chown=shelfarr:shelfarr /build/dist /app/public

WORKDIR /app

# Environment variables with sensible defaults
ENV SHELFARR_CONFIG_PATH=/config \
    SHELFARR_DB_PATH=/config/shelfarr.db \
    SHELFARR_BOOKS_PATH=/books \
    SHELFARR_AUDIOBOOKS_PATH=/audiobooks \
    SHELFARR_DOWNLOADS_PATH=/downloads \
    SHELFARR_RECYCLE_PATH=/config/recycle \
    SHELFARR_LISTEN_ADDR=:8080 \
    HARDCOVER_API_URL=https://api.hardcover.app/v1/graphql \
    TZ=UTC

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

# Define volumes for persistent data
VOLUME ["/config", "/books", "/audiobooks", "/downloads"]

# Switch to non-root user
USER shelfarr

# Run the application
ENTRYPOINT ["shelfarr"]
