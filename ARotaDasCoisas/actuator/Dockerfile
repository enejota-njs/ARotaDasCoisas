FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

ARG ACTUATOR

COPY ./actuator/${ACTUATOR} ./actuator/${ACTUATOR}

RUN CGO_ENABLED=0 GOOS=linux go build -o actuator_bin ./actuator/${ACTUATOR}

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/actuator_bin .

CMD ["./actuator_bin"]