# ==========================================
# Stage 1: Build the Go binary
# ==========================================
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the rest of our Go source code
COPY . .

# Build the Go app as a statically linked binary
# CGO_ENABLED=0 ensures it runs on any Linux environment without requiring C libraries
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway main.go

# ==========================================
# Stage 2: Create the minimal production image
# ==========================================
FROM alpine:latest

WORKDIR /root/

# Install CA certificates so our proxy can securely connect to HTTPS APIs (like OpenAI)
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /app/gateway .

# Expose the port our Go app listens on
EXPOSE 8080

# Command to run the executable
CMD ["./gateway"]