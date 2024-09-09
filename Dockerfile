# Stage 1: Build the Go application
FROM golang:1.20 AS builder

# Set environment variables for Go
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker's caching mechanism
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the statically linked binaries using the makefile
RUN make build_node build_bootstrap_node

# Stage 2: Create a minimal container with the statically linked binaries
FROM alpine:latest

# Install necessary certificates for secure connections (if needed)
RUN apk --no-cache add ca-certificates

# Create a non-root user for security
RUN adduser -D appuser

# Set the working directory
WORKDIR /app

# Copy the built statically linked binaries from the builder stage
COPY --from=builder /app/dht_node /app/dht_node
COPY --from=builder /app/bootstrap_node /app/bootstrap_node

# Copy the configuration file
COPY --from=builder /app/config.ini /app/config.ini

# Ensure the binaries are executable
RUN chmod +x /app/dht_node /app/bootstrap_node

# Change ownership of the files to the non-root user
RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

# Expose port 7400 for API communication
EXPOSE 7400/tcp
# Expose port 8000 for P2P communication
EXPOSE 8000/tcp

# Define the entrypoint command to run the correct binary based on the NODE_TYPE
CMD ["sh", "-c", "if [ \"$NODE_TYPE\" = \"node\" ]; then /app/dht_node -c /app/config.ini; elif [ \"$NODE_TYPE\" = \"bootstrap_node\" ]; then /app/bootstrap_node -c /app/config.ini; else echo \"Unknown NODE_TYPE: $NODE_TYPE\" && exit 1; fi"]