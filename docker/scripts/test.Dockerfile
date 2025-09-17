# Test runner Dockerfile for DDAO
FROM golang:1.21-alpine

# Install necessary tools
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    sqlite-dev

# Set working directory
WORKDIR /workspace

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the project to verify everything compiles
RUN go build ./...

# Default command (can be overridden)
CMD ["go", "test", "./...", "-v"]