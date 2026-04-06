FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

ARG SENSOR

COPY ./sensor/${SENSOR} ./sensor/${SENSOR}

RUN CGO_ENABLED=0 GOOS=linux go build -o sensor_bin ./sensor/${SENSOR}

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/sensor_bin .

CMD ["./sensor_bin"]