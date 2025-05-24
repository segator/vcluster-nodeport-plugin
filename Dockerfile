FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY main.go main.go
COPY hook hook

# Build the Go application
# -ldflags="-s -w" strips debug information and symbols, reducing binary size
# CGO_ENABLED=0 ensures a static binary without C dependencies
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /plugin ./cmd/plugin/main.go

# Stage 2: Create the final lightweight image
FROM alpine:3.18

# Copy the static binary from the builder stage
COPY --from=builder /plugin /plugin

# Set the entrypoint for the container
ENTRYPOINT ["/plugin"]