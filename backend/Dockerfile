# Start with a lightweight Go base image
# Using Debian-based image to avoid QEMU/Alpine compatibility issues (execve error)
# and to support Go 1.24 required by dependencies.
FROM golang:1.24-bookworm AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache for dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
# -o main: name the output binary "main"
RUN go build -o main .

# Start a new stage for a smaller final image
# Using debian-slim is more robust for multi-arch builds on QEMU than Alpine
FROM debian:bookworm-slim

# Install ca-certificates in case the app invokes HTTPS requests (likely for Tuya API)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy .env file if necessary (comment out if you prefer passing env vars via CLI)
# COPY .env . 
# Note: It's often better to pass environment variables at runtime or mount the .env file

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
