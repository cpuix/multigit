# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /multigit ./cmd/multigit

# Final stage
FROM alpine:3.18

# Install required packages
RUN apk --no-cache add \
    ca-certificates \
    git \
    openssh-client

# Copy the binary from builder
COPY --from=builder /multigit /usr/local/bin/multigit

# Set the working directory
WORKDIR /workspace

# Set the entrypoint
ENTRYPOINT ["multigit"]
