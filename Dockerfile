# Stage 1: сборка приложения
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Копируем файлы модуля и устанавливаем зависимости
COPY go.mod ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/web

# Stage 2: создание минимального образа
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/app .

# Открываем порт (при необходимости изменить)
EXPOSE 4000

# Запуск приложения
CMD ["./app"]
