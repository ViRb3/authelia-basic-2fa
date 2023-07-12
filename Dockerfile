FROM golang:1.20.6-alpine AS builder

WORKDIR /src
COPY . .

RUN apk add --no-cache git && \
    go mod download && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "authelia-basic-2fa"

FROM alpine:3.18.0

WORKDIR /

COPY --from=builder "/src/authelia-basic-2fa" "/"

ENTRYPOINT ["/authelia-basic-2fa"]
EXPOSE 8081
