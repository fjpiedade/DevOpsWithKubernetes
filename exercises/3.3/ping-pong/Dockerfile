
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o ping-pong .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/ping-pong .

ENV PORT=9901

EXPOSE 9901

CMD ["./ping-pong"]
