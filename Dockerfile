# Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/public .
COPY --from=builder /app/email ./email
# Expose port 8080 to the outside world
EXPOSE 5001

RUN apk add dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# Command to run the executable
CMD ["./main"]
