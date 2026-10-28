
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o todo-app .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/todo-app .

ENV PORT=9901

EXPOSE 9901

CMD ["./todo-app"]
