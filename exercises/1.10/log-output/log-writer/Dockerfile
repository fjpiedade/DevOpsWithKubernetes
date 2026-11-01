
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o log-writer .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/log-writer .

ENV LOG_FILE=/usr/src/app/files/log.txt

CMD ["./log-writer"]
