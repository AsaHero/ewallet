# Stage 1: Builder
FROM golang:latest AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the local package files to the container's workspace.
COPY . ./

# Installing dependencies and building
RUN make build-linux

# Stage 2: Final Image
FROM alpine:latest

# Copy the executable from the builder stage
COPY --from=builder /app/bin/ewallet ./ewallet


# Grant execution permissions to the executable
RUN chmod +x ./ewallet

# Expose port 8081
EXPOSE 8081

# Command to run the executable
CMD ["./ewallet"]
