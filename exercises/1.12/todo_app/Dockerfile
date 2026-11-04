FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o todo .

FROM alpine:3.22

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/todo .

ENV PORT=9901
ENV IMAGE_FILE=/data/image.jpg

EXPOSE 9901

CMD ["./todo"]
