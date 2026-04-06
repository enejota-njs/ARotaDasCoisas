FROM golang:1.26-alpine AS client

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o client ./client

FROM alpine:latest

WORKDIR /app

COPY --from=client /app/client .

CMD ["./client"]