# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# Final lightweight production image
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates untuk koneksi SSL/TLS ke Neon PostgreSQL
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server /app/server
COPY --from=builder /app/web /app/web

# Expose port (PORT default 8080, otomatis disesuaikan dengan env $PORT platform)
EXPOSE 8080

ENV APP_ENV=production

CMD ["/app/server"]
