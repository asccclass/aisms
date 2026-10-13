FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -a -ldflags="-s -w" -o isms-server cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata libreoffice
RUN cp /usr/share/zoneinfo/Asia/Taipei /etc/localtime && echo "Asia/Taipei" > /etc/timezone
WORKDIR /app
COPY --from=builder /src/isms-server /app/isms-server
COPY envfile /app/envfile
COPY www /app/www
COPY assets /app/assets
RUN mkdir -p /app/data /app/logs
VOLUME ["/app/data", "/app/logs"]
EXPOSE 8080
CMD ["/app/isms-server"]
