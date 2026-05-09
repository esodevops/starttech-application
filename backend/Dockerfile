# -------------------------------------------------
# Backend Dockerfile
# Builds a minimal production image for the
# Golang API using a two-stage build:
#   Stage 1: compile the Go binary
#   Stage 2: copy just the binary into a tiny image
# -------------------------------------------------

# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first (Docker caches this layer if go.mod didn't change)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build the binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# ---- Stage 2: Run ----
# Use Alpine runtime with CA certs so TLS connections (MongoDB Atlas) work reliably.
FROM alpine:3.20

WORKDIR /app

# Atlas requires TLS; install trust store used by Go's crypto/x509.
RUN apk add --no-cache ca-certificates tzdata

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/server .

# The app listens on port 8080
EXPOSE 8080

# Run as non-root user for security
RUN addgroup -S app && adduser -S -G app app
USER app:app

ENTRYPOINT ["/app/server"]
