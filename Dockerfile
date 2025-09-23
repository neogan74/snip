# syntax=docker/dockerfile:1

FROM golang:1.23-alpine3.20 AS builder
WORKDIR /app

# Reuse Go module cache between builds and produce a statically linked binary
ENV CGO_ENABLED=0 GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /app/bin/web ./cmd/web

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/web

# Stage 2: создание минимального образа
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .

COPY --from=builder /app/bin/web ./web

USER app
EXPOSE 4001
ENTRYPOINT ["/app/web"]
