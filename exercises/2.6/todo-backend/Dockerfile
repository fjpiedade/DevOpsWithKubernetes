FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o todo-backend .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/todo-backend .

CMD ["./todo-backend"]
