# File: Dockerfile

# Stage 1: Build + Test the Go binary
FROM docker.io/golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test ./... -v

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o birthdays-app ./cmd/app

# Stage 2: Minimal runtime image with CA certificates
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache --no-scripts add ca-certificates

WORKDIR /app

COPY --from=builder /app/birthdays-app .

EXPOSE 8080

CMD ["./birthdays-app"]
