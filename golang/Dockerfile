# -----Stage: Build-----
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o warden .

# -----Stage: Run-----
FROM alpine:3.23.4 

# Set the working directory inside the container
WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Add a non-root user to run the application
RUN adduser -D -u 1000 appuser

# Copy the built binary from the builder stage
COPY --from=builder --chown=appuser:appuser /app/warden .

# Switch to the non-root user
USER appuser

# Command to run the application
CMD ["./warden"]