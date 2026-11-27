FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o todo-generator .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/todo-generator .

CMD ["./todo-generator"]