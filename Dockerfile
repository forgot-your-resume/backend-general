# Используем официальный образ Golang в качестве базового образа
FROM --platform=amd64 golang:1.19.3-buster AS build

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app
ADD . .

RUN GOARCH=amd64 go build .

# final
FROM --platform=amd64 debian:buster-slim

# Копируем файлы go.mod и go.sum в рабочую директорию
COPY go.mod ./

# Загружаем зависимости
# RUN go mod download

# Копируем остальные файлы проекта в рабочую директорию
# COPY . .

# Скачиваем зависимости и строим бинарный файл
# RUN go build -o main .

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates
RUN update-ca-certificates

# Указываем порт, который будет слушать наше приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
