FROM golang:1.26-alpine AS actuator

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG ACTUATOR

RUN CGO_ENABLED=0 GOOS=linux go build -o actuator ./actuator/${ACTUATOR}

FROM alpine:latest

WORKDIR /app

COPY --from=actuator /app/actuator .

CMD ["./actuator"]