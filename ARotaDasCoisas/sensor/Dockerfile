FROM golang:1.26-alpine AS sensor

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG SENSOR

RUN CGO_ENABLED=0 GOOS=linux go build -o sensor ./sensor/${SENSOR}

FROM alpine:latest

WORKDIR /app

COPY --from=sensor /app/sensor .

CMD ["./sensor"]