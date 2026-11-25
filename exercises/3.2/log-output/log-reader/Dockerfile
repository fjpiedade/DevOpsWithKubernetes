FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o log-reader .

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/log-reader .
ENV LOG_FILE=/usr/src/app/files/log.txt
ENV COUNTER_FILE=/usr/src/app/files/pingpong.txt
ENV PORT=9901
EXPOSE 9901
CMD ["./log-reader"]
