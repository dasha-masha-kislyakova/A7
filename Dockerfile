FROM golang:1.22-alpine AS be-builder
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod ./
RUN go mod download
COPY . .
# один бинарь, который запускает нужный сервис по SERVICE
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/a7 .

# ---- Runtime ----
FROM alpine:3.20
WORKDIR /app
ENV TZ=UTC
COPY --from=be-builder /bin/a7 /app/a7
# Создаем пустую директорию FE вместо копирования из несуществующего fe-builder
RUN mkdir -p /app/FE
EXPOSE 8080 8081 8082 8083 8084
# По умолчанию запускаем прокси; конкретный сервис выбирает docker-compose через SERVICE
ENV SERVICE=proxy PORT=8080
CMD ["/app/a7"]