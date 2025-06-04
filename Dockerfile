# Build stage
FROM golang:1.24-alpine AS builder

# Override user name at build. If build-arg is not passed, will create user named `default_user`
ARG DOCKER_USER=appuser

# Create a group and user
RUN addgroup -S $DOCKER_USER && adduser -S $DOCKER_USER -G $DOCKER_USER

# Tell docker that all future commands should run as this user
USER $DOCKER_USER

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
COPY --from=builder /app/main .

# Uncomment if your application needs these
# COPY --from=builder /app/templates ./templates
# COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./main"]