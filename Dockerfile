# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Utilize build cache by copying mod files first
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Final stage
FROM scratch

# Copy SSL certificates (needed for HTTPS requests)
# COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
WORKDIR /app
COPY --from=builder /app/.env .
COPY --from=builder /app/main .

# Uncomment if your application needs these
# COPY --from=builder /app/templates ./templates
# COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./main"]