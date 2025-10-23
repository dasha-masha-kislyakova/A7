# build
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/a7 ./

# run
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/a7 /app/a7
COPY migrations /app/migrations
COPY FE /app/FE
ENV PORT=8080 SERVICE=proxy
EXPOSE 8080
CMD ["/app/a7"]
