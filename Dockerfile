# Stage 1: Build the Go binary
FROM docker.io/golang:1.21-alpine AS builder

# Install git for go module dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o birthdays-app ./cmd/app


# Stage 2: Minimal runtime image
FROM docker.io/alpine:edge

# Install ca-certificates (if your app makes HTTPS requests)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/birthdays-app .
COPY templates/ /app/templates

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./birthdays-app"]
