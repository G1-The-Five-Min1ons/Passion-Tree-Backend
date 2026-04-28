# --- Stage 1: Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/internal/auth/handler/templates ./internal/auth/handler/templates

ENV PORT=5000
EXPOSE 5000

CMD ["./main"]