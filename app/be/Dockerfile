# Build stage
FROM golang:1.21-alpine as builder

# Set working directory
WORKDIR /app

# Install required dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mafia-wars-backend ./cmd/server

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/mafia-wars-backend .

# Copy configuration files
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/migrations /app/migrations

# Set permissions
RUN chmod +x /app/mafia-wars-backend

# Expose the application port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/mafia-wars-backend"]
CMD ["--config=./configs/app.yaml"]