FROM golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags="-s -w" -o isms-server cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs tzdata libreoffice
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
