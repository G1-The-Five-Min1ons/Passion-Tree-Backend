# Production Dockerfile for Go backend (no hot-reload)
# 1) Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git build-base
COPY go.mod .
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go build -ldflags "-s -w" -o /out/app ./

# 2) Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl && update-ca-certificates
WORKDIR /app
COPY --from=builder /out/app ./app
# Create non-root user
RUN adduser -D -H -u 10001 appuser
USER 10001
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/app"]
