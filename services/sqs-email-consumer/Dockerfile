# Use an official Go runtime as a parent image
FROM golang:latest AS builder

# Add Maintainer Info
LABEL maintainer="Swaraj kumar singh Singh <sswaraj169@gmail.com>"

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the local package files to the container's workspace
COPY . .

# Build the Go application inside the container
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# FOR PROD: Command to run the executable
CMD ["./main"]