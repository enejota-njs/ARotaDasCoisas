FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./client ./client

RUN CGO_ENABLED=0 GOOS=linux go build -o client_bin ./client

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/client_bin .

CMD ["./client_bin"]