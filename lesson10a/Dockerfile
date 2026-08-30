# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o main ./cmd/api

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# We do not want the application to run as root, so we create a user and group for it
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Run the application as the non-root user
USER app

# Expose port
EXPOSE 3000

# Run the binary
CMD ["./main"]


