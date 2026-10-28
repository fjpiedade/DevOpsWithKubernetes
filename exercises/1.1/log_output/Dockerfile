
# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o log-output .

# Stage 2: Run the application
FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/log-output .

CMD ["./log-output"]
