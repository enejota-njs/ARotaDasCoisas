FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./server ./server

RUN CGO_ENABLED=0 GOOS=linux go build -o server_bin ./server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server_bin .

EXPOSE 7000 8000 9000

CMD ["./server_bin"]