# Use an official Go runtime as a parent image for building the application
FROM golang:latest AS builder

# Add Maintainer Info
LABEL maintainer="Swaraj Kumar Singh <sswaraj169@gmail.com>"

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files for dependency resolution
COPY go.mod go.sum ./

# Download dependencies early to take advantage of caching
RUN go mod download

# Copy the entire application code
COPY . .

# Build the Go application with optimizations enabled
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a lightweight image for the final stage
FROM alpine:latest

# Add Maintainer Info
LABEL maintainer="Swaraj Kumar Singh <sswaraj169@gmail.com>"

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
