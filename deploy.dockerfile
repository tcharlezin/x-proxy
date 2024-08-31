FROM golang:1.23rc2-alpine3.20 AS builder

RUN mkdir -p /app
WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o x-proxy ./cmd/

FROM alpine:3.20.2

RUN mkdir -p /app
WORKDIR /app
COPY --from=builder /app/x-proxy ./
RUN apk add --no-cache bash

ENTRYPOINT ["./x-proxy"]