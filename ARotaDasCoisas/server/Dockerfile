FROM golang:1.26-alpine AS server

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:latest

WORKDIR /app

COPY --from=server /app/server .

EXPOSE 7000 8000 9000

CMD ["./server"]