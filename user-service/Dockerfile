# Step 1: Build the Go binary
FROM golang:1.23.5 AS builder

WORKDIR /app

# Copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./

RUN go mod download

# Copy the source code
COPY . .

# Build the binary (replace <service-name> with the service name)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o user-service ./cmd/main.go

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /app/user-service /user-service

# Ensure the binary is executable
RUN chmod +x /user-service

# Run the binary
CMD ["/user-service"]